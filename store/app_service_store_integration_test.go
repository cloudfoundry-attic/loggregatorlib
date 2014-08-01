package store_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/cloudfoundry/loggregatorlib/appservice"
	. "github.com/cloudfoundry/loggregatorlib/store"
	"github.com/cloudfoundry/loggregatorlib/store/cache"
)

var _ = Describe("AppServiceStoreIntegration", func() {
	var (
		incomingChan  chan appservice.AppServices
		outAddChan    <-chan appservice.AppService
		outRemoveChan <-chan appservice.AppService
	)

	BeforeEach(func() {
		adapter := etcdRunner.Adapter()

		incomingChan = make(chan appservice.AppServices)
		c := cache.NewAppServiceCache()
		var watcher *AppServiceStoreWatcher
		watcher, outAddChan, outRemoveChan = NewAppServiceStoreWatcher(adapter, c)
		go watcher.Run()

		store := NewAppServiceStore(adapter, watcher)
		go store.Run(incomingChan)
	})

	It("should receive, store, and republish AppServices", func(done Done) {
		appServices := appservice.AppServices{AppId: "12345", Urls: []string{"syslog://foo"}}
		incomingChan <- appServices

		Expect(<-outAddChan).To(Equal(appservice.AppService{
			AppId: "12345", Url: "syslog://foo",
		}))

		appServices = appservice.AppServices{AppId: "12345", Urls: []string{"syslog://foo", "syslog://bar"}}
		incomingChan <- appServices

		Expect(<-outAddChan).To(Equal(appservice.AppService{
			AppId: "12345", Url: "syslog://bar",
		}))

		appServices = appservice.AppServices{AppId: "12345", Urls: []string{"syslog://bar"}}

		incomingChan <- appServices

		Expect(<-outRemoveChan).To(Equal(appservice.AppService{
			AppId: "12345", Url: "syslog://foo",
		}))
		close(done)
	}, 5)
})
