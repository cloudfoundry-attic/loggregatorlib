package loggregatorclient_test

import (
	"net"

	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/loggregatorclient"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UDP Client", func() {
	var (
		client      loggregatorclient.Client
		udpListener *net.UDPConn
	)

	BeforeEach(func() {
		udpAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
		udpListener, _ = net.ListenUDP("udp", udpAddr)

		var err error
		client, err = loggregatorclient.NewUDPClient(gosteno.NewLogger("TestLogger"), udpListener.LocalAddr().String(), 0)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		client.Close()
		udpListener.Close()
	})

	Describe("NewUDPClient", func() {
		Context("when the address is invalid", func() {
			It("returns an error", func() {
				_, err := loggregatorclient.NewUDPClient(gosteno.NewLogger("TestLogger"), "127.0.0.1:abc", 0)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("udpClient", func() {
		Describe("Scheme", func() {
			It("returns tls", func() {
				Expect(client.Scheme()).To(Equal("udp"))
			})
		})

		Describe("Address", func() {
			It("returns the address", func() {
				Expect(client.Address()).To(Equal(udpListener.LocalAddr().String()))
			})
		})
	})

	It("sends log messages to loggregator", func() {
		expectedOutput := []byte("Important Testmessage")

		_, err := client.Write(expectedOutput)
		Expect(err).NotTo(HaveOccurred())

		buffer := make([]byte, 4096)
		readCount, _, _ := udpListener.ReadFromUDP(buffer)

		received := string(buffer[:readCount])
		Expect(received).To(Equal(string(expectedOutput)))

	})

	It("doesn't send empty data", func() {
		bufferSize := 4096
		firstMessage := []byte("")
		secondMessage := []byte("hi")

		_, err := client.Write(firstMessage)
		Expect(err).NotTo(HaveOccurred())
		_, err = client.Write(secondMessage)
		Expect(err).NotTo(HaveOccurred())

		buffer := make([]byte, bufferSize)
		readCount, _, _ := udpListener.ReadFromUDP(buffer)

		received := string(buffer[:readCount])
		Expect(received).To(Equal(string(secondMessage)))
	})

	Describe("Close", func() {
		It("can be called multiple times", func() {
			done := make(chan struct{})
			go func() {
				client.Close()
				client.Close()
				close(done)
			}()
			Eventually(done).Should(BeClosed())
		})
	})
})
