package logmessage

type Message struct {
	logMessage       *LogMessage
	rawMessage       []byte
	rawMessageLength uint32
}

func ParseMessage(data []byte) (*Message, error) {
	logMessage, err := parseLogMessage(data)
	return &Message{logMessage, data, uint32(len(data))}, err
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
