package metrics

import (
	"fmt"
	"github.com/iris-connect/eps"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
)

const (
	prometheusPort = 2112
	prometheusPath = "/metrics"
)

func OpenPrometheusEndpoint() {
	if os.Getenv("METRICS_ENABLED") != "true" {
		return
	}

	go func() {
		eps.Log.Infof("Starting Prometheus listener at http://localhost:%d%s", prometheusPort, prometheusPath)
		http.Handle(prometheusPath, promhttp.Handler())
		err := http.ListenAndServe(fmt.Sprintf(":%d", prometheusPort), nil)
		if err != nil {
			eps.Log.Error("Could not open prometheus endpoint: ", err)
		}
	}()
}
