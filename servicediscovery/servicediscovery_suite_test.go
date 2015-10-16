package servicediscovery_test

import (
	"github.com/cloudfoundry/storeadapter/storerunner/etcdstorerunner"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"testing"
)

var etcdRunner *etcdstorerunner.ETCDClusterRunner

func TestServicediscovery(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ServiceDiscovery Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	return nil
}, func(encodedBuiltArtifacts []byte) {
	etcdPort := 5000 + (config.GinkgoConfig.ParallelNode)*10
	etcdRunner = etcdstorerunner.NewETCDClusterRunner(etcdPort, 1, nil)

	etcdRunner.Start()
})

var _ = SynchronizedAfterSuite(func() {
	if etcdRunner != nil {
		etcdRunner.Stop()
	}
}, func() {
	gexec.CleanupBuildArtifacts()
})

var _ = BeforeEach(func() {
	etcdRunner.Reset()
})
