package cfcomponent

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/yagnats"
)

type Config struct {
	Syslog     string
	VarzPort   uint32
	VarzUser   string
	VarzPass   string
	NatsHosts  []string
	NatsPort   int
	NatsUser   string
	NatsPass   string
	MbusClient yagnats.NATSClient
}

var DefaultYagnatsClientProvider = func(logger *gosteno.Logger, c *Config) (natsClient yagnats.NATSClient, err error) {
	members := []yagnats.ConnectionProvider{}
	for _, natsHost := range c.NatsHosts {
		members = append(members, &yagnats.ConnectionInfo{
			Addr:     fmt.Sprintf("%s:%d", natsHost, c.NatsPort),
			Username: c.NatsUser,
			Password: c.NatsPass,
		})
	}

	connectionInfo := &yagnats.ConnectionCluster{
		Members: members,
	}
	natsClient = yagnats.NewClient()
	err = natsClient.Connect(connectionInfo)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Could not connect to NATS: %v", err.Error()))
	}
	return natsClient, nil
}

func (c *Config) Validate(logger *gosteno.Logger) (err error) {
	c.MbusClient, err = DefaultYagnatsClientProvider(logger, c)
	return
}
