package emitter

import (
	"code.google.com/p/gogoprotobuf/proto"
	"fmt"
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/loggregatorclient"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"strings"
	"time"
)

type Emitter interface {
	Emit(string, string)
}

type loggregatoremitter struct {
	lc     loggregatorclient.LoggregatorClient
	st     logmessage.LogMessage_SourceType
	sId    string
	logger *gosteno.Logger
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
	data, err := proto.Marshal(logMessage)
	if err != nil {
		e.logger.Errorf("Error marshalling message: %s", err)
		return
	}
	e.lc.Send(data)
}

func NewEmitter(loggregatorServer, sourceType, sourceId string, logger *gosteno.Logger) (e *loggregatoremitter, err error) {
	if logger == nil {
		logger = gosteno.NewLogger("loggregatorlib.emitter")
	}

	e = new(loggregatoremitter)

	if name, ok := logmessage.LogMessage_SourceType_value[sourceType]; ok {
		e.st = logmessage.LogMessage_SourceType(name)
	} else {
		err = fmt.Errorf("Unable to map SourceType [%s] to a logmessage.LogMessage_SourceType", sourceType)
		return
	}

	e.logger = logger
	e.lc = loggregatorclient.NewLoggregatorClient(loggregatorServer, logger, loggregatorclient.DefaultBufferSize)
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
