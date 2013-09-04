package servernamer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRenameOfFullNames(t *testing.T) {
	assert.Equal(t, ServerName("45.56.78.90:7645", "hostname"), "45-56-78-90-7645-hostname")
}

func TestRenameWithoutPort(t *testing.T) {
	assert.Equal(t, ServerName("45.56.78.90", "hostname"), "45-56-78-90-hostname")
}

func TestRenameOfHostnameWithDots(t *testing.T) {
	assert.Equal(t, ServerName("1", "hostname.acme.com"), "1-hostname.acme.com")
}
