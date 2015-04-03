package store_test

import (
	. "github.com/cloudfoundry/loggregatorlib/store"
	"github.com/cloudfoundry/loggregatorlib/store/cache"
	"github.com/cloudfoundry/storeadapter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"path"
	"time"
	"sync"

	"github.com/cloudfoundry/loggregatorlib/appservice"
)

const (
	APP1_ID = "app-1"
	APP2_ID = "app-2"
	APP3_ID = "app-3"
)

var _ = Describe("AppServiceStoreWatcher", func() {
	var watcher *AppServiceStoreWatcher
	var watcherRunComplete sync.WaitGroup

	var runWatcher func()

	var adapter storeadapter.StoreAdapter
	var outAddChan <-chan appservice.AppService
	var outRemoveChan <-chan appservice.AppService

	var app1Service1 appservice.AppService
	var app1Service2 appservice.AppService
	var app2Service1 appservice.AppService

	drainOutgoingChannel := func(c <-chan appservice.AppService, count int) []appservice.AppService {
		appServices := []appservice.AppService{}
		for i := 0; i < count; i++ {
			var appService appservice.AppService
			Eventually(c).Should(Receive(&appService))
			appServices = append(appServices, appService)
		}

		return appServices
	}

	BeforeEach(func() {
		app1Service1 = appservice.AppService{AppId: APP1_ID, Url: "syslog://example.com:12345"}
		app1Service2 = appservice.AppService{AppId: APP1_ID, Url: "syslog://example.com:12346"}
		app2Service1 = appservice.AppService{AppId: APP2_ID, Url: "syslog://example.com:12345"}

		adapter = etcdRunner.Adapter()

		c := cache.NewAppServiceCache()
		watcher, outAddChan, outRemoveChan = NewAppServiceStoreWatcher(adapter, c)
		watcherRunComplete = sync.WaitGroup{}

		runWatcher = func() {
			watcherRunComplete.Add(1)
			go func() {
				watcher.Run()
				watcherRunComplete.Done()
			}()
		}
	})

	AfterEach(func(done Done) {
		Expect(adapter.Disconnect()).To(Succeed())
		watcherRunComplete.Wait()
		close(done)
	})

	Describe("Shutdown", func() {
		It("should close the outgoing channels", func() {
			runWatcher()

			time.Sleep(500 * time.Millisecond)
			adapter.Disconnect()

			Eventually(outRemoveChan).Should(BeClosed())
			Eventually(outAddChan).Should(BeClosed())
		})
	})

	Describe("Loading watcher state on startup", func() {
		Context("when the store is empty", func() {
			It("should not send anything on the output channels", func() {
				runWatcher()

				Consistently(outAddChan).Should(BeEmpty())
				Consistently(outRemoveChan).Should(BeEmpty())
			})
		})

		Context("when the store has AppServices in it", func() {
			BeforeEach(func() {
				adapter.Create(buildNode(app1Service1))
				adapter.Create(buildNode(app1Service2))
				adapter.Create(buildNode(app2Service1))
			})

			It("should send all the AppServices on the output add channel", func(done Done) {
				runWatcher()

				appServices := drainOutgoingChannel(outAddChan, 3)

				Expect(appServices).To(ContainElement(app1Service1))
				Expect(appServices).To(ContainElement(app1Service2))
				Expect(appServices).To(ContainElement(app2Service1))

				Expect(outRemoveChan).To(BeEmpty())

				close(done)
			})
		})
	})

	Describe("when the store has data and watcher is bootstrapped", func() {
		BeforeEach(func(done Done) {
			adapter.Create(buildNode(app1Service1))
			adapter.Create(buildNode(app1Service2))
			adapter.Create(buildNode(app2Service1))

			runWatcher()
			drainOutgoingChannel(outAddChan, 3)
			close(done)
		}, 5)

		It("does not send updates when the data has already been processed", func(done Done) {
			adapter.Create(buildNode(app1Service1))
			adapter.Create(buildNode(app1Service2))

			Expect(outAddChan).To(BeEmpty())
			Expect(outRemoveChan).To(BeEmpty())

			close(done)
		})

		Context("when there is new data in the store", func() {
			Context("when an existing app has a new service through a create operation", func() {
				It("adds that service to the outgoing add channel", func() {
					app2Service2 := appservice.AppService{AppId: APP2_ID, Url: "syslog://new.example.com:12345"}
					adapter.Get(key(app2Service2))

					adapter.Create(buildNode(app2Service2))

					var appService appservice.AppService
					Eventually(outAddChan).Should(Receive(&appService))
					Expect(appService).To(Equal(app2Service2))

					Expect(outRemoveChan).To(BeEmpty())
				})
			})

			Context("When an existing app gets a new service through an update operation", func() {
				It("adds that service to the outgoing add channel", func() {
					app2Service2 := appservice.AppService{AppId: APP2_ID, Url: "syslog://new.example.com:12345"}
					adapter.Get(key(app2Service2))

					err := adapter.SetMulti([]storeadapter.StoreNode{buildNode(app2Service2)})
					Expect(err).ToNot(HaveOccurred())

					var appService appservice.AppService

					Eventually(outAddChan).Should(Receive(&appService))
					Expect(appService).To(Equal(app2Service2))

					Expect(outRemoveChan).To(BeEmpty())
				})
			})

			Context("when a new app appears", func() {
				It("adds that app and its services to the outgoing add channel", func(done Done) {
					app3Service1 := appservice.AppService{AppId: APP3_ID, Url: "syslog://app3.example.com:12345"}
					app3Service2 := appservice.AppService{AppId: APP3_ID, Url: "syslog://app3.example.com:12346"}

					adapter.Create(buildNode(app3Service1))
					adapter.Create(buildNode(app3Service2))

					appServices := drainOutgoingChannel(outAddChan, 2)

					Expect(appServices).To(ConsistOf(app3Service1, app3Service2))

					Expect(outRemoveChan).To(BeEmpty())

					close(done)
				})
			})
		})

		Context("When an existing service is updated", func() {
			It("should not notify the channel again", func() {
				adapter.SetMulti([]storeadapter.StoreNode{buildNode(app2Service1)})
				Expect(outAddChan).To(BeEmpty())
				Expect(outRemoveChan).To(BeEmpty())
			})
		})

		Context("when a service or app should be removed", func() {
			Context("when an existing app loses one of its services", func() {
				It("sends that service on the output remove channel", func() {
					adapter.Delete(key(app1Service2))

					var appService appservice.AppService
					Eventually(outRemoveChan).Should(Receive(&appService))
					Expect(appService).To(Equal(app1Service2))

					Expect(outAddChan).To(BeEmpty())
				})
			})

			Context("when an existing app loses all of its services", func() {
				It("sends all of the app services on the outgoing remove channel", func() {
					adapter.Get(path.Join("/loggregator/services", APP1_ID))
					adapter.Delete(path.Join("/loggregator/services", APP1_ID))
					appServices := drainOutgoingChannel(outRemoveChan, 2)
					Expect(appServices).To(ConsistOf(app1Service1, app1Service2))
					Expect(outAddChan).To(BeEmpty())

					adapter.Create(buildNode(app1Service1))
					adapter.Create(buildNode(app1Service2))
					appServices = drainOutgoingChannel(outAddChan, 2)
					Expect(appServices).To(ConsistOf(app1Service1, app1Service2))
					Expect(outRemoveChan).To(BeEmpty())
				})
			})
		})

		Describe("with multiple updates to the same app-id", func() {
			It("should perform the updates correctly on the outgoing channels", func() {
				adapter.Get(key(app1Service2))
				adapter.Delete(key(app1Service2))

				var appService appservice.AppService
				Eventually(outRemoveChan).Should(Receive(&appService))
				Expect(appService).To(Equal(app1Service2))
				Expect(outAddChan).To(BeEmpty())

				adapter.Create(buildNode(app1Service2))

				Eventually(outAddChan).Should(Receive(&appService))
				Expect(appService).To(Equal(app1Service2))
				Expect(outRemoveChan).To(BeEmpty())

				adapter.Delete(key(app1Service1))

				Expect(outAddChan).To(BeEmpty())
				Eventually(outRemoveChan).Should(Receive(&appService))
				Expect(appService).To(Equal(app1Service1))
			})
		})

		Context("when an existing app service expires", func() {
			It("removes the app service from the cache", func() {
				app2Service2 := appservice.AppService{AppId: APP2_ID, Url: "syslog://foo/a"}
				adapter.Get(key(app2Service2))
				adapter.Create(buildNode(app2Service2))
				var appService appservice.AppService

				Eventually(outAddChan).Should(Receive(&appService))
				Expect(appService).To(Equal(app2Service2))

				adapter.UpdateDirTTL("/loggregator/services/app-2", 1)
				time.Sleep(1 * time.Second)
				Eventually(outAddChan).Should(BeEmpty())
				Expect(outAddChan).To(BeEmpty())
				appServices := drainOutgoingChannel(outRemoveChan, 2)

				Expect(appServices).To(ConsistOf(app2Service1, app2Service2))

				adapter.Create(buildNode(app2Service2))

				Eventually(outAddChan).Should(Receive(&appService))
				Expect(appService).To(Equal(app2Service2))
			})
		})
	})
})

func key(service appservice.AppService) string {
	return path.Join("/loggregator/services", service.AppId, service.Id())
}
