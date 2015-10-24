package emitter_test

import (
	"strings"

	. "github.com/cloudfoundry/loggregatorlib/emitter"
	"github.com/cloudfoundry/loggregatorlib/loggregatorclient/fakeclient"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/cloudfoundry/loggregatorlib/logmessage/testhelpers"
	"github.com/gogo/protobuf/proto"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing with Ginkgo", func() {
	var (
		emitter *LoggregatorEmitter
		client  *fakeclient.FakeClient
	)

	BeforeEach(func() {
		var err error
		emitter, err = NewEmitter("localhost:3456", "ROUTER", "42", "secret", nil)
		Expect(err).ToNot(HaveOccurred())

		client = &fakeclient.FakeClient{}
		emitter.LoggregatorClient = client

	})

	It("should emit stdout", func() {
		emitter.Emit("appid", "foo")
		receivedMessage := extractLogMessage(client.WriteArgsForCall(0))

		Expect(receivedMessage.GetMessage()).To(Equal([]byte("foo")))
		Expect(receivedMessage.GetAppId()).To(Equal("appid"))
		Expect(receivedMessage.GetSourceId()).To(Equal("42"))
		Expect(receivedMessage.GetMessageType()).To(Equal(logmessage.LogMessage_OUT))
	})

	It("should emit stderr", func() {
		emitter.EmitError("appid", "foo")
		receivedMessage := extractLogMessage(client.WriteArgsForCall(0))

		Expect(receivedMessage.GetMessage()).To(Equal([]byte("foo")))
		Expect(receivedMessage.GetAppId()).To(Equal("appid"))
		Expect(receivedMessage.GetSourceId()).To(Equal("42"))
		Expect(receivedMessage.GetMessageType()).To(Equal(logmessage.LogMessage_ERR))
	})

	It("should emit fully formed log messages", func() {
		logMessage := testhelpers.NewLogMessage("test_msg", "test_app_id")
		logMessage.SourceId = proto.String("src_id")

		emitter.EmitLogMessage(logMessage)
		receivedMessage := extractLogMessage(client.WriteArgsForCall(0))

		Expect(receivedMessage.GetMessage()).To(Equal([]byte("test_msg")))
		Expect(receivedMessage.GetAppId()).To(Equal("test_app_id"))
		Expect(receivedMessage.GetSourceId()).To(Equal("src_id"))
	})

	It("should truncate long messages", func() {
		longMessage := strings.Repeat("7", MAX_MESSAGE_BYTE_SIZE*2)
		logMessage := testhelpers.NewLogMessage(longMessage, "test_app_id")

		emitter.EmitLogMessage(logMessage)

		receivedMessage := extractLogMessage(client.WriteArgsForCall(0))
		receivedMessageText := receivedMessage.GetMessage()

		truncatedOffset := len(receivedMessageText) - len(TRUNCATED_BYTES)
		expectedBytes := append([]byte(receivedMessageText)[:truncatedOffset], TRUNCATED_BYTES...)

		Expect(receivedMessageText).To(Equal(expectedBytes))
		Expect(receivedMessageText).To(HaveLen(MAX_MESSAGE_BYTE_SIZE))
	})

	It("should split messages on new lines", func() {
		message := "message1\n\rmessage2\nmessage3\r\nmessage4\r"
		logMessage := testhelpers.NewLogMessage(message, "test_app_id")

		emitter.EmitLogMessage(logMessage)
		Expect(client.WriteCallCount()).To(Equal(4))

		for i, expectedMessage := range []string{"message1", "message2", "message3", "message4"} {
			receivedMessage := extractLogMessage(client.WriteArgsForCall(i))
			Expect(receivedMessage.GetMessage()).To(Equal([]byte(expectedMessage)))
		}
	})

	It("should build the log envelope correctly", func() {
		emitter.Emit("appid", "foo")
		receivedEnvelope := extractLogEnvelope(client.WriteArgsForCall(0))

		Expect(receivedEnvelope.GetLogMessage().GetMessage()).To(Equal([]byte("foo")))
		Expect(receivedEnvelope.GetLogMessage().GetAppId()).To(Equal("appid"))
		Expect(receivedEnvelope.GetRoutingKey()).To(Equal("appid"))
		Expect(receivedEnvelope.GetLogMessage().GetSourceId()).To(Equal("42"))
	})

	It("should sign the log message correctly", func() {
		emitter.Emit("appid", "foo")
		receivedEnvelope := extractLogEnvelope(client.WriteArgsForCall(0))
		Expect(receivedEnvelope.VerifySignature("secret")).To(BeTrue(), "Expected envelope to be signed with the correct secret key")
	})

	It("source name is set if mapping is unknown", func() {
		emitter, err := NewEmitter("localhost:3456", "XYZ", "42", "secret", nil)
		Expect(err).ToNot(HaveOccurred())
		client := &fakeclient.FakeClient{}
		emitter.LoggregatorClient = client

		emitter.Emit("test_app_id", "test_msg")
		receivedMessage := extractLogMessage(client.WriteArgsForCall(0))

		Expect(receivedMessage.GetSourceName()).To(Equal("XYZ"))
	})

	Context("when missing an app id", func() {
		It("should not emit", func() {
			emitter.Emit("", "foo")
			Expect(client.WriteCallCount()).To(Equal(0))

			emitter.Emit("    ", "foo")
			Expect(client.WriteCallCount()).To(Equal(0))
		})
	})
})

func extractLogEnvelope(data []byte) *logmessage.LogEnvelope {
	receivedEnvelope := &logmessage.LogEnvelope{}

	err := proto.Unmarshal(data, receivedEnvelope)
	Expect(err).ToNot(HaveOccurred())

	return receivedEnvelope
}

func extractLogMessage(data []byte) *logmessage.LogMessage {
	envelope := extractLogEnvelope(data)

	return envelope.GetLogMessage()
}
