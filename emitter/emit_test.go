package emitter

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent/instrumentation"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/cloudfoundry/loggregatorlib/logmessage/testhelpers"
	"github.com/cloudfoundry/loggregatorlib/signature"
	"github.com/stretchr/testify/assert"
	"testing"
)

type MockLoggregatorClient struct {
	received chan *[]byte
}

func (m MockLoggregatorClient) Send(data []byte) {
	m.received <- &data
}

func (m MockLoggregatorClient) Emit() instrumentation.Context {
	return instrumentation.Context{}
}

func TestEmit(t *testing.T) {
	received := make(chan *[]byte, 1)
	e, _ := NewLogMessageEmitter("localhost:3456", "ROUTER", "42", nil)
	e.LoggregatorClient = &MockLoggregatorClient{received}
	e.Emit("appid", "foo")
	receivedMessage := extractLogMessage(t, <-received)

	assert.Equal(t, receivedMessage.GetMessage(), []byte("foo"))
	assert.Equal(t, receivedMessage.GetAppId(), "appid")
	assert.Equal(t, receivedMessage.GetSourceId(), "42")
}


func TestEmitWithNewEmitter(t *testing.T) {
	received := make(chan *[]byte, 1)
	e, _ := NewEmitter("localhost:3456", "ROUTER", "42", nil)
	e.LoggregatorClient = &MockLoggregatorClient{received}
	e.Emit("appid", "foo")
	receivedMessage := extractLogMessage(t, <-received)

	assert.Equal(t, receivedMessage.GetMessage(), []byte("foo"))
	assert.Equal(t, receivedMessage.GetAppId(), "appid")
	assert.Equal(t, receivedMessage.GetSourceId(), "42")
}

func TestLogMessageEmit(t *testing.T) {
	received := make(chan *[]byte, 1)
	e, _ := NewLogMessageEmitter("localhost:3456", "ROUTER", "42", nil)
	e.LoggregatorClient = &MockLoggregatorClient{received}

	logMessage := testhelpers.NewLogMessage("test_msg", "test_app_id")
	logMessage.SourceId = proto.String("src_id")
	e.EmitLogMessage(logMessage)
	receivedMessage := extractLogMessage(t, <-received)

	assert.Equal(t, receivedMessage.GetMessage(), []byte("test_msg"))
	assert.Equal(t, receivedMessage.GetAppId(), "test_app_id")
	assert.Equal(t, receivedMessage.GetSourceId(), "src_id")
}

func TestLogMessageEmitTruncatesLargeMessages(t *testing.T) {
	received := make(chan *[]byte, 1)
	e, _ := NewLogMessageEmitter("localhost:3456", "ROUTER", "42", nil)
	e.LoggregatorClient = &MockLoggregatorClient{received}

	message := ""
	for i := 0; i < MAX_MESSAGE_BYTE_SIZE*2; i++ {
		message += "a"
	}
	logMessage := testhelpers.NewLogMessage(message, "test_app_id")
	logMessage.SourceId = proto.String("src_id")
	e.EmitLogMessage(logMessage)
	receivedMessage := extractLogMessage(t, <-received)

	receivedMessageText := receivedMessage.GetMessage()
	trucatedOffset := len(receivedMessageText) - len(TRUNCATED_BYTES)
	assert.Equal(t, receivedMessageText[trucatedOffset:], TRUNCATED_BYTES)
	assert.True(t, len(receivedMessageText) >= MAX_MESSAGE_BYTE_SIZE)
}

func TestLogMessageEmitSplitsMessagesOnNewlines(t *testing.T) {
	received := make(chan *[]byte, 10)
	e, _ := NewLogMessageEmitter("localhost:3456", "ROUTER", "42", nil)
	e.LoggregatorClient = &MockLoggregatorClient{received}

	message := "hi\n\rworld\nhow are you\r\ndoing\r"
	logMessage := testhelpers.NewLogMessage(message, "test_app_id")
	logMessage.SourceId = proto.String("src_id")
	e.EmitLogMessage(logMessage)

	assert.Equal(t, len(received), 4)
}

func TestLogEnvelopeEmitter(t *testing.T) {
	received := make(chan *[]byte, 1)
	e, _ := NewLogEnvelopeEmitter("localhost:3456", "ROUTER", "42", "secret", nil)
	e.LoggregatorClient = &MockLoggregatorClient{received}
	e.Emit("appid", "foo")
	receivedEnvelope := extractLogEnvelope(t, <-received)

	assert.Equal(t, receivedEnvelope.GetLogMessage().GetMessage(), []byte("foo"))
	assert.Equal(t, receivedEnvelope.GetLogMessage().GetAppId(), "appid")
	assert.Equal(t, receivedEnvelope.GetRoutingKey(), "appid")
	assert.Equal(t, receivedEnvelope.GetLogMessage().GetSourceId(), "42")
}

func TestLogEnvelopeValidRoutinKeyInTheEnvelope(t *testing.T) {
	received := make(chan *[]byte, 1)
	e, _ := NewLogEnvelopeEmitter("localhost:3456", "ROUTER", "42", "secret", nil)
	e.LoggregatorClient = &MockLoggregatorClient{received}
	e.Emit("appid", "foo")
	receivedEnvelope := extractLogEnvelope(t, <-received)

	assert.Equal(t, receivedEnvelope.GetRoutingKey(), "appid")
}

func TestLogEnvelopeSignatureInTheEnvelope(t *testing.T) {
	sharedKey := "shared key"

	received := make(chan *[]byte, 1)
	e, _ := NewLogEnvelopeEmitter("localhost:3456", "ROUTER", "42", sharedKey, nil)
	e.LoggregatorClient = &MockLoggregatorClient{received}
	e.Emit("appid", "foo")
	receivedEnvelope := extractLogEnvelope(t, <-received)

	expectedDigest := signature.Digest(receivedEnvelope.GetLogMessage().String())
	extractedDigest, err := signature.Decrypt(sharedKey, receivedEnvelope.GetSignature())
	assert.NoError(t, err)

	assert.Equal(t, extractedDigest, expectedDigest)
}

var emitters = []func(bool) (*loggregatoremitter, error){
	func(valid bool) (*loggregatoremitter, error) {
		if valid {
			return NewLogMessageEmitter("localhost:38452", "ROUTER", "42", nil)
		} else {
			return NewLogMessageEmitter("server", "FOOSERVER", "42", nil)
		}
	},
	func(valid bool) (*loggregatoremitter, error) {
		if valid {
			return NewEmitter("localhost:38452", "ROUTER", "42", nil)
		} else {
			return NewEmitter("server", "FOOSERVER", "42", nil)
		}
	},
	func(valid bool) (*loggregatoremitter, error) {
		if valid {
			return NewLogEnvelopeEmitter("localhost:38452", "ROUTER", "42", "secret", nil)
		} else {
			return NewLogEnvelopeEmitter("server", "FOOSERVER", "42", "secret", nil)
		}
	},
}

func TestLogEnvelopeInvalidSourcetype(t *testing.T) {
	for _, emitter := range emitters {
		_, err := emitter(false)
		assert.Error(t, err)
	}
}

func TestLogEnvelopeValidSourcetype(t *testing.T) {
	for _, emitter := range emitters {
		_, err := emitter(true)
		assert.NoError(t, err)
	}
}

func TestLogEnvelopeEmptyAppIdDoesNotEmit(t *testing.T) {
	for _, emitter := range emitters {
		received := make(chan *[]byte, 1)
		e, _ := emitter(true)
		e.LoggregatorClient = &MockLoggregatorClient{received}

		e.Emit("", "foo")
		select {
		case <-received:
			t.Error("This message should not have been emitted since it does not have an AppId")
		default:
			// success
		}

		e.Emit("    ", "foo")
		select {
		case <-received:
			t.Error("This message should not have been emitted since it does not have an AppId")
		default:
			// success
		}
	}
}

func TestLogEnvelopeEmptyMessageDoesNotEmit(t *testing.T) {
	for _, emitter := range emitters {

		received := make(chan *[]byte, 1)
		e, _ := emitter(true)
		e.LoggregatorClient = &MockLoggregatorClient{received}

		e.Emit("appId", "")
		select {
		case <-received:
			t.Error("This message should not have been emitted since it does not have a message")
		default:
			// success
		}

		e.Emit("appId", "   ")
		select {
		case <-received:
			t.Error("This message should not have been emitted since it does not have a message")
		default:
			// success
		}
	}
}

func extractLogEnvelope(t *testing.T, data *[]byte) *logmessage.LogEnvelope {
	receivedEnvelope := &logmessage.LogEnvelope{}

	err := proto.Unmarshal(*data, receivedEnvelope)

	if err != nil {
		t.Fatalf("Envelope invalid. %s", err)
	}
	return receivedEnvelope
}

func extractLogMessage(t *testing.T, data *[]byte) *logmessage.LogMessage {
	receivedMessage := &logmessage.LogMessage{}

	err := proto.Unmarshal(*data, receivedMessage)

	if err != nil {
		t.Fatalf("Message invalid. %s", err)
	}
	return receivedMessage
}
