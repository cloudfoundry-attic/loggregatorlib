package logmessage

import (
	"code.google.com/p/gogoprotobuf/proto"
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

	assert.Equal(t, uint32(36), message.GetRawMessageLength())
	assert.Equal(t, marshalledMessage, message.GetRawMessage())
	assert.Equal(t, unmarshalledMessage, message.GetLogMessage())
	assert.Equal(t, "APP", message.GetShortSourceTypeName())
}

func TestExtractionFromEnvelope(t *testing.T) {
	appMessageString := "AppMessage"

	unmarshalledMessage := NewLogMessage(t, appMessageString, "myApp")
	marshalledMessage := MarshallLogMessage(t, unmarshalledMessage)
	marshalledEnvelope := MarshalledLogEnvelope(t, unmarshalledMessage, "some secret")

	message, err := ParseProtobuffer(marshalledEnvelope)
	assert.NoError(t, err)

	assert.Equal(t, uint32(36), message.GetRawMessageLength())
	assert.Equal(t, marshalledMessage, message.GetRawMessage())
	assert.Equal(t, unmarshalledMessage, message.GetLogMessage())
	assert.Equal(t, "APP", message.GetShortSourceTypeName())
}

func TestExtractEnvelopeFromRawBytes(t *testing.T) {
	//This allows us to verify that the same extraction can be done on the Ruby side

	// The following code is used to produce the `data` variable below.

	//	currentTime := time.Now()
	//
	//	messageType := LogMessage_OUT
	//	sourceType := "APP"
	//	sourceId := "42"
	//	protoMessage := &LogMessage{
	//		Message:     []byte("Hello there!"),
	//		AppId:       proto.String("my_app_id"),
	//		MessageType: &messageType,
	//		SourceType:  &sourceType,
	//		SourceId:    &sourceId,
	//		Timestamp:   proto.Int64(currentTime.UnixNano()),
	//	}
	//
	//	marshalledLogEnvelope := MarshalledLogEnvelope(t, protoMessage, "secret")
	//
	//	for _, value := range marshalledLogEnvelope {
	//		print(uint8(value))
	//		print(", ")
	//	}

	data := []uint8{10, 9, 109, 121, 95, 97, 112, 112, 95, 105, 100, 18, 64, 84, 223, 196, 25, 252, 195, 218, 43, 86, 158, 108, 160, 133, 104, 142, 105, 9, 92, 81, 232, 82, 66, 130, 159, 246, 51, 40, 155, 114, 30, 250, 206, 146, 205, 73, 183, 147, 226, 18, 71, 53, 93, 254, 66, 40, 45, 202, 234, 107, 110, 186, 140, 114, 206, 55, 217, 192, 48, 18, 137, 219, 197, 146, 169, 26, 46, 10, 12, 72, 101, 108, 108, 111, 32, 116, 104, 101, 114, 101, 33, 16, 1, 24, 202, 128, 245, 211, 165, 147, 165, 178, 38, 34, 9, 109, 121, 95, 97, 112, 112, 95, 105, 100, 42, 3, 65, 80, 80, 50, 2, 52, 50}

	receivedEnvelope := &LogEnvelope{}
	err := proto.Unmarshal(data, receivedEnvelope)
	assert.NoError(t, err)
	assert.Equal(t, receivedEnvelope.GetLogMessage().GetMessage(), []byte("Hello there!"))
	assert.Equal(t, receivedEnvelope.GetLogMessage().GetAppId(), "my_app_id")
	assert.Equal(t, receivedEnvelope.GetLogMessage().GetSourceId(), "42")
	assert.Equal(t, receivedEnvelope.GetRoutingKey(), "my_app_id")

	assert.True(t, receivedEnvelope.VerifySignature("secret"))
}

func TestThatSignatureValidatesWhenItMatches(t *testing.T) {
	secret := "super-secret"
	logMessage := NewLogMessage(t, "the logs", "appid")

	envelope := &LogEnvelope{
		LogMessage: logMessage,
		RoutingKey: proto.String(*logMessage.AppId),
	}
	envelope.SignEnvelope(secret)

	assert.True(t, envelope.VerifySignature(secret))
}

func TestThatSignatureDoesNotValidateWhenItDoesntMatch(t *testing.T) {
	envelope := &LogEnvelope{
		LogMessage: &LogMessage{},
		RoutingKey: proto.String("app_id"),
		Signature:  []byte{0, 1, 2}, //some bad signature
	}

	assert.False(t, envelope.VerifySignature("super-secret"))
}

func TestThatSignatureDoesNotValidateWhenSecretIsIncorrect(t *testing.T) {
	secret := "super-secret"
	logMessage := NewLogMessage(t, "the logs", "appid")

	envelope := &LogEnvelope{
		LogMessage: logMessage,
		RoutingKey: proto.String(*logMessage.AppId),
	}
	envelope.SignEnvelope(secret)

	assert.False(t, envelope.VerifySignature(secret+"not the right secret"))
}

func NewLogMessage(t *testing.T, messageString, appId string) *LogMessage {
	currentTime := time.Now()

	messageType := LogMessage_OUT
	sourceType := "APP"
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
	envelope := &LogEnvelope{
		LogMessage: unmarshalledMessage,
		RoutingKey: proto.String(*unmarshalledMessage.AppId),
	}
	envelope.SignEnvelope(secret)

	marshalledEnvelope, err := proto.Marshal(envelope)
	assert.NoError(t, err)

	return marshalledEnvelope
}

func UnmarshalLogEnvelope(t *testing.T, data []byte) *LogEnvelope {
	logEnvelope := new(LogEnvelope)
	err := proto.Unmarshal(data, logEnvelope)
	assert.NoError(t, err)
	return logEnvelope
}
