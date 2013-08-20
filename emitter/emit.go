package emitter

import (
	"code.google.com/p/gogoprotobuf/proto"
	"fmt"
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/loggregatorclient"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"time"
)

type emitter struct {
	lc     loggregatorclient.LoggregatorClient
	st     logmessage.LogMessage_SourceType
	logger *gosteno.Logger
}

func (e *emitter) Emit(appid, message string) {
	data, err := proto.Marshal(e.newLogMessage(appid, message))
	if err != nil {
		e.logger.Errorf("Error marshalling message: %s", err)
		return
	}

	go e.lc.Send(data)
}

func NewEmitter(loggregatorServer, sourceType string, logger *gosteno.Logger) (e *emitter, err error) {
	if logger == nil {
		logger = gosteno.NewLogger("loggregatorlib.emitter")
	}

	e = new(emitter)

	if name, ok := logmessage.LogMessage_SourceType_value[sourceType]; ok {
		e.st = logmessage.LogMessage_SourceType(name)
	} else {
		err = fmt.Errorf("Unable to map SourceType [%s] to a logmessage.LogMessage_SourceType", sourceType)
		return
	}

	e.logger = logger
	e.lc = loggregatorclient.NewLoggregatorClient(loggregatorServer, logger, loggregatorclient.DefaultBufferSize)

	return
}

func (e *emitter) newLogMessage(appId, message string) *logmessage.LogMessage {
	currentTime := time.Now()
	mt := logmessage.LogMessage_OUT

	return &logmessage.LogMessage{
		Message:     []byte(message),
		AppId:       proto.String(appId),
		MessageType: &mt,
		SourceType:  &e.st,
		Timestamp:   proto.Int64(currentTime.UnixNano()),
	}
}
