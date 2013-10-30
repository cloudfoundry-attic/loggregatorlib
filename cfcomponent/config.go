package cfcomponent

import (
	"errors"
	"fmt"
	cfmessagebus "github.com/cloudfoundry/go_cfmessagebus"
	"github.com/cloudfoundry/gosteno"
)

type Config struct {
	Syslog     string
	VarzPort   uint32
	VarzUser   string
	VarzPass   string
	NatsHost   string
	NatsPort   int
	NatsUser   string
	NatsPass   string
	MbusClient cfmessagebus.MessageBus
}

func (c *Config) Validate(logger *gosteno.Logger) (err error) {
	c.MbusClient, err = cfmessagebus.NewMessageBus("NATS")
	if err != nil {
		return errors.New(fmt.Sprintf("Can not create message bus to NATS: %s", err))
	}
	c.MbusClient.Configure(c.NatsHost, c.NatsPort, c.NatsUser, c.NatsPass)
	c.MbusClient.SetLogger(logger)
	err = c.MbusClient.Connect()
	if err != nil {
		return errors.New(fmt.Sprintf("Could not connect to NATS: %v", err.Error()))
	}
	return
}
