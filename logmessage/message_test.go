package logmessage

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMessage(t *testing.T) {
	appMessageString := "AppMessage"

	unmarshalledMessage := NewMessage(appMessageString, "myApp")
	marshalledMessage := MarshallLogMessage(t, unmarshalledMessage)

	message, err := ParseMessage(marshalledMessage)
	assert.NoError(t, err)

	assert.Equal(t, uint32(33), message.GetRawMessageLength())
	assert.Equal(t, marshalledMessage, message.GetRawMessage())
	assert.Equal(t, unmarshalledMessage, message.GetLogMessage())
}

func NewMessage(messageString, appId string) *LogMessage {
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
