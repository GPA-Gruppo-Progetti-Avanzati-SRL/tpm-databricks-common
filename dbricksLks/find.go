package dbricksLks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/databricks/databricks-sdk-go/service/sql"
	"github.com/rs/zerolog/log"
)

type ResultSet struct {
	Columns []sql.ColumnInfo
	Rows    []map[string]interface{}
}

func (lks *LinkedService) Find(ctx context.Context, query string, mustFind bool) (ResultSet, error) {
	const semLogContext = "databricks-linked-service::find"

	query, err := lks.resolveQuery(query)
	if err != nil {
		log.Error().Err(err).Msg(semLogContext)
		return ResultSet{}, err
	}

	resp, err := lks.w.StatementExecution.ExecuteAndWait(ctx, sql.ExecuteStatementRequest{
		WarehouseId: lks.cfg.WarehouseID,
		Statement:   query,
	})

	var rs ResultSet
	if err != nil {
		log.Error().Err(err).Msg(semLogContext)
		return rs, err
	}

	if resp.Result == nil || resp.Manifest == nil {
		if mustFind {
			err = errors.New("empty result")
			log.Error().Err(err).Msg(semLogContext)
			return ResultSet{}, err
		}

		return rs, nil
	}

	rs.Columns = resp.Manifest.Schema.Columns
	totalChunks := resp.Manifest.TotalChunkCount
	log.Info().Int("total-chunks", totalChunks).Int("num-columns", len(rs.Columns)).Msg(semLogContext)

	rows, err := parseRows(rs.Columns, resp.Result.DataArray)
	if err != nil {
		log.Error().Err(err).Msg(semLogContext)
		return rs, err
	}
	rs.Rows = append(rs.Rows, rows...)

	log.Info().Str("chunk", fmt.Sprintf("%d/%d", 1, totalChunks)).Int("rows", len(rows)).Int("total-rows", len(rs.Rows)).Msg(semLogContext)

	for chunkIdx := 1; chunkIdx < totalChunks; chunkIdx++ {
		chunk, err := lks.w.StatementExecution.GetStatementResultChunkN(ctx, sql.GetStatementResultChunkNRequest{
			StatementId: resp.StatementId,
			ChunkIndex:  chunkIdx,
		})
		if err != nil {
			err = fmt.Errorf("getting chunk %d: %w", chunkIdx, err)
			log.Error().Err(err).Msg(semLogContext)
			return rs, err
		}

		rows, err = parseRows(rs.Columns, chunk.DataArray)
		if err != nil {
			log.Error().Err(err).Msg(semLogContext)
			return rs, err
		}

		rs.Rows = append(rs.Rows, rows...)
		log.Info().Str("chunk", fmt.Sprintf("%d/%d", chunkIdx+1, totalChunks)).Int("rows", len(rows)).Int("total-rows", len(rs.Rows)).Msg(semLogContext)
	}

	return rs, nil
}

var resourcePattern = regexp.MustCompile(`\[dbrks:([^\]]+)\]`)

func (lks *LinkedService) resolveQuery(query string) (string, error) {
	const semLogContext = "databricks-linked-service::resolve-query"

	var resolveErr error
	resolved := resourcePattern.ReplaceAllStringFunc(query, func(match string) string {
		if resolveErr != nil {
			return match
		}
		resourceId := resourcePattern.FindStringSubmatch(match)[1]
		for _, r := range lks.cfg.Resources {
			if r.Id == resourceId {
				return r.Name
			}
		}
		resolveErr = fmt.Errorf("resource %q not found in config", resourceId)
		log.Error().Err(resolveErr).Msg(semLogContext)
		return match
	})

	if resolveErr != nil {
		return "", resolveErr
	}

	return resolved, nil
}

func parseRows(cols []sql.ColumnInfo, rows [][]string) ([]map[string]interface{}, error) {
	const semLogContext = "databricks-linked-service::parse-rows"

	var resultSet []map[string]interface{}
	for _, row := range rows {
		parsedRow, err := parseRow(cols, row)
		if err != nil {
			log.Error().Err(err).Msg(semLogContext)
			return nil, err
		}

		resultSet = append(resultSet, parsedRow)
	}

	return resultSet, nil
}

// parseRow converts a raw Databricks DataArray row into a typed map.
// Numeric and boolean columns are cast to their native Go types; ARRAY, MAP and
// STRUCT columns are JSON-deserialized into interface{}. Everything else is
// left as a string. An empty string is treated as SQL NULL and mapped to nil
// for all non-string types.
func parseRow(cols []sql.ColumnInfo, row []string) (map[string]interface{}, error) {
	const semLogContext = "databricks-linked-service::parse-row"

	rec := make(map[string]interface{}, len(cols))
	for i, col := range cols {
		if i >= len(row) {
			break
		}
		v, err := parseValue(col.TypeName, row[i])
		if err != nil {
			err = fmt.Errorf("parsing column %s: %w", col.Name, err)
			log.Error().Err(err).Msg(semLogContext)
			return nil, err
		}
		rec[col.Name] = v
	}

	return rec, nil
}

func parseValue(typeName sql.ColumnInfoTypeName, s string) (interface{}, error) {
	const semLogContext = "databricks-linked-service::parse-value"
	var err error
	var v interface{}
	switch typeName {
	case sql.ColumnInfoTypeNameNull:
		return nil, nil

	case sql.ColumnInfoTypeNameByte, sql.ColumnInfoTypeNameShort,
		sql.ColumnInfoTypeNameInt, sql.ColumnInfoTypeNameLong:
		if s == "" {
			return nil, nil
		}
		v, err = strconv.ParseInt(s, 10, 64)

	case sql.ColumnInfoTypeNameFloat, sql.ColumnInfoTypeNameDouble,
		sql.ColumnInfoTypeNameDecimal:
		if s == "" {
			return nil, nil
		}
		v, err = strconv.ParseFloat(s, 64)

	case sql.ColumnInfoTypeNameBoolean:
		if s == "" {
			return nil, nil
		}
		v, err = strconv.ParseBool(s)

	case sql.ColumnInfoTypeNameArray, sql.ColumnInfoTypeNameMap,
		sql.ColumnInfoTypeNameStruct:
		if s == "" {
			return nil, nil
		}
		err = json.Unmarshal([]byte(s), &v)

	default:
		// STRING, CHAR, DATE, TIMESTAMP, BINARY, INTERVAL, USER_DEFINED_TYPE
		v = s
	}

	if err != nil {
		log.Error().Err(err).Msg(semLogContext)
		return nil, err
	}

	return v, nil
}
