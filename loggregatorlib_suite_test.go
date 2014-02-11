package loggregatorlib_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLoggregatorlib(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Loggregatorlib Suite")
}
