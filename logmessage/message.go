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

func ParseProtobuffer(data []byte, secret string) (*Message, error) {
	message := &Message{}

	err := message.parseProtoBuffer(data, secret)
	if err != nil {
		return nil, err
	}

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

func (m *Message) parseProtoBuffer(data []byte, secret string) error {
	logMessage := new(LogMessage)
	err := proto.Unmarshal(data, logMessage)
	if err == nil {
		m.logMessage = logMessage
		m.rawMessage = data
		m.rawMessageLength = uint32(len(m.rawMessage))
		return nil
	}

	logEnvelope := new(LogEnvelope)
	err = proto.Unmarshal(data, logEnvelope)
	if err == nil {
		if !logEnvelope.VerifySignature(secret) {
			return errors.New("Invalid Envelope Signature")
		}

		m.logMessage = logEnvelope.LogMessage
		m.rawMessage, err = proto.Marshal(m.logMessage)
		if err == nil {
			m.rawMessageLength = uint32(len(m.rawMessage))
			return nil
		}
		m.rawMessageLength = uint32(len(m.rawMessage))
		return err
	}

	return err
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
