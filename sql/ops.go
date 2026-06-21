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

type Operation struct {
	Text     string
	MustFind bool
	Type     DatabricksSQLOperationType
}

func NewOperation(operationType DatabricksSQLOperationType, text string, mustFind bool) (*Operation, error) {
	return &Operation{
		Text:     text,
		MustFind: mustFind,
		Type:     operationType,
	}, nil
}

func (op *Operation) Execute(ctx context.Context, lks *dbricksLks.LinkedService) ([]byte, error) {
	const semLogContext = "databricks::sql-operation"
	var err error
	var b []byte

	switch op.Type {
	case FindOperation:
		b, err = JsonFind(ctx, lks, op.Text)
	case FindOneOperation:
		b, err = JsonFindOne(ctx, lks, op.Text, op.MustFind)
	default:
		err = errors.New("invalid op type: " + string(op.Type))
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
