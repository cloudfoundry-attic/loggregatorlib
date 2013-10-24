package logmessage

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/loggregatorlib/signature"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestExtractionFromMessage(t *testing.T) {
	appMessageString := "AppMessage"

	unmarshalledMessage := NewLogMessage(t, appMessageString, "myApp")
	marshalledMessage := MarshallLogMessage(t, unmarshalledMessage)

	message, err := ParseProtobuffer(marshalledMessage)
	assert.NoError(t, err)

	assert.Equal(t, uint32(33), message.GetRawMessageLength())
	assert.Equal(t, marshalledMessage, message.GetRawMessage())
	assert.Equal(t, unmarshalledMessage, message.GetLogMessage())
	assert.Equal(t, "App", message.GetShortSourceTypeName())
}

func TestExtractionFromEnvelope(t *testing.T) {
	appMessageString := "AppMessage"

	unmarshalledMessage := NewLogMessage(t, appMessageString, "myApp")
	marshalledMessage := MarshallLogMessage(t, unmarshalledMessage)
	marshalledEnvelope := MarshalledLogEnvelope(t, unmarshalledMessage, "some secret")

	message, err := ParseProtobuffer(marshalledEnvelope)
	assert.NoError(t, err)

	assert.Equal(t, uint32(33), message.GetRawMessageLength())
	assert.Equal(t, marshalledMessage, message.GetRawMessage())
	assert.Equal(t, unmarshalledMessage, message.GetLogMessage())
	assert.Equal(t, "App", message.GetShortSourceTypeName())
}

func NewLogMessage(t *testing.T, messageString, appId string) *LogMessage {
	currentTime := time.Now()

	messageType := LogMessage_OUT
	sourceType := LogMessage_WARDEN_CONTAINER
	protoMessage := &LogMessage{
		Message:     []byte(messageString),
		AppId:       proto.String(appId),
		MessageType: &messageType,
		SourceType:  &sourceType,
		Timestamp:   proto.Int64(currentTime.UnixNano()),
	}
	return protoMessage
}

func MarshallLogMessage(t *testing.T, unmarshalledMessage *LogMessage) []byte {
	message, err := proto.Marshal(unmarshalledMessage)
	assert.NoError(t, err)

	return message
}

func MarshalledLogEnvelope(t *testing.T, unmarshalledMessage *LogMessage, secret string) []byte {
	signatureOfMessage, err := signature.Encrypt(secret, signature.Digest(unmarshalledMessage.String()))
	assert.NoError(t, err)

	envelope := &LogEnvelope{
		LogMessage: unmarshalledMessage,
		RoutingKey: proto.String(*unmarshalledMessage.AppId),
		Signature:  signatureOfMessage,
	}

	marshalledEnvelope, err := proto.Marshal(envelope)
	assert.NoError(t, err)

	return marshalledEnvelope
}
