package loggregatorclient_test

import (
	"net"
	"strconv"

	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/loggregatorclient"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
)

var _ = Describe("loggregatorclient", func() {
	var (
		loggregatorClient  loggregatorclient.LoggregatorClient
		udpListener        *net.UDPConn
		loggregatorAddress string
	)

	BeforeEach(func() {
		port := 9875 + config.GinkgoConfig.ParallelNode
		loggregatorAddress = net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
		loggregatorClient = loggregatorclient.NewLoggregatorClient(loggregatorAddress, gosteno.NewLogger("TestLogger"), 0)

		udpAddr, _ := net.ResolveUDPAddr("udp", loggregatorAddress)
		udpListener, _ = net.ListenUDP("udp", udpAddr)
	})

	AfterEach(func() {
		loggregatorClient.Stop()
		udpListener.Close()
	})

	It("sends log messages to loggregator", func() {
		expectedOutput := []byte("Important Testmessage")

		loggregatorClient.Send(expectedOutput)

		buffer := make([]byte, 4096)
		readCount, _, _ := udpListener.ReadFromUDP(buffer)

		received := string(buffer[:readCount])
		Expect(received).To(Equal(string(expectedOutput)))

	})

	Context("Metrics", func() {
		var client loggregatorclient.LoggregatorClient
		BeforeEach(func() {
			client = loggregatorclient.NewLoggregatorClient(loggregatorAddress, gosteno.NewLogger("TestLogger"), 0)
			expectedOutput := []byte("Important Testmessage")

			client.Send(expectedOutput)

			buffer := make([]byte, 4096)
			udpListener.ReadFromUDP(buffer)

			client.Stop()
		})

		It("emits over varz", func() {
			metrics := client.Emit().Metrics
			Expect(metrics).To(HaveLen(5))

			for _, metric := range metrics {
				Expect(metric.Tags).To(HaveKeyWithValue("loggregatorAddress", "127.0.0.1"))

				switch metric.Name {
				case "currentBufferCount":
					Expect(metric.Value).To(Equal(uint64(0)))
				case "sentMessageCount":
					Expect(metric.Value).To(Equal(uint64(1)))
				case "sentByteCount":
					Expect(metric.Value).To(Equal(uint64(21)))
				case "receivedMessageCount":
					Expect(metric.Value).To(Equal(uint64(1)))
				case "receivedByteCount":
					Expect(metric.Value).To(Equal(uint64(21)))
				default:
					Fail("Got an invalid metric name: " + metric.Name)
				}
			}
		})
	})

	It("doesn't send empty data", func() {
		bufferSize := 4096
		firstMessage := []byte("")
		secondMessage := []byte("hi")

		loggregatorClient.Send(firstMessage)
		loggregatorClient.Send(secondMessage)

		buffer := make([]byte, bufferSize)
		readCount, _, _ := udpListener.ReadFromUDP(buffer)

		received := string(buffer[:readCount])
		Expect(received).To(Equal(string(secondMessage)))
	})
})
