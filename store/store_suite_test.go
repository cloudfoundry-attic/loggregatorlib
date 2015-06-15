package store_test

import (
	"path"
	"testing"

	"github.com/cloudfoundry/loggregatorlib/appservice"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent"
	"github.com/cloudfoundry/loggregatorlib/loggertesthelper"
	"github.com/cloudfoundry/storeadapter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestStore(t *testing.T) {
	RegisterFailHandler(Fail)
	cfcomponent.Logger = loggertesthelper.Logger()

	RunSpecs(t, "Store Suite")
}

func buildNode(appService appservice.AppService) storeadapter.StoreNode {
	return storeadapter.StoreNode{
		Key:   path.Join("/loggregator/services", appService.AppId, appService.Id()),
		Value: []byte(appService.Url),
	}
}
