package logmessage

import (
	"bytes"
	"code.google.com/p/gogoprotobuf/proto"
	"encoding/binary"
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

func ParseDumpedMessages(b []byte) (messages []*Message, err error) {
	buffer := bytes.NewBuffer(b)
	var length uint32
	for buffer.Len() > 0 {
		lengthBytes := bytes.NewBuffer(buffer.Next(4))
		err = binary.Read(lengthBytes, binary.BigEndian, &length)
		if err != nil {
			return
		}

		msgBytes := buffer.Next(int(length))
		var msg *Message
		msg, err = ParseMessage(msgBytes)
		if err != nil {
			return
		}
		messages = append(messages, msg)
	}
	return
}

func DumpMessage(msg Message, buffer *bytes.Buffer) {
	binary.Write(buffer, binary.BigEndian, msg.GetRawMessageLength())
	buffer.Write(msg.GetRawMessage())
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
