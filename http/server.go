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

package http

import (
	"context"
	cryptoTls "crypto/tls"
	"fmt"
	"github.com/iris-connect/eps"
	epsNet "github.com/iris-connect/eps/net"
	"github.com/iris-connect/eps/tls"
	"net"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Handler func(*Context)
type Decorator func(*Route) Handler

type RouteGroup struct {
	Routes    []*Route
	Handlers  []Handler
	Subgroups []*RouteGroup
}

type Route struct {
	Pattern  string
	Regexp   *regexp.Regexp
	Handlers []Handler
}

type H map[string]interface{}

type Hooks struct {
	Finished Handler
}

type HTTPServer struct {
	settings    *HTTPServerSettings
	tlsConfig   *cryptoTls.Config
	listener    net.Listener
	mutex       sync.Mutex
	running     bool
	hooks       *Hooks
	err         error
	server      *http.Server
	routeGroups []*RouteGroup
}

func initializeRouteGroup(routeGroup *RouteGroup) error {

	var err error

	for i, route := range routeGroup.Routes {
		// we only allow patterns that match the entire URL path...
		if !strings.HasPrefix(route.Pattern, "^") || !strings.HasSuffix(route.Pattern, "$") {
			return fmt.Errorf("route %d: not a complete regular expression (needs to match entire string)", i)
		}
		// we precompile all the regexp pattern in the routes
		if route.Regexp, err = regexp.Compile(route.Pattern); err != nil {
			return err
		}
	}
	for _, subgroup := range routeGroup.Subgroups {
		if err := initializeRouteGroup(subgroup); err != nil {
			return err
		}
	}
	return nil
}

func MakeHTTPServer(settings *HTTPServerSettings, routeGroups []*RouteGroup) (*HTTPServer, error) {

	for _, routeGroup := range routeGroups {
		if err := initializeRouteGroup(routeGroup); err != nil {
			return nil, err
		}
	}

	s := &HTTPServer{
		settings:    settings,
		routeGroups: routeGroups,
		mutex:       sync.Mutex{},
		hooks:       &Hooks{},
		server: &http.Server{
			Addr: settings.BindAddress,
			// we disable HTTP/2 for all servers as there seems to be a bug in the Golang
			// HTTP/2 implementation that causes EOF errors when reading from the server
			// response, which in turn causes trouble with our proxy server when terminating
			// This can be re-eneabled once the bug is fixed upstream...
			// more info: https://github.com/golang/go/issues/46071
			TLSNextProto: make(map[string]func(*http.Server, *cryptoTls.Conn, http.Handler)),
		},
	}

	// we add the handler
	s.server.Handler = s

	return s, nil
}

func (h *HTTPServer) SetHooks(hooks *Hooks) {
	h.hooks = hooks
}

func (h *HTTPServer) SetListener(listener net.Listener) {
	h.listener = listener
}

func (h *HTTPServer) SetTLSConfig(config *cryptoTls.Config) {
	h.tlsConfig = config
}

func handleRouteGroup(context *Context, group *RouteGroup, handlers []Handler) {

	for i, route := range group.Routes {
		// routes only match full URLs
		if groups := route.Regexp.FindStringSubmatch(context.Request.URL.Path); groups != nil {
			eps.Log.Debugf("Route %d matched path '%s'.", i, context.Request.URL.Path)
			context.RouteParams = groups[1:]
			// we combine the group handlers with the route handlers
			for j, handler := range append(append(handlers, group.Handlers...), route.Handlers...) {
				handler(context)
				if context.Aborted {
					eps.Log.Debugf("Handler %d aborted the processing.", j)
					// the handler has aborted the processing of this request
					// so we break out of the loop...
					break
				}
			}
		}
		if context.Aborted {
			break
		}
	}

	for _, subgroup := range group.Subgroups {
		handleRouteGroup(context, subgroup, append(handlers, group.Handlers...))
	}

}
func (s *HTTPServer) ServeHTTP(writer http.ResponseWriter, request *http.Request) {

	context := MakeContext(writer, request)

	for _, routeGroup := range s.routeGroups {
		handleRouteGroup(context, routeGroup, []Handler{})
	}

	if s.hooks.Finished != nil {
		s.hooks.Finished(context)
	}
}

func (s *HTTPServer) Start() error {

	var listener func() error

	useTLS := false
	if s.settings.TLS != nil && s.tlsConfig == nil {

		var err error
		s.tlsConfig, err = tls.TLSServerConfig(s.settings.TLS)

		if err != nil {
			return err
		}
	}

	if s.tlsConfig != nil {
		useTLS = true
		s.server.TLSConfig = s.tlsConfig
	}

	if s.listener == nil {
		if listener, err := net.Listen("tcp", s.settings.BindAddress); err != nil {
			return err
		} else if s.settings.TCPRateLimits != nil {
			s.listener = epsNet.MakeRateLimitedListener(listener, s.settings.TCPRateLimits)
		} else {
			s.listener = listener
		}
	}

	listener = func() error {
		if useTLS {
			return s.server.ServeTLS(s.listener, "", "")
		}
		return s.server.Serve(s.listener)
	}

	go func() {
		// always returns error. ErrServerClosed on graceful close
		if err := listener(); err != http.ErrServerClosed {

			// something went wrong, we log and store the error...

			eps.Log.Error(err)

			s.mutex.Lock()
			s.err = err
			s.running = false
			s.mutex.Unlock()

		} else {
			// the server shut down gracefully...
			s.mutex.Lock()
			s.running = false
			s.err = nil
			s.mutex.Unlock()
		}
	}()

	time.Sleep(time.Millisecond * 100)
	s.mutex.Lock()
	running := s.running
	s.mutex.Unlock()

	// we check if the server is running 1 second after starting it
	// if not something probably went wrong, so we return the error

	if !running {
		return s.err
	}

	// seems nothing went wrong
	return nil

}

func (s *HTTPServer) Stop() error {
	eps.Log.Debugf("Shutting down HTTP server...")
	return s.server.Shutdown(context.TODO())
}
