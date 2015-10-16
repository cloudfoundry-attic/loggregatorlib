package servicediscovery_test

import (
	"github.com/cloudfoundry/gunk/workpool"
	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"
	"github.com/cloudfoundry/loggregatorlib/servicediscovery"
	"github.com/cloudfoundry/storeadapter"
	"github.com/cloudfoundry/storeadapter/etcdstoreadapter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceDiscovery", func() {
	var (
		storeAdapter storeadapter.StoreAdapter
		list         servicediscovery.ServerAddressList
		node         storeadapter.StoreNode
	)

	BeforeEach(func() {
		node = storeadapter.StoreNode{
			Key:   "/healthstatus/loggregator/z1/loggregator_z1",
			Value: []byte("10.0.0.1"),
		}

		workPool, err := workpool.NewWorkPool(10)
		Expect(err).NotTo(HaveOccurred())
		options := &etcdstoreadapter.ETCDOptions{
			ClusterUrls: etcdRunner.NodeURLS(),
		}
		storeAdapter, err = etcdstoreadapter.New(options, workPool)
		Expect(err).NotTo(HaveOccurred())

		err = storeAdapter.Connect()
		Expect(err).NotTo(HaveOccurred())

		list = servicediscovery.NewServerAddressList(storeAdapter, "/healthstatus/loggregator", loggertesthelper.Logger())
	})

	AfterEach(func() {
		list.Stop()
	})

	Context("when a services is created", func() {
		Context("before servicediscovery begins", func() {
			It("gets the state of the world at startup", func() {
				Expect(list.GetAddresses()).To(HaveLen(0))

				storeAdapter.Create(node)
				go list.Run()

				Eventually(list.GetAddresses).Should(ConsistOf(string(node.Value)))
			})
		})

		Context("after servicediscovery begins", func() {
			It("adds servers that appear later", func() {
				go list.Run()
				Consistently(list.GetAddresses, 1).Should(BeEmpty())

				storeAdapter.Create(node)

				Eventually(list.GetAddresses).Should(ConsistOf(string(node.Value)))
			})
		})
	})

	Context("when services exists", func() {
		BeforeEach(func() {
			err := storeAdapter.Create(node)
			Expect(err).NotTo(HaveOccurred())

			go list.Run()
			Eventually(list.GetAddresses).Should(HaveLen(1))
		})

		Context("when a node is removed", func() {
			It("removes the service", func() {
				err := storeAdapter.Delete(node.Key)
				Expect(err).NotTo(HaveOccurred())
				Eventually(list.GetAddresses).Should(BeEmpty())
			})
		})

		Context("when a node is updated", func() {
			It("the old value is replaced", func() {
				node.Value = []byte("10.0.0.2")
				err := storeAdapter.Update(node)
				Expect(err).NotTo(HaveOccurred())

				Eventually(list.GetAddresses).Should(ConsistOf(string(node.Value)))
			})
		})

		It("only finds nodes for the server type", func() {
			router := storeadapter.StoreNode{
				Key:   "/healthstatus/router/z1/router_z1",
				Value: []byte("10.99.99.99"),
			}
			storeAdapter.Create(router)

			Consistently(list.GetAddresses).Should(ConsistOf(string(node.Value)))
		})

		Context("when ETCD is not running", func() {
			JustBeforeEach(func() {
				etcdRunner.Stop()
			})

			AfterEach(func() {
				etcdRunner.Start()
			})

			It("continues to run if the store times out", func() {
				Consistently(list.GetAddresses).Should(HaveLen(1))
			})
		})
	})
})
