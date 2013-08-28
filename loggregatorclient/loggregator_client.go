package loggregatorclient

import (
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent/instrumentation"
	"net"
	"sync/atomic"
)

const DefaultBufferSize = 4096

type LoggregatorClient interface {
	instrumentation.Instrumentable
	Send([]byte)
	IncLogStreamRawByteCount(uint64)
	IncLogStreamPbByteCount(uint64)
}

type udpLoggregatorClient struct {
	receivedMessageCount  *uint64
	sentMessageCount      *uint64
	receivedByteCount     *uint64
	sentByteCount         *uint64
	logStreamRawByteCount *uint64
	logStreamPbByteCount  *uint64
	sendChannel           chan []byte
}

func NewLoggregatorClient(loggregatorAddress string, logger *gosteno.Logger, bufferSize int) LoggregatorClient {
	loggregatorClient := &udpLoggregatorClient{receivedMessageCount: new(uint64), sentMessageCount: new(uint64),
		receivedByteCount: new(uint64), sentByteCount: new(uint64), logStreamRawByteCount: new(uint64),
		logStreamPbByteCount: new(uint64)}

	la, err := net.ResolveUDPAddr("udp", loggregatorAddress)
	if err != nil {
		logger.Fatalf("Error resolving loggregator address %s, %s", loggregatorAddress, err)
	}

	connection, err := net.ListenPacket("udp", "")
	if err != nil {
		logger.Fatalf("Error opening udp stuff")
	}

	loggregatorClient.sendChannel = make(chan []byte, bufferSize)

	go func() {
		for {
			dataToSend := <-loggregatorClient.sendChannel
			if len(dataToSend) > 0 {
				writeCount, err := connection.WriteTo(dataToSend, la)
				if err != nil {
					logger.Errorf("Writing to loggregator %s failed %s", loggregatorAddress, err)
					continue
				}
				logger.Debugf("Wrote %d bytes to %s", writeCount, loggregatorAddress)
				atomic.AddUint64(loggregatorClient.sentMessageCount, 1)
				atomic.AddUint64(loggregatorClient.sentByteCount, uint64(writeCount))
			} else {
				logger.Debugf("Skipped writing of 0 byte message to %s", loggregatorAddress)
			}
		}
	}()

	return loggregatorClient
}

func (loggregatorClient *udpLoggregatorClient) Send(data []byte) {
	atomic.AddUint64(loggregatorClient.receivedMessageCount, 1)
	atomic.AddUint64(loggregatorClient.receivedByteCount, uint64(len(data)))
	loggregatorClient.sendChannel <- data
}

func (loggregatorClient *udpLoggregatorClient) Emit() instrumentation.Context {
	return instrumentation.Context{Name: "loggregatorClient",
		Metrics: []instrumentation.Metric{
			instrumentation.Metric{Name: "currentBufferCount", Value: uint64(len(loggregatorClient.sendChannel))},
			instrumentation.Metric{Name: "sentMessageCount", Value: atomic.LoadUint64(loggregatorClient.sentMessageCount)},
			instrumentation.Metric{Name: "receivedMessageCount", Value: atomic.LoadUint64(loggregatorClient.receivedMessageCount)},
			instrumentation.Metric{Name: "sentByteCount", Value: atomic.LoadUint64(loggregatorClient.sentByteCount)},
			instrumentation.Metric{Name: "receivedByteCount", Value: atomic.LoadUint64(loggregatorClient.receivedByteCount)},
			instrumentation.Metric{Name: "logStreamRawByteCount", Value: atomic.LoadUint64(loggregatorClient.logStreamRawByteCount)},
			instrumentation.Metric{Name: "logStreamPbByteCount", Value: atomic.LoadUint64(loggregatorClient.logStreamPbByteCount)},
		},
	}
}

func (loggregatorClient *udpLoggregatorClient) IncLogStreamRawByteCount(count uint64) {
	atomic.AddUint64(loggregatorClient.logStreamRawByteCount, count)
}

func (loggregatorClient *udpLoggregatorClient) IncLogStreamPbByteCount(count uint64) {
	atomic.AddUint64(loggregatorClient.logStreamPbByteCount, count)
}
