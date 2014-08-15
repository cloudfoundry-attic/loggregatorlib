package store_test

import (
	"errors"
	. "github.com/cloudfoundry/loggregatorlib/store"
	"github.com/cloudfoundry/loggregatorlib/store/cache"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppServiceStoreWatcherUnit", func() {
	Context("when there is an error", func() {
		var adapter *FakeAdapter
		var stopChan chan struct{}

		BeforeEach(func() {
			adapter = &FakeAdapter{}
			watcher, _, _ := NewAppServiceStoreWatcher(adapter, cache.NewAppServiceCache())

			stopChan = make(chan struct{})
			go watcher.Run(stopChan)
		})

		AfterEach(func() {
			close(stopChan)
		})

		It("calls watch again", func() {
			Eventually(adapter.GetWatchCounter).Should(Equal(1))
			adapter.WatchErrChannel <- errors.New("Haha")
			Eventually(adapter.GetWatchCounter).Should(Equal(2))
		})
	})
})
