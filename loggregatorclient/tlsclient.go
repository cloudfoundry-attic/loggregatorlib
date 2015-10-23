package loggregatorclient

import (
	"net"
	"sync"
	"time"

	"github.com/cloudfoundry/gosteno"
)

const timeout = 1 * time.Second

type tlsClient struct {
	address string
	logger  *gosteno.Logger

	lock sync.Mutex
	conn net.Conn
}

func NewTLSClient(logger *gosteno.Logger, address string) (Client, error) {
	c := &tlsClient{
		address: address,
		logger:  logger,
	}

	_ = c.connect()
	return c, nil
}

func (c *tlsClient) Scheme() string {
	return "tls"
}

func (c *tlsClient) Address() string {
	return c.address
}

func (c *tlsClient) Stop() {
	c.lock.Lock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.lock.Unlock()
}

func (c *tlsClient) Send(data []byte) {
	c.lock.Lock()
	if c.conn == nil {
		if c.connect() != nil {
			return
		}
	}
	conn := c.conn
	c.lock.Unlock()

	_, err := conn.Write(data)
	if err != nil {
		c.Stop()
	}
}

func (c *tlsClient) connect() error {
	var err error
	c.conn, err = net.DialTimeout("tcp", c.address, timeout)
	if err != nil {
		c.logger.Warnd(map[string]interface{}{
			"error":   err,
			"address": c.address,
		}, "Failed to connect over TLS")
	}
	return err
}
