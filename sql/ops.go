package sql

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/tpm-databricks-common/dbricksLks"
	"github.com/rs/zerolog/log"
)

type DatabricksSQLOperationType string

const (
	FindOperation    DatabricksSQLOperationType = "sql-find"
	FindOneOperation DatabricksSQLOperationType = "sql-find-one"
)

type Operation struct {
	Text     string                     `json:"text,omitempty" yaml:"text,omitempty"`
	MustFind bool                       `json:"must-find,omitempty" yaml:"must-find,omitempty"`
	Type     DatabricksSQLOperationType `json:"type,omitempty" yaml:"type,omitempty"`
}

func (op *Operation) ToJsonString() string {
	b, err := json.Marshal(op)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to marshal operation")
	}

	return string(b)
}

type OperationResult struct {
	StatusCode   int `json:"status-code,omitempty" yaml:"status-code,omitempty"`
	MatchedCOunt int `json:"matched-count,omitempty" yaml:"matched-count,omitempty"`
}

func NewOperation(operationType DatabricksSQLOperationType, text string, mustFind bool) (*Operation, error) {
	return &Operation{
		Text:     text,
		MustFind: mustFind,
		Type:     operationType,
	}, nil
}

func (op *Operation) Execute(ctx context.Context, lks *dbricksLks.LinkedService) (OperationResult, []byte, error) {
	const semLogContext = "databricks::sql-operation"
	var err error
	var b []byte
	opResult := OperationResult{StatusCode: http.StatusInternalServerError}

	switch op.Type {
	case FindOperation:
		opResult, b, err = JsonFind(ctx, lks, op.Text)
	case FindOneOperation:
		opResult, b, err = JsonFindOne(ctx, lks, op.Text)
	default:
		err = errors.New("invalid op type: " + string(op.Type))
	}

	if err != nil {
		log.Error().Err(err).Msg(semLogContext)
	}

	return opResult, b, err
}

func JsonFind(ctx context.Context, lks *dbricksLks.LinkedService, query string) (OperationResult, []byte, error) {
	rs, err := Find(ctx, lks, query, false)
	if err != nil {
		return OperationResult{StatusCode: http.StatusInternalServerError}, nil, err
	}

	j, err := rs.ToJson()
	if err != nil {
		return OperationResult{StatusCode: http.StatusInternalServerError}, nil, err
	}

	return OperationResult{StatusCode: http.StatusOK}, j, nil
}

func JsonFindOne(ctx context.Context, lks *dbricksLks.LinkedService, query string) (OperationResult, []byte, error) {
	rs, err := Find(ctx, lks, query, false)
	if err != nil {
		return OperationResult{StatusCode: http.StatusInternalServerError}, nil, err
	}

	if len(rs.Rows) == 0 {
		return OperationResult{StatusCode: http.StatusNotFound}, []byte(`{}`), errors.New("no rows found")
	}

	j, err := rs.Rows[0].ToJson()
	if err != nil {
		return OperationResult{StatusCode: http.StatusInternalServerError}, nil, err
	}

	return OperationResult{StatusCode: http.StatusOK}, j, nil
}
