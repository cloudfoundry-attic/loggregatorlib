package testhelpers

import (
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"time"
	"testing"
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/stretchr/testify/assert"
)

func MarshalledErrorLogMessage(t *testing.T, messageString string, appId string) []byte {
	messageType := logmessage.LogMessage_ERR
	sourceType := logmessage.LogMessage_DEA
	protoMessage := generateLogMessage(messageString, appId, messageType, sourceType)

	return marshalProtoBuf(t, protoMessage)
}

func MarshalledLogMessage(t *testing.T, messageString string, appId string) []byte {
	messageType := logmessage.LogMessage_OUT
	sourceType := logmessage.LogMessage_DEA
	protoMessage := generateLogMessage(messageString, appId, messageType, sourceType)

	return marshalProtoBuf(t, protoMessage)
}

func MarshalledDrainedLogMessage(t *testing.T, messageString string, appId string, drainUrls ...string) []byte {
	messageType := logmessage.LogMessage_OUT
	sourceType := logmessage.LogMessage_WARDEN_CONTAINER
	protoMessage := generateLogMessage(messageString, appId, messageType, sourceType)
	protoMessage.DrainUrls = drainUrls

	return marshalProtoBuf(t, protoMessage)
}

func MarshalledDrainedNonWardenLogMessage(t *testing.T, messageString string, appId string, drainUrls ...string) []byte {
	messageType := logmessage.LogMessage_OUT
	sourceType := logmessage.LogMessage_DEA
	protoMessage := generateLogMessage(messageString, appId, messageType, sourceType)
	protoMessage.DrainUrls = drainUrls

	return marshalProtoBuf(t, protoMessage)
}


func NewLogMessage(messageString, appId string) *logmessage.LogMessage {
	messageType := logmessage.LogMessage_OUT
	sourceType := logmessage.LogMessage_WARDEN_CONTAINER

	return generateLogMessage(messageString, appId, messageType, sourceType)
}

func NewMessage(t *testing.T, messageString, appId string) *logmessage.Message {
	currentTime := time.Now()
	messageType := logmessage.LogMessage_OUT
	sourceType := logmessage.LogMessage_WARDEN_CONTAINER

	logMessage := &logmessage.LogMessage{
		Message:     []byte(messageString),
		AppId:       proto.String(appId),
		MessageType: &messageType,
		SourceType:  &sourceType,
		Timestamp:   proto.Int64(currentTime.UnixNano()),
	}
	marshalledLogMessage, err := proto.Marshal(logMessage)
	assert.NoError(t, err)

	return logmessage.NewMessage(logMessage, marshalledLogMessage)
}

func MarshalledLogEnvelopeForMessage(t *testing.T, msg, appName, secret string, drainUrls ...string) []byte {
	logMessage := NewLogMessage(msg, appName)
	logMessage.DrainUrls = drainUrls
	return MarshalledLogEnvelope(t, logMessage, secret)
}

func MarshalledLogEnvelope(t *testing.T, unmarshalledMessage *logmessage.LogMessage, secret string) []byte {
	envelope := &logmessage.LogEnvelope{
		LogMessage: unmarshalledMessage,
		RoutingKey: proto.String(*unmarshalledMessage.AppId),
	}
	envelope.SignEnvelope(secret)

	return marshalProtoBuf(t, envelope)
}

func AssertProtoBufferMessageEquals(t *testing.T, expectedMessage string, actual []byte) {
	receivedMessage := &logmessage.LogMessage{}
	err := proto.Unmarshal(actual, receivedMessage)
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, string(receivedMessage.GetMessage()))
}

func generateLogMessage(messageString, appId string, messageType logmessage.LogMessage_MessageType, sourceType logmessage.LogMessage_SourceType) *logmessage.LogMessage {
	currentTime := time.Now()
	logMessage := &logmessage.LogMessage{
		Message:     []byte(messageString),
		AppId:       proto.String(appId),
		MessageType: &messageType,
		SourceType:  &sourceType,
		Timestamp:   proto.Int64(currentTime.UnixNano()),
	}

	return logMessage
}

func marshalProtoBuf(t *testing.T, pb proto.Message) []byte {
	marshalledProtoBuf, err := proto.Marshal(pb)
	assert.NoError(t, err)

	return marshalledProtoBuf
}
