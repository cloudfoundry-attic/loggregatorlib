package loggregatorclient

import (
	"net"

	"github.com/cloudfoundry/gosteno"
)

const DefaultBufferSize = 4096

type Client interface {
	Scheme() string
	Address() string
	Send([]byte)
	Stop()
}

type udpClient struct {
	addr   *net.UDPAddr
	conn   net.PacketConn
	logger *gosteno.Logger
}

func NewUDPClient(logger *gosteno.Logger, address string, bufferSize int) (Client, error) {
	la, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	connection, err := net.ListenPacket("udp", "")
	if err != nil {
		return nil, err
	}

	loggregatorClient := &udpClient{
		addr:   la,
		conn:   connection,
		logger: logger,
	}
	return loggregatorClient, nil
}

func (c *udpClient) Scheme() string {
	return "udp"
}

func (c *udpClient) Address() string {
	return c.addr.String()
}

func (c *udpClient) Stop() {
	c.conn.Close()
}

func (c *udpClient) Send(data []byte) {
	if len(data) == 0 {
		c.logger.Debugf("Skipped writing of 0 byte message to %s", c.Address())
		return
	}

	writeCount, err := c.conn.WriteTo(data, c.addr)
	if err != nil {
		c.logger.Errorf("Writing to loggregator %s failed %s", c.Address(), err)
		return
	}
	c.logger.Debugf("Wrote %d bytes to %s", writeCount, c.Address())
}
