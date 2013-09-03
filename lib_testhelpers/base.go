package lib_testhelpers

import (
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"os"
	"testing"
	"github.com/stretchr/testify/assert"
	"code.google.com/p/gogoprotobuf/proto"
	"time"
)

func Logger() *gosteno.Logger {
	return getLogger(false)
}

func getLogger(debug bool) *gosteno.Logger {
	if debug {
		level := gosteno.LOG_DEBUG

		loggingConfig := &gosteno.Config{
			Sinks:     make([]gosteno.Sink, 1),
			Level:     level,
			Codec:     gosteno.NewJsonCodec(),
			EnableLOC: true,
		}

		loggingConfig.Sinks[0] = gosteno.NewIOSink(os.Stdout)

		gosteno.Init(loggingConfig)
	}

	return gosteno.NewLogger("TestLogger")
}


func MarshalledLogMessage(t *testing.T, messageString string, appId string) []byte {
	currentTime := time.Now()

	messageType := logmessage.LogMessage_OUT
	sourceType := logmessage.LogMessage_DEA
	protoMessage := &logmessage.LogMessage{
		Message:     []byte(messageString),
		AppId:       proto.String(appId),
			MessageType: &messageType,
			SourceType:  &sourceType,
			Timestamp:   proto.Int64(currentTime.UnixNano()),
		}

	message, err := proto.Marshal(protoMessage)
	assert.NoError(t, err)

	return message
}
