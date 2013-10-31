package logmessage

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/signature"
)

type Message struct {
	logMessage       *LogMessage
	rawMessage       []byte
	rawMessageLength uint32
}

func ParseProtobuffer(data []byte, logger *gosteno.Logger) (*Message, error) {
	message := &Message{}

	err := message.parseProtoBuffer(data, logger)
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
	return m.logMessage.GetSourceType()
}

func (m *Message) parseProtoBuffer(data []byte, logger *gosteno.Logger) error {
	logMessage := new(LogMessage)
	err := proto.Unmarshal(data, logMessage)
	if err == nil {
		m.logMessage = logMessage
		m.rawMessage = data
		m.rawMessageLength = uint32(len(m.rawMessage))
		logger.Debugf("Data unmarshalled into LogMessage: %s.", logMessage.String())
		return nil
	}
	logger.Debugf("Error unmarshalling into LogMessage: %s.", err)

	logEnvelope := new(LogEnvelope)
	err = proto.Unmarshal(data, logEnvelope)
	if err == nil {
		m.logMessage = logEnvelope.LogMessage
		m.rawMessage, err = proto.Marshal(m.logMessage)
		if err == nil {
			m.rawMessageLength = uint32(len(m.rawMessage))
			logger.Debugf("Data unmarshalled into LogEnvelope: %s.", logEnvelope.String())
			return nil
		}
		logger.Debugf("Error marshaling into rawMessage: %s.", err)
		m.rawMessageLength = uint32(len(m.rawMessage))
		return err
	}

	logger.Debugf("Error unmarshalling into LogEnvelope: %s.", err)
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

func (e *LogEnvelope) SignEnvelope(sharedSecret string) {
	e.Signature, _ = signature.Encrypt(sharedSecret, e.logMessageDigest())
}

func (e *LogEnvelope) logMessageDigest() []byte {
	return signature.DigestBytes(e.LogMessage.GetMessage())
}
