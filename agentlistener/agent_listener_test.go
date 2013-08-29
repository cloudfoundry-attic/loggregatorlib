package agentlistener

import (
	"github.com/cloudfoundry/gosteno"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestThatItListens(t *testing.T) {

	listener := NewAgentListener("127.0.0.1:3456", gosteno.NewLogger("TestLogger"))
	dataChannel := listener.Start()

	expectedData := "Some Data"
	otherData := "More stuff"

	connection, err := net.Dial("udp", "localhost:3456")

	_, err = connection.Write([]byte(expectedData))
	assert.NoError(t, err)

	_, err = connection.Write([]byte(otherData))
	assert.NoError(t, err)

	received := <-dataChannel
	assert.Equal(t, expectedData, string(received))

	receivedAgain := <-dataChannel
	assert.Equal(t, otherData, string(receivedAgain))

	metrics := listener.Emit().Metrics
	assert.Equal(t, len(metrics), 3) //make sure all expected metrics are present
	for _, metric := range metrics {
		switch metric.Name {
		case "currentBufferCount":
			assert.Equal(t, metric.Value, 0)
		case "receivedMessageCount":
			assert.Equal(t, metric.Value, uint64(2))
		case "receivedByteCount":
			assert.Equal(t, metric.Value, uint64(19))
		default:
			t.Error("Got an invalid metric name: ", metric.Name)
		}
	}
}
