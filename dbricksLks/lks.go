package dbricksLks

import (
	"errors"
	"fmt"

	"github.com/databricks/databricks-sdk-go"
	"github.com/rs/zerolog/log"
)

type LinkedService struct {
	cfg Config
	w   *databricks.WorkspaceClient
}

func NewLinkedServiceWithConfig(cfg Config) (*LinkedService, error) {
	const semLogContext = "databricks-linked-service::new"
	var err error
	var w *databricks.WorkspaceClient

	if cfg.WarehouseID == "" {
		err = errors.New("databricks wharehouse-id param is required")
		return nil, err
	}

	switch cfg.AuthType {
	case AuthTypeAzureClientSecret:
		w, err = databricks.NewWorkspaceClient(&databricks.Config{
			Host:              cfg.Host,
			AzureClientID:     cfg.ServicePrincipal.ClientID,
			AzureClientSecret: cfg.ServicePrincipal.ClientSecret,
			AzureTenantID:     cfg.ServicePrincipal.TenantID,
			AzureResourceID:   cfg.WorkspaceResourceID,
		})
	default:
		err = fmt.Errorf("databricks linked service auth type (%s) not supported", cfg.AuthType)
	}

	if err != nil {
		log.Error().Err(err).Msg(semLogContext)
		return nil, err
	}

	lks := &LinkedService{cfg: cfg, w: w}
	return lks, nil
}

func NewLinkedService(name string, opts ...Option) (*LinkedService, error) {
	cfg := Config{Name: name}

	for _, o := range opts {
		o(&cfg)
	}

	return NewLinkedServiceWithConfig(cfg)
}
