package servicediscovery

import (
	"sync"

	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/storeadapter"
)

type ServerAddressList interface {
	Run()
	Stop()
	GetAddresses() []string
	DiscoverAddresses()
}

type serverAddressList struct {
	addressMap     map[string]struct{}
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
		addressMap:     make(map[string]struct{}),
		stopChan:       make(chan struct{}),
		storeKeyPrefix: storeKeyPrefix,
		logger:         logger,
	}
}

func (list *serverAddressList) Run() {
	events, stopWatch, errors := list.storeAdapter.Watch(list.storeKeyPrefix)
	list.DiscoverAddresses()

	for {
		select {
		case <-list.stopChan:
			close(stopWatch)
			return
		case event := <-events:
			list.handleEvent(&event)
		case err := <-errors:
			list.logger.Errord(map[string]interface{}{
				"error": err,
			},
				"ServerAddressList.Run: Watch failed")
			events, stopWatch, errors = list.storeAdapter.Watch(list.storeKeyPrefix)
			list.DiscoverAddresses()
		}
	}
}

func (list *serverAddressList) handleEvent(event *storeadapter.WatchEvent) {
	var value string
	if event.Node != nil {
		value = string(event.Node.Value)
	}
	list.Lock()
	switch event.Type {
	case storeadapter.CreateEvent:
		list.addressMap[value] = struct{}{}
	case storeadapter.DeleteEvent:
		fallthrough
	case storeadapter.ExpireEvent:
		prevValue := string(event.PrevNode.Value)
		delete(list.addressMap, prevValue)
	case storeadapter.UpdateEvent:
		prevValue := string(event.PrevNode.Value)
		if value != prevValue {
			delete(list.addressMap, prevValue)
			list.addressMap[value] = struct{}{}
		}
	}
	list.addresses = keys(list.addressMap)
	list.Unlock()
}

func keys(serviceMap map[string]struct{}) []string {
	a := make([]string, 0, len(serviceMap))
	for k, _ := range serviceMap {
		a = append(a, k)
	}
	return a
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

	addressMap := make(map[string]struct{})

	for _, leaf := range leaves {
		addressMap[string(leaf.Value)] = struct{}{}
	}

	addresses := keys(addressMap)

	list.Lock()
	list.addressMap = addressMap
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
