package clientpool_test

import (
	"github.com/cloudfoundry/loggregatorlib/clientpool"

	"fmt"
	steno "github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/loggregatorlib/loggregatorclient"
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = BeforeSuite(func() {
	rand.Seed(int64(time.Now().Nanosecond()))
})

var _ = Describe("LoggregatorClientPool", func() {
	var (
		pool       *clientpool.LoggregatorClientPool
		logger     *steno.Logger
		fakeGetter *fakeAddressGetter
	)

	BeforeEach(func() {
		logger = steno.NewLogger("TestLogger")
		fakeGetter = &fakeAddressGetter{}
		pool = clientpool.NewLoggregatorClientPool(logger, 3456, fakeGetter)
	})

	Describe("ListClients", func() {
		Context("with empty address list", func() {
			It("returns an empty client list", func() {
				fakeGetter.addresses = []string{}
				Expect(pool.ListClients()).To(HaveLen(0))
			})
		})

		Context("with a non-empty address list", func() {
			It("returns a client for every address", func() {
				fakeGetter.addresses = []string{"127.0.0.1", "127.0.0.2"}
				Expect(pool.ListClients()).To(HaveLen(2))
			})
		})

		It("re-uses existing clients", func() {
			fakeGetter.addresses = []string{"127.0.0.1"}
			client1 := pool.ListClients()[0]
			client2 := pool.ListClients()[0]
			Expect(client1).To(Equal(client2))
		})
	})

	Describe("RandomClient", func() {
		Context("with a non-empty client pool", func() {
			It("chooses a client with roughly uniform distribution", func() {
				for i := 0; i < 5; i++ {
					fakeGetter.addresses = append(fakeGetter.addresses, fmt.Sprintf("127.0.0.%d", i))
				}

				counts := make(map[loggregatorclient.LoggregatorClient]int)
				for i := 0; i < 100000; i++ {
					pick, _ := pool.RandomClient()
					counts[pick]++
				}

				for _, count := range counts {
					Expect(count).To(BeNumerically("~", 20000, 500))
				}
			})
		})

		Context("with an empty client pool", func() {
			It("returns an error", func() {
				_, err := pool.RandomClient()
				Expect(err).To(Equal(clientpool.ErrorEmptyClientPool))
			})
		})
	})

})

type fakeAddressGetter struct {
	addresses []string
}

func (getter *fakeAddressGetter) GetAddresses() []string {
	return getter.addresses
}
