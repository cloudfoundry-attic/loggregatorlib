package store_test

import (
	"github.com/cloudfoundry/loggregatorlib/appservice"
	. "github.com/cloudfoundry/loggregatorlib/store"
	"github.com/cloudfoundry/loggregatorlib/store/cache"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppServiceStoreIntegration", func() {
	var (
		incomingChan  chan appservice.AppServices
		outAddChan    <-chan appservice.AppService
		outRemoveChan <-chan appservice.AppService
		stopChan chan struct{}
	)

	BeforeEach(func() {
		adapter := etcdRunner.Adapter()

		incomingChan = make(chan appservice.AppServices)
		c := cache.NewAppServiceCache()
		var watcher *AppServiceStoreWatcher
		watcher, outAddChan, outRemoveChan = NewAppServiceStoreWatcher(adapter, c)
		stopChan = make(chan struct{})
		go watcher.Run(stopChan)

		store := NewAppServiceStore(adapter, watcher)
		go store.Run(incomingChan)
	})

	AfterEach(func() {
		close(stopChan)
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
