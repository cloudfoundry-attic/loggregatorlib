package store_test

import (
	"github.com/cloudfoundry/loggregatorlib/appservice"
	. "github.com/cloudfoundry/loggregatorlib/store"
	"github.com/cloudfoundry/loggregatorlib/store/cache"
	"github.com/cloudfoundry/storeadapter"
	"github.com/cloudfoundry/storeadapter/fakestoreadapter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"path"
)

var _ = Describe("AppServiceUnit", func() {
	Context("when the store has data", func() {
		var store *AppServiceStore
		var adapter *fakestoreadapter.FakeStoreAdapter
		var incomingChan chan appservice.AppServices
		var app1Service1 appservice.AppService

		BeforeEach(func() {
			adapter = fakestoreadapter.New()
			c := cache.NewAppServiceCache()
			incomingChan = make(chan appservice.AppServices)
			app1Service1 = appservice.AppService{AppId: "app-1", Url: "syslog://example.com:12345"}
			store = NewAppServiceStore(adapter, c)

			go store.Run(incomingChan)
		})

		It("does not modify the store, when deleting data that doesn't exist", func() {
			key := path.Join("/loggregator/services", app1Service1.AppId, app1Service1.Id())

			incomingChan <- appservice.AppServices{
				AppId: app1Service1.AppId,
				Urls:  []string{app1Service1.Url},
			}

			incomingChan <- appservice.AppServices{
				AppId: app1Service1.AppId,
				Urls:  []string{},
			}

			incomingChan <- appservice.AppServices{
				AppId: app1Service1.AppId,
				Urls:  []string{},
			}

			_, err := adapter.Get(key)
			Expect(err).To(Equal(storeadapter.ErrorKeyNotFound))
		})
	})
})
