package routerregistrar_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRouterregistrar(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Routerregistrar Suite")
}
