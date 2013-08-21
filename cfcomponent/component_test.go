package cfcomponent

import (
	"encoding/json"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent/instrumentation"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"runtime"
	"testing"
	"time"
)

type GoodHealthMonitor struct{}

func (hm GoodHealthMonitor) Ok() bool {
	return true
}

type BadHealthMonitor struct{}

func (hm BadHealthMonitor) Ok() bool {
	return false
}

func TestGoodHealthzEndpoint(t *testing.T) {
	component := &Component{
		HealthMonitor:     GoodHealthMonitor{},
		StatusPort:        7877,
		Type:              "loggregator",
		StatusCredentials: []string{"user", "pass"},
	}

	go component.StartMonitoringEndpoints()

	req, err := http.NewRequest("GET", "http://localhost:7877/healthz", nil)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, resp.StatusCode, 200)
	assert.Equal(t, resp.Header.Get("Content-Type"), "text/plain")
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, string(body), "ok")
}

func TestBadHealthzEndpoint(t *testing.T) {
	component := &Component{
		HealthMonitor:     BadHealthMonitor{},
		StatusPort:        9878,
		Type:              "loggregator",
		StatusCredentials: []string{"user", "pass"},
	}

	go component.StartMonitoringEndpoints()

	req, err := http.NewRequest("GET", "http://localhost:9878/healthz", nil)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, resp.StatusCode, 200)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, string(body), "bad")
}

func TestPanicWhenFailingToMonitorEndpoints(t *testing.T) {
	component := &Component{
		HealthMonitor:     GoodHealthMonitor{},
		StatusPort:        7879,
		Type:              "loggregator",
		StatusCredentials: []string{"user", "pass"},
	}

	finishChan := make(chan bool)

	go func() {
		err := component.StartMonitoringEndpoints()
		assert.NoError(t, err)
	}()
	time.Sleep(50 * time.Millisecond)

	go func() {
		//Monitoring a second time should fail because the port is already in use.
		err := component.StartMonitoringEndpoints()
		assert.Error(t, err)
		finishChan <- true
	}()

	<-finishChan
}

type testInstrumentable struct {
	name    string
	metrics []instrumentation.Metric
}

func (t testInstrumentable) Emit() instrumentation.Context {
	return instrumentation.Context{t.name, t.metrics}
}

func TestVarzEndpoint(t *testing.T) {
	component := &Component{
		HealthMonitor:     GoodHealthMonitor{},
		StatusPort:        1234,
		Type:              "loggregator",
		StatusCredentials: []string{"user", "pass"},
		Instrumentables: []instrumentation.Instrumentable{
			testInstrumentable{
				"agentListener",
				[]instrumentation.Metric{
					instrumentation.Metric{"messagesReceived", 2004},
					instrumentation.Metric{"queueLength", 5},
				},
			},
			testInstrumentable{
				"cfSinkServer",
				[]instrumentation.Metric{
					instrumentation.Metric{"activeSinkCount", 3},
				},
			},
		},
	}

	go component.StartMonitoringEndpoints()

	req, err := http.NewRequest("GET", "http://localhost:1234/varz", nil)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)

	assert.Equal(t, resp.StatusCode, 200)
	assert.Equal(t, resp.Header.Get("Content-Type"), "application/json")
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	expected := map[string]interface{}{
		"name":          "loggregator",
		"numCPUS":       runtime.NumCPU(),
		"numGoRoutines": runtime.NumGoroutine(),
		"contexts": []interface{}{
			map[string]interface{}{
				"name": "agentListener",
				"metrics": []interface{}{
					map[string]interface{}{
						"name":  "messagesReceived",
						"value": 2004,
					},
					map[string]interface{}{
						"name":  "queueLength",
						"value": 5,
					},
				},
			},
			map[string]interface{}{
				"name": "cfSinkServer",
				"metrics": []interface{}{
					map[string]interface{}{
						"name":  "activeSinkCount",
						"value": 3,
					},
				},
			},
		},
	}

	var actualMap map[string]interface{}
	json.Unmarshal(body, &actualMap)
	assert.Equal(t, expected, actualMap)
}
