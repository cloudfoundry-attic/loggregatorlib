package collectorregistrar

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent"
	"github.com/cloudfoundry/yagnats"
)

type ClientProvider func(*gosteno.Logger, *cfcomponent.Config) (yagnats.NATSConn, error)

type CollectorRegistrar interface {
	Run(stopChan <-chan struct{})
}

type collectorRegistrar struct {
	clientProvider ClientProvider
	interval       time.Duration
	logger         *gosteno.Logger
	cfc            cfcomponent.Component
	client         yagnats.NATSConn
	config         *cfcomponent.Config
}

func NewCollectorRegistrar(clientProvider ClientProvider, cfc cfcomponent.Component, interval time.Duration, config *cfcomponent.Config) CollectorRegistrar {
	return &collectorRegistrar{
		clientProvider: clientProvider,
		logger:         cfc.Logger,
		cfc:            cfc,
		interval:       interval,
		config:         config,
	}
}

func (registrar *collectorRegistrar) Run(stopChan <-chan struct{}) {
	ticker := time.NewTicker(registrar.interval)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			err := registrar.announceMessage()
			if err != nil {
				if registrar.client != nil {
					registrar.client.Close()
					registrar.client = nil
				}
				registrar.logger.Warn(err.Error())
			}
		}
	}
}

func (registrar *collectorRegistrar) announceMessage() error {
	if registrar.client == nil {
		registrar.logger.Debugf("creating NATS client")

		var err error
		registrar.client, err = registrar.clientProvider(registrar.logger, registrar.config)
		if err != nil {
			return fmt.Errorf("Failed to create client: %s", err)
		}
	}

	json, err := json.Marshal(NewAnnounceComponentMessage(registrar.cfc))
	if err != nil {
		return fmt.Errorf("Failed to marshal component message: %s", err)
	}

	err = registrar.client.Publish(AnnounceComponentMessageSubject, json)
	if err != nil {
		return fmt.Errorf("Failed to publish: %s", err)
	}

	return nil
}
