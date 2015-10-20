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

var _ = Describe("Udp Client", func() {
	var (
		client             loggregatorclient.Client
		udpListener        *net.UDPConn
		loggregatorAddress string
	)

	BeforeEach(func() {
		port := 9875 + config.GinkgoConfig.ParallelNode
		loggregatorAddress = net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
		var err error
		client, err = loggregatorclient.NewUDPClient(gosteno.NewLogger("TestLogger"), loggregatorAddress, 0)
		Expect(err).NotTo(HaveOccurred())

		udpAddr, _ := net.ResolveUDPAddr("udp", loggregatorAddress)
		udpListener, _ = net.ListenUDP("udp", udpAddr)
	})

	AfterEach(func() {
		client.Stop()
		udpListener.Close()
	})

	It("sends log messages to loggregator", func() {
		expectedOutput := []byte("Important Testmessage")

		client.Send(expectedOutput)

		buffer := make([]byte, 4096)
		readCount, _, _ := udpListener.ReadFromUDP(buffer)

		received := string(buffer[:readCount])
		Expect(received).To(Equal(string(expectedOutput)))

	})

	It("doesn't send empty data", func() {
		bufferSize := 4096
		firstMessage := []byte("")
		secondMessage := []byte("hi")

		client.Send(firstMessage)
		client.Send(secondMessage)

		buffer := make([]byte, bufferSize)
		readCount, _, _ := udpListener.ReadFromUDP(buffer)

		received := string(buffer[:readCount])
		Expect(received).To(Equal(string(secondMessage)))
	})
})
