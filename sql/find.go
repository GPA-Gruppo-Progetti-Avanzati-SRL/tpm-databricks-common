package sql

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-databricks-common/dbricksLks"
	"github.com/databricks/databricks-sdk-go/service/sql"
	"github.com/rs/zerolog/log"
)

type Row map[string]interface{}

func (r Row) ToJson() ([]byte, error) {
	return json.Marshal(r)
}

type ResultSet struct {
	Columns []sql.ColumnInfo
	Rows    []Row
}

func (rs ResultSet) ToJson() ([]byte, error) {
	const semLogContext = "sql-result-set::json"

	var sb bytes.Buffer
	sb.WriteString("[")
	for i, r := range rs.Rows {
		if i > 0 {
			sb.WriteString(", ")
		}

		b, err := json.Marshal(r)
		if err != nil {
			log.Error().Err(err).Msg(semLogContext)
			return nil, err
		}

		sb.Write(b)
	}

	sb.WriteString("]")
	return sb.Bytes(), nil
}

func Find(ctx context.Context, lks *dbricksLks.LinkedService, query string, mustFind bool) (ResultSet, error) {
	const semLogContext = "databricks-linked-service::find"

	query, err := lks.ResolveQuery(query)
	if err != nil {
		log.Error().Err(err).Msg(semLogContext)
		return ResultSet{}, err
	}

	resp, err := lks.W.StatementExecution.ExecuteAndWait(ctx, sql.ExecuteStatementRequest{
		WarehouseId: lks.Cfg.WarehouseID,
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
		chunk, err := lks.W.StatementExecution.GetStatementResultChunkN(ctx, sql.GetStatementResultChunkNRequest{
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

func parseRows(cols []sql.ColumnInfo, rows [][]string) ([]Row, error) {
	const semLogContext = "databricks-linked-service::parse-rows"

	var resultSet []Row
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
func parseRow(cols []sql.ColumnInfo, row []string) (Row, error) {
	const semLogContext = "databricks-linked-service::parse-row"

	rec := make(Row, len(cols))
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
