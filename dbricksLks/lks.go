package dbricksLks

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/databricks/databricks-sdk-go"
	"github.com/rs/zerolog/log"
)

type LinkedService struct {
	Cfg Config
	W   *databricks.WorkspaceClient
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
			AuthType:          string(cfg.AuthType),
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

	lks := &LinkedService{Cfg: cfg, W: w}
	return lks, nil
}

func NewLinkedService(name string, opts ...Option) (*LinkedService, error) {
	cfg := Config{Name: name}

	for _, o := range opts {
		o(&cfg)
	}

	return NewLinkedServiceWithConfig(cfg)
}

var resourcePattern = regexp.MustCompile(`\[dbrks:([^\]]+)\]`)

func (lks *LinkedService) ResolveQuery(query string) (string, error) {
	const semLogContext = "databricks-linked-service::resolve-query"

	var resolveErr error
	resolved := resourcePattern.ReplaceAllStringFunc(query, func(match string) string {
		if resolveErr != nil {
			return match
		}
		resourceId := resourcePattern.FindStringSubmatch(match)[1]
		for _, r := range lks.Cfg.Resources {
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
