package routerregistrar

import (
	mbus "github.com/cloudfoundry/go_cfmessagebus"
	"github.com/cloudfoundry/go_cfmessagebus/mock_cfmessagebus"
	testhelpers "github.com/cloudfoundry/loggregatorlib/lib_testhelpers"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestGreetRouter(t *testing.T) {
	mbus := mock_cfmessagebus.NewMockMessageBus()
	routerReceivedChannel := make(chan []byte)
	fakeRouter(mbus, routerReceivedChannel)

	registrar := NewRouterRegistrar(mbus, testhelpers.Logger())
	err := registrar.greetRouter()
	assert.NoError(t, err)

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

	select {
	case <-resultChan:
	case <-time.After(2 * time.Second):
		t.Error("Router did not receive a router.start in time!")
	}
}

func TestDefaultIntervalIsSetWhenGreetRouterFails(t *testing.T) {
	mbus := mock_cfmessagebus.NewMockMessageBus()
	routerReceivedChannel := make(chan []byte)
	fakeBrokenGreeterRouter(mbus, routerReceivedChannel)

	registrar := NewRouterRegistrar(mbus, testhelpers.Logger())
	err := registrar.greetRouter()
	assert.Error(t, err)

	resultChan := make(chan bool)
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

	select {
	case <-resultChan:
	case <-time.After(2 * time.Second):
		t.Error("Default register interval was never set!")
	}
}

func TestDefaultIntervalIsSetWhenGreetWithoutRouter(t *testing.T) {
	mbus := mock_cfmessagebus.NewMockMessageBus()
	registrar := NewRouterRegistrar(mbus, testhelpers.Logger())
	err := registrar.greetRouter()
	assert.Error(t, err)

	resultChan := make(chan bool)
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

	select {
	case <-resultChan:
	case <-time.After(2 * time.Second):
		t.Error("Default register interval was never set!")
	}
}

func TestKeepRegisteringWithRouter(t *testing.T) {
	mbus := mock_cfmessagebus.NewMockMessageBus()
	os.Setenv("LOG_TO_STDOUT", "false")
	routerReceivedChannel := make(chan []byte)
	fakeRouter(mbus, routerReceivedChannel)

	registrar := NewRouterRegistrar(mbus, testhelpers.Logger())
	registrar.routerRegisterInterval = 50 * time.Millisecond
	registrar.keepRegisteringWithRouter("13.12.14.15", 8083, []string{"foobar.vcap.me"})

	for i := 0; i < 3; i++ {
		time.Sleep(55 * time.Millisecond)
		select {
		case msg := <-routerReceivedChannel:
			assert.Equal(t, `registering:{"host":"13.12.14.15","port":8083,"uris":["foobar.vcap.me"]}`, string(msg))
		default:
			t.Error("Router did not receive a router.register in time!")
		}
	}
}

func TestSubscribeToRouterStart(t *testing.T) {
	mbus := mock_cfmessagebus.NewMockMessageBus()
	registrar := NewRouterRegistrar(mbus, testhelpers.Logger())
	registrar.subscribeToRouterStart()

	err := mbus.Publish("router.start", []byte(messageFromRouter))
	assert.NoError(t, err)

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

	select {
	case <-resultChan:
	case <-time.After(2 * time.Second):
		t.Error("Router did not receive a router.start in time!")
	}
}

func TestUnregisterFromRouter(t *testing.T) {
	mbus := mock_cfmessagebus.NewMockMessageBus()
	routerReceivedChannel := make(chan []byte)
	fakeRouter(mbus, routerReceivedChannel)

	registrar := NewRouterRegistrar(mbus, testhelpers.Logger())
	registrar.UnregisterFromRouter("13.12.14.15", 8083, []string{"foobar.vcap.me"})

	select {
	case msg := <-routerReceivedChannel:
		host := "13.12.14.15"
		assert.Equal(t, `unregistering:{"host":"`+host+`","port":8083,"uris":["foobar.vcap.me"]}`, string(msg))
	case <-time.After(2 * time.Second):
		t.Error("Router did not receive a router.unregister in time!")
	}
}

const messageFromRouter = `{
  							"id": "some-router-id",
  							"hosts": ["1.2.3.4"],
  							"minimumRegisterIntervalInSeconds": 42
							}`

func fakeRouter(mbus mbus.MessageBus, returnChannel chan []byte) {
	mbus.RespondToChannel("router.greet", func(_ []byte) []byte {
		response := []byte(messageFromRouter)
		return response
	})

	mbus.RespondToChannel("router.register", func(content []byte) []byte {
		returnChannel <- []byte("registering:" + string(content))
		return content
	})

	mbus.RespondToChannel("router.unregister", func(content []byte) []byte {
		returnChannel <- []byte("unregistering:" + string(content))
		return content
	})
}

func fakeBrokenGreeterRouter(mbus mbus.MessageBus, returnChannel chan []byte) {
	mbus.RespondToChannel("router.greet", func(_ []byte) []byte {
		response := []byte("garbel garbel")
		return response
	})
}
