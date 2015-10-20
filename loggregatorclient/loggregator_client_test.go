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

		udpAddr, err := net.ResolveUDPAddr("udp", loggregatorAddress)
		Expect(err).NotTo(HaveOccurred())

		udpListener, err = net.ListenUDP("udp", udpAddr)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		loggregatorClient.Stop()
		udpListener.Close()
	})

	It("sends log messages to loggregator", func() {
		expectedOutput := []byte("Important Testmessage")

		loggregatorClient.Send(expectedOutput)

		buffer := make([]byte, 4096)
		readCount, _, err := udpListener.ReadFromUDP(buffer)
		Expect(err).NotTo(HaveOccurred())

		received := string(buffer[:readCount])
		Expect(received).To(Equal(string(expectedOutput)))
	})

	It("doesn't send empty data", func() {
		bufferSize := 4096
		firstMessage := []byte("")
		secondMessage := []byte("hi")

		loggregatorClient.Send(firstMessage)
		loggregatorClient.Send(secondMessage)

		buffer := make([]byte, bufferSize)
		readCount, _, err := udpListener.ReadFromUDP(buffer)
		Expect(err).NotTo(HaveOccurred())

		received := string(buffer[:readCount])
		Expect(received).To(Equal(string(secondMessage)))
	})
})
