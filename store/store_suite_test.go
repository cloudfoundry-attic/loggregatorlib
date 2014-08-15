package store_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/loggregatorlib/appservice"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent"
	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"
	"github.com/cloudfoundry/storeadapter"
	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	"github.com/onsi/ginkgo/config"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"testing"
	"time"
)

var etcdRunner *etcdstorerunner.ETCDClusterRunner

func TestStore(t *testing.T) {
	RegisterFailHandler(Fail)

	cmd := exec.Command("which", "etcd")
	error := cmd.Run()
	if error != nil {
		panic("etcd binary not found; required for StoreSuite")
	}

	SetDefaultEventuallyTimeout(5 * time.Second)
	cfcomponent.Logger = loggertesthelper.Logger()
	registerSignalHandler()

	etcdPort := 5800 + (config.GinkgoConfig.ParallelNode-1)*10
	etcdRunner = etcdstorerunner.NewETCDClusterRunner(etcdPort, 1)
	etcdRunner.Start()
	RunSpecs(t, "Store Suite")
	etcdRunner.Adapter().Disconnect()
	etcdRunner.Stop()
}

var _ = BeforeEach(func() {
	etcdRunner.Adapter().Disconnect()
	etcdRunner.Reset()
	etcdRunner.Adapter().Connect()
})

func registerSignalHandler() {
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill)

		select {
		case <-c:
			etcdRunner.Stop()
			os.Exit(0)
		}
	}()
}

func buildNode(appService appservice.AppService) storeadapter.StoreNode {
	return storeadapter.StoreNode{
		Key:   path.Join("/loggregator/services", appService.AppId, appService.Id()),
		Value: []byte(appService.Url),
	}
}

type FakeAdapter struct {
	DeleteCount     int
	WatchErrChannel chan error
	WatchCounter    int
	sync.Mutex
}

func (adapter *FakeAdapter) GetWatchCounter() int {
	adapter.Lock()
	defer adapter.Unlock()
	return adapter.WatchCounter
}

func (adapter *FakeAdapter) Connect() error                      { return nil }
func (adapter *FakeAdapter) Create(storeadapter.StoreNode) error { return nil }
func (adapter *FakeAdapter) Update(storeadapter.StoreNode) error { return nil }
func (adapter *FakeAdapter) CompareAndSwap(storeadapter.StoreNode, storeadapter.StoreNode) error {
	return nil
}
func (adapter *FakeAdapter) CompareAndSwapByIndex(uint64, storeadapter.StoreNode) error {
	return nil
}

func (adapter *FakeAdapter) SetMulti(nodes []storeadapter.StoreNode) error { return nil }
func (adapter *FakeAdapter) Get(key string) (storeadapter.StoreNode, error) {
	return storeadapter.StoreNode{}, nil
}
func (adapter *FakeAdapter) ListRecursively(key string) (storeadapter.StoreNode, error) {
	return storeadapter.StoreNode{}, nil
}
func (adapter *FakeAdapter) Delete(keys ...string) error {
	adapter.DeleteCount++
	return nil
}
func (adapter *FakeAdapter) CompareAndDelete(storeadapter.StoreNode) error { return nil }
func (adapter *FakeAdapter) UpdateDirTTL(key string, ttl uint64) error     { return nil }
func (adapter *FakeAdapter) Watch(key string) (events <-chan storeadapter.WatchEvent, stop chan<- bool, errors <-chan error) {
	adapter.Lock()
	defer adapter.Unlock()
	adapter.WatchCounter++
	adapter.WatchErrChannel = make(chan error, 1)

	return nil, make(chan bool), adapter.WatchErrChannel
}
func (adapter *FakeAdapter) Disconnect() error { return nil }
func (adapter *FakeAdapter) MaintainNode(storeNode storeadapter.StoreNode) (lostNode <-chan bool, releaseNode chan chan bool, err error) {
	return nil, nil, nil
}
