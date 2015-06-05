package routerregistrar

import (
	"github.com/apcera/nats"
	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"
	"github.com/cloudfoundry/yagnats/fakeyagnats"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Router Registrar", func() {
	It("greets router", func() {
		routerReceivedChannel := make(chan *nats.Msg, 10)
		resultChan := make(chan bool)

		mbus := fakeyagnats.Connect()
		fakeRouter(mbus, routerReceivedChannel)
		registrar := NewRouterRegistrar(mbus, loggertesthelper.Logger())

		go func() {
			err := registrar.greetRouter()
			Expect(err).NotTo(HaveOccurred())
		}()

		go func() {
			for {
				registrar.lock.RLock()
				if registrar.routerRegisterInterval == 42*time.Second {
					resultChan <- true
					registrar.lock.RUnlock()
					break
				}
				registrar.lock.RUnlock()
				time.Sleep(5 * time.Millisecond)
			}
		}()

		Eventually(resultChan, 2).Should(Receive())
		Expect(mbus.Subscriptions("router.greet")).To(HaveLen(1))
		Expect(mbus.Subscriptions("router.register")).To(HaveLen(1))
	})

	It("sets default interval when greet router fails", func() {
		routerReceivedChannel := make(chan *nats.Msg)
		resultChan := make(chan bool)

		mbus := fakeyagnats.Connect()
		fakeBrokenGreeterRouter(mbus, routerReceivedChannel)
		registrar := NewRouterRegistrar(mbus, loggertesthelper.Logger())

		go func() {
			err := registrar.greetRouter()
			Expect(err).To(HaveOccurred())
		}()

		go func() {
			for {
				registrar.lock.RLock()
				if registrar.routerRegisterInterval == 20*time.Second {
					resultChan <- true
					registrar.lock.RUnlock()
					break
				}
				registrar.lock.RUnlock()
				time.Sleep(5 * time.Millisecond)
			}
		}()

		Eventually(resultChan, 2).Should(Receive())
	})

	It("sets default interval when greet without router", func() {
		resultChan := make(chan bool)

		mbus := fakeyagnats.Connect()
		registrar := NewRouterRegistrar(mbus, loggertesthelper.Logger())

		go func() {
			err := registrar.greetRouter()
			Expect(err).To(HaveOccurred())
		}()

		go func() {
			for {
				registrar.lock.RLock()
				if registrar.routerRegisterInterval == 20*time.Second {
					resultChan <- true
					registrar.lock.RUnlock()
					break
				}
				registrar.lock.RUnlock()
				time.Sleep(5 * time.Millisecond)
			}
		}()

		Eventually(resultChan, 32).Should(Receive())
	})

	It("keeps registering with router", func() {
		mbus := fakeyagnats.Connect()
		os.Setenv("LOG_TO_STDOUT", "false")
		routerReceivedChannel := make(chan *nats.Msg)
		fakeRouter(mbus, routerReceivedChannel)

		registrar := NewRouterRegistrar(mbus, loggertesthelper.Logger())
		registrar.routerRegisterInterval = 50 * time.Millisecond
		registrar.keepRegisteringWithRouter("13.12.14.15", 8083, []string{"foobar.vcap.me"})

		var message *nats.Msg
		Eventually(routerReceivedChannel).Should(Receive(&message))
		Expect(string(message.Data)).To(Equal(`registering:{"host":"13.12.14.15","port":8083,"uris":["foobar.vcap.me"]}`))
	})

	It("subscribes to router start", func() {
		mbus := fakeyagnats.Connect()
		registrar := NewRouterRegistrar(mbus, loggertesthelper.Logger())
		registrar.subscribeToRouterStart()

		err := mbus.Publish("router.start", []byte(messageFromRouter))
		Expect(err).NotTo(HaveOccurred())

		resultChan := make(chan bool)
		go func() {
			for {
				registrar.lock.RLock()
				if registrar.routerRegisterInterval == 42*time.Second {
					resultChan <- true
					registrar.lock.RUnlock()
					break
				}
				registrar.lock.RUnlock()
				time.Sleep(5 * time.Millisecond)
			}
		}()

		Eventually(resultChan, 2).Should(Receive())
	})

	It("unregisters from router", func() {
		mbus := fakeyagnats.Connect()
		routerReceivedChannel := make(chan *nats.Msg, 10)
		fakeRouter(mbus, routerReceivedChannel)

		registrar := NewRouterRegistrar(mbus, loggertesthelper.Logger())
		registrar.UnregisterFromRouter("13.12.14.15", 8083, []string{"foobar.vcap.me"})

		host := "13.12.14.15"
		var message *nats.Msg
		Eventually(routerReceivedChannel, 2).Should(Receive(&message))
		Expect(string(message.Data)).To(Equal(`unregistering:{"host":"` + host + `","port":8083,"uris":["foobar.vcap.me"]}`))
	})
})

const messageFromRouter = `{
  							"id": "some-router-id",
  							"hosts": ["1.2.3.4"],
  							"minimumRegisterIntervalInSeconds": 42
							}`

func fakeRouter(mbus *fakeyagnats.FakeNATSConn, returnChannel chan *nats.Msg) {
	mbus.Subscribe("router.greet", func(msg *nats.Msg) {
		mbus.Publish(msg.Reply, []byte(messageFromRouter))
	})

	mbus.Subscribe("router.register", func(msg *nats.Msg) {
		returnChannel <- &nats.Msg{
			Subject: msg.Subject,
			Reply:   msg.Reply,
			Data:    []byte("registering:" + string(msg.Data)),
		}

		mbus.Publish(msg.Reply, msg.Data)
	})

	mbus.Subscribe("router.unregister", func(msg *nats.Msg) {
		returnChannel <- &nats.Msg{
			Subject: msg.Subject,
			Reply:   msg.Reply,
			Data:    []byte("unregistering:" + string(msg.Data)),
		}
		mbus.Publish(msg.Reply, msg.Data)
	})
}

func fakeBrokenGreeterRouter(mbus *fakeyagnats.FakeNATSConn, returnChannel chan *nats.Msg) {

	mbus.Subscribe("router.greet", func(msg *nats.Msg) {
		mbus.Publish(msg.Reply, []byte("garbel garbel"))
	})
}
