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
	address     string
	sendChannel chan []byte
	doneChannel chan struct{}
}

func NewUDPClient(logger *gosteno.Logger, address string, bufferSize int) (Client, error) {
	loggregatorClient := &udpClient{
		address: address,
	}

	la, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	connection, err := net.ListenPacket("udp", "")
	if err != nil {
		return nil, err
	}

	loggregatorClient.sendChannel = make(chan []byte, bufferSize)
	loggregatorClient.doneChannel = make(chan struct{})

	go func() {
		for dataToSend := range loggregatorClient.sendChannel {
			if len(dataToSend) == 0 {
				logger.Debugf("Skipped writing of 0 byte message to %s", address)
				continue
			}

			writeCount, err := connection.WriteTo(dataToSend, la)
			if err != nil {
				logger.Errorf("Writing to loggregator %s failed %s", address, err)
				continue
			}
			logger.Debugf("Wrote %d bytes to %s", writeCount, address)
		}

		close(loggregatorClient.doneChannel)
	}()

	return loggregatorClient, nil
}

func (udpclient *udpClient) Scheme() string {
	return "udp"
}

func (udpclient *udpClient) Address() string {
	return udpclient.address
}

func (udpclient *udpClient) Stop() {
	close(udpclient.sendChannel)

	<-udpclient.doneChannel
}

func (udpclient *udpClient) Send(data []byte) {
	udpclient.sendChannel <- data
}
