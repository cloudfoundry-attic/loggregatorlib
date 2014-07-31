package agentlistener

import (
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent/instrumentation"
	"net"
	"sync"
	"sync/atomic"
)

type AgentListener interface {
	instrumentation.Instrumentable
	Start()
	Stop()
}

type agentListener struct {
	*gosteno.Logger
	host                 string
	receivedMessageCount uint64
	receivedByteCount    uint64
	dataChannel          chan []byte
	connection           net.PacketConn
	contextName          string
	sync.RWMutex
}

func NewAgentListener(host string, givenLogger *gosteno.Logger, name string) (AgentListener, <-chan []byte) {
	byteChan := make(chan []byte, 1024)
	return &agentListener{Logger: givenLogger, host: host, dataChannel: byteChan, contextName: name}, byteChan
}

func (agentListener *agentListener) Start() {
	connection, err := net.ListenPacket("udp", agentListener.host)
	if err != nil {
		agentListener.Fatalf("Failed to listen on port. %s", err)
	}
	agentListener.Infof("Listening on port %s", agentListener.host)
	agentListener.Lock()
	agentListener.connection = connection
	agentListener.Unlock()

	readBuffer := make([]byte, 65535) //buffer with size = max theoretical UDP size
	defer close(agentListener.dataChannel)
	for {
		readCount, senderAddr, err := connection.ReadFrom(readBuffer)
		if err != nil {
			agentListener.Debugf("Error while reading. %s", err)
			return
		}
		agentListener.Debugf("AgentListener: Read %d bytes from address %s", readCount, senderAddr)

		readData := make([]byte, readCount) //pass on buffer in size only of read data
		copy(readData, readBuffer[:readCount])

		atomic.AddUint64(&agentListener.receivedMessageCount, 1)
		atomic.AddUint64(&agentListener.receivedByteCount, uint64(readCount))
		agentListener.dataChannel <- readData
	}

}

func (agentListener *agentListener) Stop() {
	agentListener.Lock()
	defer agentListener.Unlock()
	agentListener.connection.Close()
}

func (agentListener *agentListener) metrics() []instrumentation.Metric {
	return []instrumentation.Metric{
		instrumentation.Metric{Name: "currentBufferCount", Value: len(agentListener.dataChannel)},
		instrumentation.Metric{Name: "receivedMessageCount", Value: atomic.LoadUint64(&agentListener.receivedMessageCount)},
		instrumentation.Metric{Name: "receivedByteCount", Value: atomic.LoadUint64(&agentListener.receivedByteCount)},
	}
}

func (agentListener *agentListener) Emit() instrumentation.Context {
	return instrumentation.Context{Name: agentListener.contextName,
		Metrics: agentListener.metrics(),
	}
}
