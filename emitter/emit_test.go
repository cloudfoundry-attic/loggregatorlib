package emitter

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent/instrumentation"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockLoggregatorClient struct {
	received chan *[]byte
}

func (m MockLoggregatorClient) Send(data []byte) {
	go func() {
		m.received <- &data
	}()
}

func (m MockLoggregatorClient) Emit() instrumentation.Context {
	return instrumentation.Context{}
}

func (m MockLoggregatorClient) IncLogStreamRawByteCount(uint64) {

}

func (m MockLoggregatorClient) IncLogStreamPbByteCount(uint64) {

}

func TestEmitter(t *testing.T) {
	received := make(chan *[]byte)
	e, _ := NewEmitter("localhost:3456", "ROUTER", nil)
	e.lc = &MockLoggregatorClient{received}
	e.Emit("appid", "foo")
	receivedMessage := getBackendMessage(t, <-received)
	assert.Equal(t, receivedMessage.GetMessage(), []byte("foo"))
	assert.Equal(t, receivedMessage.GetAppId(), "appid")
}

func TestInvalidSourcetype(t *testing.T) {
	_, err := NewEmitter("server", "FOOSERVER", nil)
	assert.Error(t, err)
}

func TestValidSourcetype(t *testing.T) {
	_, err := NewEmitter("localhost:38452", "ROUTER", nil)
	assert.NoError(t, err)
}

func getBackendMessage(t *testing.T, data *[]byte) *logmessage.LogMessage {
	receivedMessage := &logmessage.LogMessage{}

	err := proto.Unmarshal(*data, receivedMessage)

	if err != nil {
		t.Fatalf("Message invalid. %s", err)
	}
	return receivedMessage
}
