package servicediscovery

import (
	"sync"
	"time"

	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/storeadapter"
)

type ServerAddressList interface {
	Run(updateInterval time.Duration)
	Stop()
	GetAddresses() []string
	DiscoverAddresses()
}

type serverAddressList struct {
	addresses      []string
	storeAdapter   storeadapter.StoreAdapter
	stopChan       chan struct{}
	storeKeyPrefix string
	logger         *gosteno.Logger
	sync.RWMutex
}

func NewServerAddressList(storeAdapter storeadapter.StoreAdapter, storeKeyPrefix string, logger *gosteno.Logger) ServerAddressList {
	return &serverAddressList{
		storeAdapter:   storeAdapter,
		addresses:      []string{},
		stopChan:       make(chan struct{}),
		storeKeyPrefix: storeKeyPrefix,
		logger:         logger,
	}
}

func (list *serverAddressList) Run(updateInterval time.Duration) {
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-list.stopChan:
			return
		case <-ticker.C:
			list.DiscoverAddresses()
		}
	}
}

func (list *serverAddressList) DiscoverAddresses() {
	node, err := list.storeAdapter.ListRecursively(list.storeKeyPrefix)

	if err == storeadapter.ErrorKeyNotFound {
		list.logger.Debugf("ServerAddressList.Run: Unable to recursively find keys with prefix %s", list.storeKeyPrefix)
		return
	}

	if err == storeadapter.ErrorTimeout {
		list.logger.Debug("ServerAddressList.Run: Timed out talking to store; will try again soon.")
		return
	}

	if err != nil {
		panic(err) //FIXME: understand error modes and recovery cases better
	}

	leaves := leafNodes(node)

	addresses := []string{}

	for _, leaf := range leaves {
		addresses = append(addresses, string(leaf.Value))
	}

	list.Lock()
	list.addresses = addresses
	list.Unlock()
}

func (list *serverAddressList) Stop() {
	close(list.stopChan)
}

func (list *serverAddressList) GetAddresses() []string {
	list.RLock()
	defer list.RUnlock()

	return list.addresses
}

func leafNodes(root storeadapter.StoreNode) []storeadapter.StoreNode {
	if !root.Dir {
		if len(root.Value) == 0 {
			return []storeadapter.StoreNode{}
		} else {
			return []storeadapter.StoreNode{root}
		}
	}

	leaves := []storeadapter.StoreNode{}
	for _, node := range root.ChildNodes {
		leaves = append(leaves, leafNodes(node)...)
	}
	return leaves
}
