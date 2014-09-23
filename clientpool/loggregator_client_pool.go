package clientpool

import (
	"errors"
	"fmt"
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/loggregatorclient"
	"math/rand"
	"sync"
)

var ErrorEmptyClientPool = errors.New("loggregator client pool is empty")

type AddressGetter interface {
	GetAddresses() []string
}

type LoggregatorClientPool struct {
	clients         map[string]*loggregatorclient.LoggregatorClient
	logger          *gosteno.Logger
	loggregatorPort int
	sync.RWMutex
	serverAddressGetter AddressGetter
}

func NewLoggregatorClientPool(logger *gosteno.Logger, port int, getter AddressGetter) *LoggregatorClientPool {
	return &LoggregatorClientPool{
		loggregatorPort:     port,
		clients:             make(map[string]*loggregatorclient.LoggregatorClient),
		logger:              logger,
		serverAddressGetter: getter,
	}
}

func (pool *LoggregatorClientPool) ListClients() []loggregatorclient.LoggregatorClient {
	pool.syncWithAddressList(pool.serverAddressGetter.GetAddresses())

	pool.RLock()
	defer pool.RUnlock()

	val := make([]loggregatorclient.LoggregatorClient, 0, len(pool.clients))
	for _, client := range pool.clients {
		val = append(val, *client)
	}

	return val
}

func (pool *LoggregatorClientPool) RandomClient() (loggregatorclient.LoggregatorClient, error) {
	list := pool.ListClients()
	if len(list) == 0 {
		return nil, ErrorEmptyClientPool
	}

	return list[rand.Intn(len(list))], nil
}

func (pool *LoggregatorClientPool) syncWithAddressList(addresses []string) {
	pool.Lock()
	defer pool.Unlock()

	newClients := make(map[string]*loggregatorclient.LoggregatorClient, len(addresses))

	for _, address := range addresses {
		clientIdentifier := fmt.Sprintf("%s:%d", address, pool.loggregatorPort)

		if pool.hasServerFor(clientIdentifier) {
			newClients[clientIdentifier] = pool.clients[clientIdentifier]
		} else {
			client := loggregatorclient.NewLoggregatorClient(clientIdentifier, pool.logger, loggregatorclient.DefaultBufferSize)
			newClients[clientIdentifier] = &client
		}
	}
	pool.clients = newClients
}

func (pool *LoggregatorClientPool) hasServerFor(addr string) bool {
	_, ok := pool.clients[addr]
	return ok
}
