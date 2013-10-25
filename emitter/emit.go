package emitter

import (
	"code.google.com/p/gogoprotobuf/proto"
	"fmt"
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/loggregatorclient"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/cloudfoundry/loggregatorlib/signature"
	"strings"
	"time"
)

type Emitter interface {
	Emit(string, string)
	EmitLogMessage(*logmessage.LogMessage)
}

type loggregatoremitter struct {
	LoggregatorClient loggregatorclient.LoggregatorClient
	st                logmessage.LogMessage_SourceType
	sId               string
	sharedSecret      string
	logger            *gosteno.Logger
}

func isEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

func (e *loggregatoremitter) Emit(appid, message string) {
	if isEmpty(appid) || isEmpty(message) {
		return
	}
	logMessage := e.newLogMessage(appid, message)
	e.logger.Debugf("Logging message from %s of type %s with appid %s and with data %s", logMessage.SourceType, logMessage.MessageType, logMessage.AppId, string(logMessage.Message))

	e.EmitLogMessage(logMessage)
}

func (e *loggregatoremitter) EmitLogMessage(logMessage *logmessage.LogMessage) {
	if e.sharedSecret == "" {
		marshalledLogMessage, err := proto.Marshal(logMessage)
		if err != nil {
			e.logger.Errorf("Error marshalling message: %s", err)
			return
		}
		e.LoggregatorClient.Send(marshalledLogMessage)
	} else {
		logEnvelope := e.newLogEnvelope(*logMessage.AppId, logMessage)
		marshalledLogEnvelope, err := proto.Marshal(logEnvelope)
		if err != nil {
			e.logger.Errorf("Error marshalling envelope: %s", err)
			return
		}
		e.LoggregatorClient.Send(marshalledLogEnvelope)
	}
}

func NewLogMessageEmitter(loggregatorServer, sourceType, sourceId string, logger *gosteno.Logger) (e *loggregatoremitter, err error) {
	return NewLogEnvelopeEmitter(loggregatorServer, sourceType, sourceId, "", logger)
}

func NewLogEnvelopeEmitter(loggregatorServer, sourceType, sourceId, sharedSecret string, logger *gosteno.Logger) (e *loggregatoremitter, err error) {
	if logger == nil {
		logger = gosteno.NewLogger("loggregatorlib.emitter")
	}

	e = &loggregatoremitter{sharedSecret: sharedSecret}

	if name, ok := logmessage.LogMessage_SourceType_value[sourceType]; ok {
		e.st = logmessage.LogMessage_SourceType(name)
	} else {

		err = fmt.Errorf("Unable to map SourceType [%s] to a logmessage.LogMessage_SourceType", sourceType)
		return
	}

	e.logger = logger
	e.LoggregatorClient = loggregatorclient.NewLoggregatorClient(loggregatorServer, logger, loggregatorclient.DefaultBufferSize)
	e.sId = sourceId

	return
}

func (e *loggregatoremitter) newLogMessage(appId, message string) *logmessage.LogMessage {
	currentTime := time.Now()
	mt := logmessage.LogMessage_OUT

	return &logmessage.LogMessage{
		Message:     []byte(message),
		AppId:       proto.String(appId),
		MessageType: &mt,
		SourceType:  &e.st,
		SourceId:    &e.sId,
		Timestamp:   proto.Int64(currentTime.UnixNano()),
	}
}

func (e *loggregatoremitter) newLogEnvelope(appId string, message *logmessage.LogMessage) *logmessage.LogEnvelope {
	return &logmessage.LogEnvelope{
		LogMessage: message,
		RoutingKey: proto.String(appId),
		Signature:  e.sign(message),
	}
}

func (e *loggregatoremitter) sign(logMessage *logmessage.LogMessage) []byte {
	signatureOfMessage, _ := signature.Encrypt(e.sharedSecret, signature.Digest(logMessage.String()))
	return signatureOfMessage
}
