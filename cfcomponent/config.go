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

var DefaultYagnatsClientProvider = func(logger *gosteno.Logger) yagnats.NATSClient {
	client := yagnats.NewClient()
	client.SetLogger(logger)
	return client
}

func (c *Config) Validate(logger *gosteno.Logger) (err error) {
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

	natsClient := yagnats.NewClient()

	err = natsClient.Connect(connectionInfo)

	if err != nil {
		return errors.New(fmt.Sprintf("Could not connect to NATS: %v", err.Error()))
	}

	c.MbusClient = natsClient
	return
}
