package agentlistener_test

import (
	"fmt"
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/agentlistener"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net"
)

var _ = Describe("AgentListener", func() {

	Context("with a listner running", func() {

		var listener agentlistener.AgentListener
		var dataChannel <-chan []byte

		BeforeEach(func() {
			listener, dataChannel = agentlistener.NewAgentListener("127.0.0.1:3456", gosteno.NewLogger("TestLogger"), "agentListener")
			go listener.Start()
		})

		AfterEach(func() {
			listener.Stop()
		})

		It("should listen to the socket", func(done Done) {

			expectedData := "Some Data"
			otherData := "More stuff"

			connection, err := net.Dial("udp", "localhost:3456")

			_, err = connection.Write([]byte(expectedData))
			Expect(err).To(BeNil())

			received := <-dataChannel
			Expect(string(received)).To(Equal(expectedData))

			_, err = connection.Write([]byte(otherData))
			Expect(err).To(BeNil())

			receivedAgain := <-dataChannel
			Expect(string(receivedAgain)).To(Equal(otherData))

			metrics := listener.Emit().Metrics
			Expect(metrics).To(HaveLen(3)) //make sure all expected metrics are present
			for _, metric := range metrics {
				switch metric.Name {
				case "currentBufferCount":
					Expect(metric.Value).To(Equal(0))
				case "receivedMessageCount":
					Expect(metric.Value).To(Equal(uint64(2)))
				case "receivedByteCount":
					Expect(metric.Value).To(Equal(uint64(19)))
				default:
					Fail(fmt.Sprintf("Got an invalid metric name: %s", metric.Name))
				}
			}
			close(done)
		}, 2)
	})

	Context("Emit", func() {
		It("uses the given name for the context", func() {
			listener, _ := agentlistener.NewAgentListener("127.0.0.1:3456", gosteno.NewLogger("TestLogger"), "secretAgentOrange")
			context := listener.Emit()

			Expect(context.Name).To(Equal("secretAgentOrange"))
		})
	})
})
