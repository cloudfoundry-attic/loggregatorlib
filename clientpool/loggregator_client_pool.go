package clientpool

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/loggregatorclient"
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
	serverAddressGetterInZone   AddressGetter
	serverAddressGetterAllZones AddressGetter
}

func NewLoggregatorClientPool(logger *gosteno.Logger, port int, getterInZone AddressGetter, getterAllZones AddressGetter) *LoggregatorClientPool {
	return &LoggregatorClientPool{
		loggregatorPort: port,
		clients:         make(map[string]*loggregatorclient.LoggregatorClient),
		logger:          logger,
		serverAddressGetterInZone:   getterInZone,
		serverAddressGetterAllZones: getterAllZones,
	}
}

func (pool *LoggregatorClientPool) ListClients() []loggregatorclient.LoggregatorClient {
	pool.syncWithAddressList(pool.serverAddressGetterInZone.GetAddresses())

	if len(pool.clients) == 0 {
		pool.syncWithAddressList(pool.serverAddressGetterAllZones.GetAddresses())
	}

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
