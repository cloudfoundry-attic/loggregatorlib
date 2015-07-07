package collectorregistrar_test

import (
	"encoding/json"

	"github.com/cloudfoundry/loggregatorlib/cfcomponent"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent/registrars/collectorregistrar"
	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Collectormessages", func() {
	Context("given a component without a jobName", func() {
		var component cfcomponent.Component

		BeforeEach(func() {
			component, _ = cfcomponent.NewComponent(loggertesthelper.Logger(), "compType", 3, nil, 9999, []string{"username", "password"}, nil)
			component.UUID = "OurUUID"
		})

		It("should not marshal jobName to JSON", func() {
			json, _ := json.Marshal(collectorregistrar.NewAnnounceComponentMessage(component))
			Expect(json).To(MatchRegexp(`^\{"type":"compType","index":3,"host":"[^:]*:9999","uuid":"3-OurUUID","credentials":\["username","password"\]\}$`))
		})
	})

	Context("given a component with a jobName", func() {
		var component cfcomponent.Component

		BeforeEach(func() {
			component, _ = cfcomponent.NewComponent(loggertesthelper.Logger(), "compType", 3, nil, 9999, []string{"username", "password"}, nil, "jobname")
			component.UUID = "OurUUID"
		})

		It("should marshal jobName to JSON", func() {
			json, _ := json.Marshal(collectorregistrar.NewAnnounceComponentMessage(component))
			Expect(json).To(MatchRegexp(`^\{"type":"compType","index":3,"job_name":"jobname","host":"[^:]*:9999","uuid":"3-OurUUID","credentials":\["username","password"\]\}$`))
		})

	})
})
