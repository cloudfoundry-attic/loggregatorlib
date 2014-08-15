package store_test

import (
	"errors"
	. "github.com/cloudfoundry/loggregatorlib/store"
	"github.com/cloudfoundry/loggregatorlib/store/cache"
	"github.com/cloudfoundry/storeadapter"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sync"
)

var _ = Describe("AppServiceStoreWatcherUnit", func() {
	Context("when there is an error", func() {
		var adapter *FakeAdapter

		BeforeEach(func() {
			adapter = &FakeAdapter{}
			watcher, _, _ := NewAppServiceStoreWatcher(adapter, cache.NewAppServiceCache())

			go watcher.Run()
		})

		It("calls watch again", func() {
			Eventually(adapter.GetWatchCounter).Should(Equal(1))
			adapter.WatchErrChannel <- errors.New("Haha")
			Eventually(adapter.GetWatchCounter).Should(Equal(2))
		})
	})
})
