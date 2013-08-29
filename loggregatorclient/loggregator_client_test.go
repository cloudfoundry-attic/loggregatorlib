package loggregatorclient

import (
	"github.com/cloudfoundry/gosteno"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestSend(t *testing.T) {
	bufferSize := 4096
	expectedOutput := []byte("Important Testmessage")
	loggregatorClient := NewLoggregatorClient("127.0.0.1:9876", gosteno.NewLogger("TestLogger"), bufferSize)

	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:9876")
	assert.NoError(t, err)

	udpListener, err := net.ListenUDP("udp", udpAddr)
	defer udpListener.Close()
	assert.NoError(t, err)

	loggregatorClient.IncLogStreamRawByteCount(uint64(len(expectedOutput)))
	loggregatorClient.IncLogStreamPbByteCount(uint64(len(expectedOutput)))
	loggregatorClient.Send(expectedOutput)

	buffer := make([]byte, bufferSize)
	readCount, _, err := udpListener.ReadFromUDP(buffer)
	assert.NoError(t, err)

	received := string(buffer[:readCount])
	assert.Equal(t, string(expectedOutput), received)

	metrics := loggregatorClient.Emit().Metrics
	assert.Equal(t, len(metrics), 7) //make sure all expected metrics are present
	for _, metric := range metrics {
		switch metric.Name {
		case "127:0:0:1.currentBufferCount":
			assert.Equal(t, metric.Value, uint64(0))
		case "127:0:0:1.sentMessageCount":
			assert.Equal(t, metric.Value, uint64(1))
		case "127:0:0:1.sentByteCount":
			assert.Equal(t, metric.Value, uint64(21))
		case "127:0:0:1.receivedMessageCount":
			assert.Equal(t, metric.Value, uint64(1))
		case "127:0:0:1.receivedByteCount":
			assert.Equal(t, metric.Value, uint64(21))
		case "127:0:0:1.logStreamRawByteCount":
			assert.Equal(t, metric.Value, uint64(len(expectedOutput)))
		case "127:0:0:1.logStreamPbByteCount":
			assert.Equal(t, metric.Value, uint64(len(expectedOutput)))
		default:
			t.Error("Got an invalid metric name: ", metric.Name)
		}
	}
}

func TestDontSendEmptyData(t *testing.T) {
	bufferSize := 4096
	firstMessage := []byte("")
	secondMessage := []byte("hi")
	loggregatorClient := NewLoggregatorClient("localhost:9876", gosteno.NewLogger("TestLogger"), bufferSize)

	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:9876")
	assert.NoError(t, err)

	udpListener, err := net.ListenUDP("udp", udpAddr)
	defer udpListener.Close()
	assert.NoError(t, err)

	loggregatorClient.Send(firstMessage)
	loggregatorClient.Send(secondMessage)

	buffer := make([]byte, bufferSize)
	readCount, _, err := udpListener.ReadFromUDP(buffer)

	assert.NoError(t, err)

	received := string(buffer[:readCount])
	assert.Equal(t, string(secondMessage), received)
}
