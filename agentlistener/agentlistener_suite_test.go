package agentlistener_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestAgentlistener(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Agentlistener Suite")
}
