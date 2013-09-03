package appid

import (
	testhelpers "github.com/cloudfoundry/loggregatorlib/lib_testhelpers"
	"github.com/stretchr/testify/assert"
	"net/url"
	"testing"
)

func TestFromUrl(t *testing.T) {
	theUrl, err := url.Parse("wss://loggregator.loggregatorci.cf-app.com:4443/tail/?app=11bfecc7-7128-4e56-83a0-d8e0814ed7e6")
	assert.NoError(t, err)
	appid := FromUrl(theUrl)
	assert.Equal(t, "11bfecc7-7128-4e56-83a0-d8e0814ed7e6", appid)
}

func TestFromLogMessage(t *testing.T) {
	message := testhelpers.MarshalledLogMessage(t, "message", "my_app_id")
	appid, err := FromLogMessage(message)
	assert.NoError(t, err)
	assert.Equal(t, "my_app_id", appid)
}
