package cfcomponent

import (
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent/instrumentation"
	"github.com/cloudfoundry/loggregatorlib/cfcomponent/localip"
	"net/http"
)

type Component struct {
	IpAddress         string
	HealthMonitor     HealthMonitor
	WebPort           uint32
	Type              string //Used by the collector to find data processing class
	Index             uint
	UUID              string
	StatusPort        uint32
	StatusCredentials []string
	Instrumentables   []instrumentation.Instrumentable
}

func NewComponent(webPort uint32, componentType string, index uint, heathMonitor HealthMonitor, statusPort uint32, statusCreds []string, instrumentables []instrumentation.Instrumentable) (Component, error) {
	ip, err := localip.LocalIP()
	if err != nil {
		return Component{}, err
	}

	return Component{
		IpAddress:         ip,
		WebPort:           webPort,
		Type:              componentType,
		Index:             index,
		HealthMonitor:     heathMonitor,
		StatusPort:        statusPort,
		StatusCredentials: statusCreds,
		Instrumentables:   instrumentables,
	}, nil
}

func (c Component) StartMonitoringEndpoints() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthzHandlerFor(c))
	mux.HandleFunc("/varz", varzHandlerFor(c))

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", c.IpAddress, c.StatusPort), mux)
	return err
}

func healthzHandlerFor(c Component) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		if c.HealthMonitor.Ok() {
			fmt.Fprintf(w, "ok")
		} else {
			fmt.Fprintf(w, "bad")
		}
	}
}

func varzHandlerFor(c Component) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		message, err := instrumentation.NewVarzMessage(c.Type, c.Instrumentables)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}

		json, err := json.Marshal(message)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		w.Write(json)
	}
}
