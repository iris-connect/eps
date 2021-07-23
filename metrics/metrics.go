// IRIS Endpoint-Server (EPS)
// Copyright (C) 2021-2021 The IRIS Endpoint-Server Authors (see AUTHORS.md)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
