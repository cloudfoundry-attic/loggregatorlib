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

func (m *Message) GetShortSourceTypeName() string {
	sourceTypeNames := map[LogMessage_SourceType]string{
		LogMessage_CLOUD_CONTROLLER: "API",
		LogMessage_ROUTER:           "RTR",
		LogMessage_UAA:              "UAA",
		LogMessage_DEA:              "DEA",
		LogMessage_WARDEN_CONTAINER: "App",
	}

	return sourceTypeNames[m.logMessage.GetSourceType()]
}
