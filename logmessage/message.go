package logmessage

import (
	"code.google.com/p/gogoprotobuf/proto"
)

type Message struct {
	logMessage       *LogMessage
	rawMessage       []byte
	rawMessageLength uint32
}

func ParseMessage(data []byte) (*Message, error) {
	logMessage := new(LogMessage)
	err := proto.Unmarshal(data, logMessage)
	if err != nil {
		return new(Message), err
	}
	return &Message{logMessage, data, uint32(len(data))}, nil
}

func (m *Message) GetLogMessage() *LogMessage {
	return m.logMessage
}

func (m *Message) GetRawMessage() []byte {
	return m.rawMessage
}

func (m *Message) GetRawMessageLength() uint32 {
	return m.rawMessageLength
}
