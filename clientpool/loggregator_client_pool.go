package clientpool

import (
	"errors"
	"fmt"
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/loggregatorclient"
	"github.com/cloudfoundry/storeadapter"
	"math/rand"
	"sync"
	"time"
	"github.com/cloudfoundry/loggregatorlib/servicediscovery"
)

var ErrorEmptyClientPool = errors.New("loggregator client pool is empty")

type LoggregatorClientPool struct {
	clients         map[string]*loggregatorclient.LoggregatorClient
	logger          *gosteno.Logger
	loggregatorPort int
	sync.RWMutex
	serverAddressList servicediscovery.ServerAddressList
}

func NewLoggregatorClientPool(logger *gosteno.Logger, port int) *LoggregatorClientPool {
	return &LoggregatorClientPool{
		loggregatorPort: port,
		clients:         make(map[string]*loggregatorclient.LoggregatorClient),
		logger:          logger,
	}
}

func (pool *LoggregatorClientPool) RandomClient() (loggregatorclient.LoggregatorClient, error) {
	list := pool.ListClients()
	if len(list) == 0 {
		return nil, ErrorEmptyClientPool
	}

	return list[rand.Intn(len(list))], nil
}

func (pool *LoggregatorClientPool) RunUpdateLoop(storeAdapter storeadapter.StoreAdapter, key string, stopChan <-chan struct{}, interval time.Duration) {
	pool.serverAddressList = servicediscovery.NewServerAddressList(storeAdapter, key, pool.logger)
	go pool.serverAddressList.Run(interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pool.syncWithAddressList(pool.serverAddressList.GetAddresses())
		case <-stopChan:
			pool.serverAddressList.Stop()
			return
		}
	}
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

func (pool *LoggregatorClientPool) ListClients() []loggregatorclient.LoggregatorClient {
	pool.RLock()
	defer pool.RUnlock()

	val := make([]loggregatorclient.LoggregatorClient, 0, len(pool.clients))
	for _, client := range pool.clients {
		val = append(val, *client)
	}

	return val
}

func (pool *LoggregatorClientPool) ListAddresses() []string {
	pool.RLock()
	defer pool.RUnlock()

	val := make([]string, 0, len(pool.clients))
	for addr := range pool.clients {
		val = append(val, addr)
	}

	return val
}

func (pool *LoggregatorClientPool) Add(address string, client loggregatorclient.LoggregatorClient) {
	pool.Lock()
	defer pool.Unlock()

	pool.clients[address] = &client
}

func (pool *LoggregatorClientPool) hasServerFor(addr string) bool {
	_, ok := pool.clients[addr]
	return ok
}

func leafNodes(root storeadapter.StoreNode) []storeadapter.StoreNode {
	if !root.Dir {
		return []storeadapter.StoreNode{root}
	}

	leaves := []storeadapter.StoreNode{}
	for _, node := range root.ChildNodes {
		leaves = append(leaves, leafNodes(node)...)
	}
	return leaves
}
