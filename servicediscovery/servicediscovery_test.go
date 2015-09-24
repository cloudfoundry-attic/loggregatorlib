package servicediscovery_test

import (
	"time"

	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"
	"github.com/cloudfoundry/loggregatorlib/servicediscovery"
	"github.com/cloudfoundry/storeadapter"
	"github.com/cloudfoundry/storeadapter/fakestoreadapter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceDiscovery", func() {
	var (
		storeAdapter *fakestoreadapter.FakeStoreAdapter
		list         servicediscovery.ServerAddressList
	)

	BeforeEach(func() {
		storeAdapter = fakestoreadapter.New()
		list = servicediscovery.NewServerAddressList(storeAdapter, "/healthstatus/loggregator", loggertesthelper.Logger())
	})

	AfterEach(func() {
		list.Stop()
	})

	It("gets the state of the world at startup", func() {
		node := storeadapter.StoreNode{
			Key:   "/healthstatus/loggregator/z1/loggregator_z1",
			Value: []byte("10.0.0.1"),
		}
		storeAdapter.Create(node)

		go list.Run(2 * time.Second)

		expectedAddresses := []string{"10.0.0.1"}

		Eventually(list.GetAddresses, 1).Should(ConsistOf(expectedAddresses))
	})

	It("adds servers that appear later", func() {
		go list.Run(2 * time.Second)

		Consistently(list.GetAddresses, 1).Should(BeEmpty())

		node := storeadapter.StoreNode{
			Key:   "/healthstatus/loggregator/z1/loggregator_z1",
			Value: []byte("10.0.0.1"),
		}
		storeAdapter.Create(node)

		expectedAddresses := []string{"10.0.0.1"}

		Eventually(list.GetAddresses, 2).Should(ConsistOf(expectedAddresses))
	})

	It("removes servers that disappear later", func() {
		node := storeadapter.StoreNode{
			Key:   "/healthstatus/loggregator/z1/loggregator_z1",
			Value: []byte("10.0.0.1"),
		}

		storeAdapter.Create(node)

		list := servicediscovery.NewServerAddressList(storeAdapter, "/healthstatus/loggregator", loggertesthelper.Logger())

		go list.Run(1 * time.Millisecond)

		storeAdapter.Delete("/healthstatus/loggregator/z1/loggregator_z1")

		Eventually(list.GetAddresses).Should(BeEmpty())
	})

	It("only finds nodes for the server type", func() {
		node := storeadapter.StoreNode{
			Key:   "/healthstatus/loggregator/z1/loggregator_z1",
			Value: []byte("10.0.0.1"),
		}

		storeAdapter.Create(node)

		node = storeadapter.StoreNode{
			Key:   "/healthstatus/router/z1/router_z1",
			Value: []byte("10.99.99.99"),
		}

		storeAdapter.Create(node)

		go list.Run(1 * time.Millisecond)

		expectedAddresses := []string{"10.0.0.1"}

		Eventually(list.GetAddresses).Should(ConsistOf(expectedAddresses))
	})

	It("only returns one copy of each server", func() {
		node := storeadapter.StoreNode{
			Key:   "/healthstatus/loggregator/z1/loggregator_z1",
			Value: []byte("10.0.0.1"),
		}

		storeAdapter.Create(node)

		go list.Run(1 * time.Millisecond)

		expectedAddresses := []string{"10.0.0.1"}

		Eventually(list.GetAddresses).Should(ConsistOf(expectedAddresses))
		Consistently(list.GetAddresses).Should(HaveLen(1))
	})

	It("continues to run if the key is not found", func() {
		node := storeadapter.StoreNode{
			Key:   "/healthstatus/loggregator/z1/loggregator_z1",
			Value: []byte("10.0.0.1"),
		}

		storeAdapter.Create(node)

		go list.Run(1 * time.Millisecond)

		Eventually(list.GetAddresses).Should(HaveLen(1))

		storeAdapter.Lock()
		storeAdapter.ListErrInjector = fakestoreadapter.NewFakeStoreAdapterErrorInjector("", storeadapter.ErrorKeyNotFound)
		storeAdapter.Unlock()

		Consistently(list.GetAddresses).Should(HaveLen(1))
	})

	It("continues to run if the store times out", func() {
		node := storeadapter.StoreNode{
			Key:   "/healthstatus/loggregator/z1/loggregator_z1",
			Value: []byte("10.0.0.1"),
		}

		storeAdapter.Create(node)

		go list.Run(1 * time.Millisecond)

		Eventually(list.GetAddresses).Should(HaveLen(1))

		storeAdapter.Lock()
		storeAdapter.ListErrInjector = fakestoreadapter.NewFakeStoreAdapterErrorInjector("", storeadapter.ErrorTimeout)
		storeAdapter.Unlock()

		Consistently(list.GetAddresses).Should(HaveLen(1))
	})

	It("excludes nodes with no value", func() {
		node := storeadapter.StoreNode{
			Key:   "/healthstatus/loggregator/z1/loggregator_z1",
			Value: []byte{},
		}

		storeAdapter.Create(node)

		go list.Run(1 * time.Millisecond)

		Consistently(list.GetAddresses).Should(BeEmpty())
	})
})
