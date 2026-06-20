package dbricksLks

import (
	"errors"

	"github.com/rs/zerolog/log"
)

type LinkedServices []*LinkedService

var theRegistry LinkedServices

func Initialize(cfgs []Config) (LinkedServices, error) {

	const semLogContext = "databricks-registry::initialize"
	if len(cfgs) == 0 {
		log.Info().Msg(semLogContext + " no config provided....skipping")
		return nil, nil
	}

	if len(theRegistry) != 0 {
		log.Warn().Msg(semLogContext + " registry already configured.. overwriting")
	}

	log.Info().Int("no-linked-services", len(cfgs)).Msg(semLogContext)

	var r LinkedServices
	for _, kcfg := range cfgs {
		lks, err := NewLinkedServiceWithConfig(kcfg)
		if err != nil {
			return nil, err
		}

		r = append(r, lks)
		log.Info().Str("aws-name", kcfg.Name).Msg(semLogContext + " databricks instance configured")
	}

	theRegistry = r
	return r, nil
}

func GetLinkedService(name string) (*LinkedService, error) {
	const semLogContext = "databricks-registry::get-lks"
	for _, stg := range theRegistry {
		if stg.cfg.Name == name {
			return stg, nil
		}
	}

	err := errors.New("databricks linked service not found by name " + name)
	log.Error().Err(err).Str("stg-name", name).Msg(semLogContext)
	return nil, err
}
