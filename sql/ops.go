package sql

import (
	"context"
	"errors"

	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-databricks-common/dbricksLks"
	"github.com/rs/zerolog/log"
)

type DatabricksSQLOperationType string

const (
	FindOperation    DatabricksSQLOperationType = "sql-find"
	FindOneOperation DatabricksSQLOperationType = "sql-find-one"
)

func ExecuteOperation(ctx context.Context, lks *dbricksLks.LinkedService, opType DatabricksSQLOperationType, stmt string, mustFind bool) ([]byte, error) {
	const semLogContext = "databricks::sql-operation"
	var err error
	var b []byte

	switch opType {
	case FindOperation:
		b, err = JsonFind(ctx, lks, stmt)
	case FindOneOperation:
		b, err = JsonFindOne(ctx, lks, stmt, mustFind)
	default:
		err = errors.New("invalid op type: " + string(opType))
	}

	if err != nil {
		log.Error().Err(err).Msg(semLogContext)
	}

	return b, err
}

func JsonFind(ctx context.Context, lks *dbricksLks.LinkedService, query string) ([]byte, error) {
	rs, err := Find(ctx, lks, query, false)
	if err != nil {
		return nil, err
	}

	return rs.ToJson()
}

func JsonFindOne(ctx context.Context, lks *dbricksLks.LinkedService, query string, mustFind bool) ([]byte, error) {
	rs, err := Find(ctx, lks, query, mustFind)
	if err != nil {
		return nil, err
	}

	if len(rs.Rows) == 0 {
		return []byte(`{}`), nil
	}

	return rs.Rows[0].ToJson()
}
