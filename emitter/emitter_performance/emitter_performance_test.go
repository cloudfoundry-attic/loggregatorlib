package emitter_performance

import (
	"code.google.com/p/gogoprotobuf/proto"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent/instrumentation"
	"github.com/cloudfoundry/loggregatorlib/emitter"
	"github.com/cloudfoundry/loggregatorlib/logmessage/testhelpers"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const SECOND = int64(1 * time.Second)

type MockLoggregatorClient struct {
	received chan *[]byte
}

func (m MockLoggregatorClient) Send(data []byte) {
	m.received <- &data
}

func (m MockLoggregatorClient) Emit() instrumentation.Context {
	return instrumentation.Context{}
}

func Test1000LogMessageEmit(t *testing.T) {
	received := make(chan *[]byte, 1)
	e, _ := emitter.NewLogMessageEmitter("localhost:3456", "ROUTER", "42", nil)
	e.LoggregatorClient = &MockLoggregatorClient{received}

	message := longMessage()

	startTime := time.Now().UnixNano()
	logMessage := testhelpers.NewLogMessage(message, "test_app_id")
	logMessage.SourceId = proto.String("src_id")
	for i := 0; i < 1000; i++ {
		e.EmitLogMessage(logMessage)
		<-received
	}

	elapsedTime := time.Now().UnixNano() - startTime
	assert.True(t, elapsedTime < SECOND, fmt.Sprintf("Elapsed time should have been below 1s, but was %vs", float64(elapsedTime)/float64(SECOND)))
}

func Test1000LogEnvelopeEmit(t *testing.T) {
	received := make(chan *[]byte, 1)
	e, _ := emitter.NewLogEnvelopeEmitter("localhost:3456", "ROUTER", "42", "secret", nil)
	e.LoggregatorClient = &MockLoggregatorClient{received}

	message := longMessage()

	startTime := time.Now().UnixNano()
	for i := 0; i < 1000; i++ {
		e.Emit("appid", message)
		<-received
	}

	elapsedTime := time.Now().UnixNano() - startTime
	s := fmt.Sprintf("Elapsed time should have been below 1s, but was %v s", float64(elapsedTime/SECOND))
	println(s)
	assert.True(t, elapsedTime < 2*SECOND, fmt.Sprintf("Elapsed time should have been below 2s, but was %vs", float64(elapsedTime)/float64(SECOND)))
}

func longMessage() string {
	message := ""
	for i := 0; i < emitter.MAX_MESSAGE_BYTE_SIZE*2; i++ {
		message += "a"
	}
	return message
}
