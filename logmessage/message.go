package logmessage

import (
	"code.google.com/p/gogoprotobuf/proto"
	"errors"
	"github.com/cloudfoundry/loggregatorlib/signature"
)

type Message struct {
	logMessage       *LogMessage
	rawMessage       []byte
	rawMessageLength uint32
}

func ParseMessage(data []byte) (*Message, error) {
	logMessage, err := parseLogMessage(data)
	return &Message{logMessage, data, uint32(len(data))}, err
}

func ParseEnvelope(data []byte, secret string) (message *Message, err error) {
	message = &Message{}
	logEnvelope := &LogEnvelope{}

	if err := proto.Unmarshal(data, logEnvelope); err != nil {
		return nil, err
	}
	if !logEnvelope.VerifySignature(secret) {
		return nil, errors.New("Invalid Envelope Signature")
	}

	message.rawMessage, err = proto.Marshal(logEnvelope.LogMessage)
	if err != nil {
		return nil, err
	}

	message.logMessage = logEnvelope.LogMessage
	message.rawMessageLength = uint32(len(message.rawMessage))
	return message, nil
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

	sourceName := sourceTypeNames[m.logMessage.GetSourceType()]
	if sourceName == "" {
		sourceName = m.logMessage.GetSourceName()
	}

	return sourceName
}

func (e *LogEnvelope) VerifySignature(sharedSecret string) bool {
	messageDigest, err := signature.Decrypt(sharedSecret, e.GetSignature())
	if err != nil {
		return false
	}

	expectedDigest := e.logMessageDigest()
	return string(messageDigest) == string(expectedDigest)
}

func (e *LogEnvelope) SignEnvelope(sharedSecret string) error {
	signature, err := signature.Encrypt(sharedSecret, e.logMessageDigest())
	if err == nil {
		e.Signature = signature
	}

	return err
}

func (e *LogEnvelope) logMessageDigest() []byte {
	return signature.DigestBytes(e.LogMessage.GetMessage())
}
