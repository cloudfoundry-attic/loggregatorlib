package routerregistrar

import (
	"encoding/json"
	"errors"
	"fmt"
	mbus "github.com/cloudfoundry/go_cfmessagebus"
	"github.com/cloudfoundry/gosteno"
	"time"
	"sync"
)

type registrar struct {
	*gosteno.Logger
	mBusClient mbus.MessageBus
	routerRegisterInterval	time.Duration
	lock sync.RWMutex
}

func NewRouterRegistrar(mBusClient mbus.MessageBus, logger *gosteno.Logger) *registrar {
	return &registrar{mBusClient: mBusClient, Logger: logger}
}

func (r *registrar) RegisterWithRouter(hostname string, port uint32, uris []string) error {
	r.subscribeToRouterStart()
	err := r.greetRouter()
	if err != nil {
		return err
	}
	r.keepRegisteringWithRouter(hostname, port, uris)

	return nil
}

func (r *registrar) greetRouter() (err error) {
	response := make(chan []byte)

	r.mBusClient.Request(RouterGreetMessageSubject, []byte{}, func(payload []byte) {
		response <- payload
	})

	routerRegisterInterval := 20 * time.Second
	select {
	case msg := <-response:
		routerResponse := &RouterResponse{}
		err = json.Unmarshal(msg, routerResponse)
		if err != nil {
			r.Errorf("Error unmarshalling the greet response: %v\n", err)
		} else {
			routerRegisterInterval = routerResponse.RegisterInterval * time.Second
			r.Infof("Greeted the router. Setting register interval to %v seconds\n", routerResponse.RegisterInterval)

		}
	case <-time.After(30 * time.Second):
		err = errors.New("Did not get a response to router.greet!")
	}

	r.lock.Lock()
	r.routerRegisterInterval = routerRegisterInterval
	r.lock.Unlock()

	return err
}

func (r *registrar) subscribeToRouterStart() {
	r.mBusClient.Subscribe(RouterStartMessageSubject, func(payload []byte) {
		routerResponse := &RouterResponse{}
		err := json.Unmarshal(payload, routerResponse)
		if err != nil {
			r.Errorf("Error unmarshalling the router start message: %v\n", err)
		} else {
			r.Infof("Received router.start. Setting register interval to %v seconds\n", routerResponse.RegisterInterval)
			r.lock.Lock()
			r.routerRegisterInterval = routerResponse.RegisterInterval * time.Second
			r.lock.Unlock()
		}
	})
	r.Info("Subscribed to router.start")

	return
}

func (r *registrar) keepRegisteringWithRouter(hostname string, port uint32, uris []string) {
	go func() {
		for {
			err := r.publishRouterMessage(hostname, port, uris, RouterRegisterMessageSubject)
			if err != nil {
				r.Error(err.Error())
			}
			r.Debug("Reregistered with router")
			<-time.After(r.routerRegisterInterval)
		}
	}()
}

func (r *registrar) UnregisterFromRouter(hostname string, port uint32, uris []string) {
	err := r.publishRouterMessage(hostname, port, uris, RouterUnregisterMessageSubject)
	if err != nil {
		r.Error(err.Error())
	}
	r.Info("Unregistered from router")
}

func (r *registrar) publishRouterMessage(hostname string, port uint32, uris []string, subject string) error {
	message := &RouterMessage{
		Host: hostname,
		Port: port,
		Uris: uris,
	}

	json, err := json.Marshal(message)
	if err != nil {
		return errors.New(fmt.Sprintf("Error marshalling the router message: %v\n", err))
	}

	err = r.mBusClient.Publish(subject, json)
	if err != nil {
		return errors.New(fmt.Sprintf("Publishing %s failed: %v", subject, err))
	}
	return nil
}
