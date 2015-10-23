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

func (c *tlsClient) Close() error {
	var err error
	c.lock.Lock()
	if c.conn != nil {
		err = c.conn.Close()
		c.conn = nil
	}
	c.lock.Unlock()

	return err
}

func (c *tlsClient) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}

	c.lock.Lock()
	if c.conn == nil {
		if err := c.connect(); err != nil {
			return 0, err
		}
	}
	conn := c.conn
	c.lock.Unlock()

	return conn.Write(data)
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
