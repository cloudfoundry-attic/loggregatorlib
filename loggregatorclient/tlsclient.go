package loggregatorclient

import "github.com/cloudfoundry/gosteno"

type tlsClient struct {
	address string
	logger  *gosteno.Logger
}

func NewTLSClient(logger *gosteno.Logger, address string) (Client, error) {
	loggregatorClient := &tlsClient{
		address: address,
		logger:  logger,
	}
	return loggregatorClient, nil
}

func (c *tlsClient) Scheme() string {
	return "tls"
}

func (c *tlsClient) Address() string {
	return c.address
}

func (c *tlsClient) Stop() {
}

func (c *tlsClient) Send(data []byte) {
	c.logger.Warn("Sending over TLS (unsupported)")
}
