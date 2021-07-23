package metrics

import (
	"context"
	"github.com/iris-connect/eps"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

type PrometheusMetricsServer struct {
	server *http.Server
}

func MakePrometheusMetricsServer(settings *eps.MetricsSettings) *PrometheusMetricsServer {

	if settings == nil {
		return nil
	}

	p := &PrometheusMetricsServer{
		server: &http.Server{Addr: settings.BindAddress, Handler: promhttp.Handler()},
	}

	eps.Log.Infof("Serving metrics on %s...", settings.BindAddress)

	go func() {
		if err := p.server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				eps.Log.Error(err)
			}
		}
	}()

	return p
}

func (p *PrometheusMetricsServer) Stop() error {

	eps.Log.Info("Shutting down Prometheus metrics server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.server.Shutdown(ctx)

	return nil
}
