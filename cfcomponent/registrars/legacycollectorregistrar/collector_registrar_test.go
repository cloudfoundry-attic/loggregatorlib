package legacycollectorregistrar

import (
	"encoding/json"

	"github.com/apcera/nats"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent"
	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"
	"github.com/cloudfoundry/yagnats/fakeyagnats"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Legacy Collector Registrar", func() {
	It("announce component", func() {
		cfc := getTestCFComponent()
		mbus := fakeyagnats.Connect()

		called := make(chan *nats.Msg, 10)
		mbus.Subscribe(AnnounceComponentMessageSubject, func(response *nats.Msg) {
			called <- response
		})

		registrar := NewCollectorRegistrar(mbus, loggertesthelper.Logger())
		registrar.announceComponent(cfc)

		expectedJson, _ := createYagnatsMessage(GinkgoT(), AnnounceComponentMessageSubject)

		payloadBytes := (<-called).Data
		Expect(payloadBytes).To(Equal(expectedJson))
	})

	It("subscribe to component discover", func() {
		cfc := getTestCFComponent()
		mbus := fakeyagnats.Connect()

		called := make(chan *nats.Msg, 10)
		mbus.Subscribe(DiscoverComponentMessageSubject, func(response *nats.Msg) {
			called <- response
		})

		registrar := NewCollectorRegistrar(mbus, loggertesthelper.Logger())
		registrar.subscribeToComponentDiscover(cfc)

		expectedJson, _ := createYagnatsMessage(GinkgoT(), DiscoverComponentMessageSubject)
		mbus.PublishRequest(DiscoverComponentMessageSubject, "unused-reply", expectedJson)

		payloadBytes := (<-called).Data
		Expect(payloadBytes).To(Equal(expectedJson))
	})
})

func createYagnatsMessage(t GinkgoTInterface, subject string) ([]byte, *nats.Msg) {

	expected := &AnnounceComponentMessage{
		Type:        "Loggregator Server",
		Index:       0,
		Host:        "1.2.3.4:5678",
		UUID:        "0-abc123",
		Credentials: []string{"user", "pass"},
	}

	expectedJson, err := json.Marshal(expected)
	Expect(err).NotTo(HaveOccurred())

	yagnatsMsg := &nats.Msg{
		Subject: subject,
		Reply:   "reply_to",
		Data:    expectedJson,
	}

	return expectedJson, yagnatsMsg
}

func getTestCFComponent() cfcomponent.Component {
	return cfcomponent.Component{
		IpAddress:         "1.2.3.4",
		Type:              "Loggregator Server",
		Index:             0,
		StatusPort:        5678,
		StatusCredentials: []string{"user", "pass"},
		UUID:              "abc123",
	}
}
