package loggregatorclient

import (
	"net"

	"github.com/cloudfoundry/gosteno"
)

const DefaultBufferSize = 4096

type LoggregatorClient interface {
	Send([]byte)
	Stop()
}

type udpLoggregatorClient struct {
	sendChannel        chan []byte
	loggregatorAddress string
	doneChannel        chan struct{}
}

func NewLoggregatorClient(loggregatorAddress string, logger *gosteno.Logger, bufferSize int) LoggregatorClient {
	loggregatorClient := &udpLoggregatorClient{}

	la, err := net.ResolveUDPAddr("udp", loggregatorAddress)
	if err != nil {
		logger.Fatalf("Error resolving loggregator address %s, %s", loggregatorAddress, err)
	}

	connection, err := net.ListenPacket("udp", "")
	if err != nil {
		logger.Fatalf("Error opening udp stuff")
	}

	loggregatorClient.loggregatorAddress = la.IP.String()
	loggregatorClient.sendChannel = make(chan []byte, bufferSize)
	loggregatorClient.doneChannel = make(chan struct{})

	go func() {
		for dataToSend := range loggregatorClient.sendChannel {
			if len(dataToSend) == 0 {
				logger.Debugf("Skipped writing of 0 byte message to %s", loggregatorAddress)
				continue
			}

			writeCount, err := connection.WriteTo(dataToSend, la)
			if err != nil {
				logger.Errorf("Writing to loggregator %s failed %s", loggregatorAddress, err)
				continue
			}
			logger.Debugf("Wrote %d bytes to %s", writeCount, loggregatorAddress)
		}

		close(loggregatorClient.doneChannel)
	}()

	return loggregatorClient
}

func (loggregatorClient *udpLoggregatorClient) Stop() {
	close(loggregatorClient.sendChannel)

	<-loggregatorClient.doneChannel
}

func (loggregatorClient *udpLoggregatorClient) Send(data []byte) {
	loggregatorClient.sendChannel <- data
}
