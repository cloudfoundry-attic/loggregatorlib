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
	println(error == nil)
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
