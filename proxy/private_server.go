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

/*
The private proy server creates outgoing TCP connections to the public proxy
server and forwards them to an internal endpoint.
*/

package proxy

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/iris-connect/eps"
	epsForms "github.com/iris-connect/eps/forms"
	"github.com/iris-connect/eps/helpers"
	"github.com/iris-connect/eps/http"
	"github.com/iris-connect/eps/jsonrpc"
	epsTls "github.com/iris-connect/eps/tls"
	"github.com/kiprotect/go-helpers/forms"
	"io/ioutil"
	"net"
	goHttp "net/http"
	"strings"
	"sync"
	"time"
)

type PrivateServer struct {
	dataStore     eps.Datastore
	settings      *PrivateServerSettings
	announcements []*PrivateAnnouncement
	jsonrpcServer *jsonrpc.JSONRPCServer
	jsonrpcClient *jsonrpc.Client
	tlsConfig     *tls.Config
	stop          chan bool
	l             net.Listener
	mutex         sync.Mutex
}

type ProxyConnection struct {
	settings      *InternalEndpointSettings
	tlsConfig     *tls.Config
	jsonrpcClient *jsonrpc.Client
	proxyEndpoint string
	token         []byte
}

func MakeProxyConnection(proxyEndpoint string, token []byte, settings *InternalEndpointSettings, tlsConfig *tls.Config) *ProxyConnection {
	p := &ProxyConnection{
		settings:      settings,
		tlsConfig:     tlsConfig,
		proxyEndpoint: proxyEndpoint,
		token:         token,
	}
	if settings.JSONRPCClient != nil {
		p.jsonrpcClient = jsonrpc.MakeClient(settings.JSONRPCClient)
	}
	return p
}

func (p *ProxyConnection) Run() error {

	proxyConnection, err := net.Dial("tcp", p.proxyEndpoint)

	if err != nil {
		return err
	}

	if n, err := proxyConnection.Write(p.token); err != nil {
		proxyConnection.Close()
		return err
	} else if n != len(p.token) {
		proxyConnection.Close()
		return fmt.Errorf("could not write token")
	}

	if p.settings.TLS != nil {
		return p.TerminateTLS(proxyConnection)
	} else {
		return p.ForwardTLS(proxyConnection)
	}
}

type ProxyListener struct {
	connection net.Conn
	close      chan bool
}

func (p *ProxyListener) Accept() (net.Conn, error) {
	if p.connection != nil {
		defer func() { p.connection = nil }()
		return p.connection, nil
	} else {
		// we block indefinitely
		p.close <- true
		return nil, fmt.Errorf("no more connections")
	}
}

func (p *ProxyListener) Close() error {
	p.connection = nil
	// if the connection was already requested we resolve the block
	select {
	case <-p.close:
	}
	return nil
}

// we don't implement Addr
func (p *ProxyListener) Addr() net.Addr {
	return nil
}

type Server interface {
	Start() error
	Stop() error
}

func (p *ProxyConnection) jsonrpcHandler(done chan bool) func(request *jsonrpc.Context) *jsonrpc.Response {
	return func(c *jsonrpc.Context) *jsonrpc.Response {
		jsonData, err := json.Marshal(c.Request)

		if err != nil {
			eps.Log.Error(err)
			c.HTTPContext.AbortWithStatus(goHttp.StatusInternalServerError)
			return nil
		}

		jsonrpcEndpoint := fmt.Sprintf("%s", p.settings.JSONRPCClient.Endpoint)

		eps.Log.Debugf("Forwarding JSON-RPC request to '%s'", jsonrpcEndpoint)

		proxyRequest, err := goHttp.NewRequest("POST", jsonrpcEndpoint, bytes.NewReader(jsonData))

		if err != nil {
			eps.Log.Errorf("Cannot form JSON-RPC request: %v", err)
			c.HTTPContext.AbortWithStatus(goHttp.StatusInternalServerError)
			return nil
		}
		proxyRequest.Header = make(goHttp.Header)
		for k, v := range c.HTTPContext.Request.Header {
			proxyRequest.Header[k] = v
		}

		httpClient := goHttp.Client{}

		if resp, err := httpClient.Do(proxyRequest); err != nil {
			eps.Log.Errorf("An error occurred when forwarding the JSON-RPC request: %v", err)
			c.HTTPContext.AbortWithStatus(goHttp.StatusBadGateway)
			return nil
		} else {
			eps.Log.Debugf("Request successfully proxied, returning response...")
			c.HTTPContext.AbortWithResponse(resp)
		}

		return nil
	}
}

func (p *ProxyConnection) httpHandler(done chan bool) func(c *http.Context) {

	return func(c *http.Context) {
		// we need to buffer the body if we want to read it here and send it
		// in the request.
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatus(goHttp.StatusInternalServerError)
			return
		}

		pathAndQuery := c.Request.URL.Path

		if c.Request.URL.RawQuery != "" {
			pathAndQuery += fmt.Sprintf("?%s", c.Request.URL.RawQuery)
		}

		proxyRequest, err := goHttp.NewRequest(c.Request.Method, fmt.Sprintf("http://%s%s", p.settings.Address, pathAndQuery), bytes.NewReader(body))
		if err != nil {
			eps.Log.Error(err)
			c.AbortWithStatus(goHttp.StatusInternalServerError)
			return
		}
		proxyRequest.Header = make(goHttp.Header)
		for k, v := range c.Request.Header {
			proxyRequest.Header[k] = v
		}

		httpClient := goHttp.Client{}

		if resp, err := httpClient.Do(proxyRequest); err != nil {
			c.AbortWithStatus(goHttp.StatusBadGateway)
		} else {
			c.AbortWithResponse(resp)
		}

	}
}

func (p *ProxyConnection) TerminateTLS(proxyConnection net.Conn) error {

	proxyListener := &ProxyListener{
		connection: proxyConnection,
		close:      make(chan bool),
	}

	var server Server
	var httpServer *http.HTTPServer

	done := make(chan bool, 1)
	if p.settings.JSONRPCClient != nil {
		if jsonrpcServer, err := jsonrpc.MakeJSONRPCServer(&jsonrpc.JSONRPCServerSettings{
			Cors:        nil,
			TLS:         nil,
			Path:        p.settings.JSONRPCPath,
			BindAddress: "",
		}, p.jsonrpcHandler(done)); err != nil {
			return err
		} else {
			server = jsonrpcServer
			httpServer = jsonrpcServer.HTTPServer()
		}
	} else {
		var err error

		routeGroups := []*http.RouteGroup{
			{
				// these handlers will be executed for all routes in the group
				Handlers: []http.Handler{},
				Routes: []*http.Route{
					{
						Pattern: "^.*$",
						Handlers: []http.Handler{
							p.httpHandler(done),
						},
					},
				},
			},
		}

		if httpServer, err = http.MakeHTTPServer(&http.HTTPServerSettings{
			TLS:         nil,
			BindAddress: "",
		}, routeGroups); err != nil {
			return err
		} else {
			server = httpServer
		}
	}

	httpServer.SetListener(proxyListener)
	httpServer.SetTLSConfig(p.tlsConfig)
	httpServer.SetHooks(&http.Hooks{
		Finished: func(c *http.Context) {
			done <- true
		},
	})

	if err := server.Start(); err != nil {
		eps.Log.Errorf("Error: %v", err)
		return err
	} else {
		defer server.Stop()
		select {
		case <-done:
		case <-time.After(time.Duration(p.settings.Timeout) * time.Second):
			break
			return fmt.Errorf("timeout handling request")
		}
	}

	proxyConnection.Close()

	return nil
}

func (p *ProxyConnection) ForwardTLS(proxyConnection net.Conn) error {

	eps.Log.Debugf("Forwarding TLS connection to '%s'", p.settings.Address)

	internalConnection, err := net.Dial("tcp", p.settings.Address)

	if err != nil {
		proxyConnection.Close()
		return err
	}

	close := func() {
		proxyConnection.Close()
		internalConnection.Close()
	}

	pipe := func(left, right net.Conn) {
		buf := make([]byte, 1024)
		for {
			n, err := left.Read(buf)
			if err != nil {
				eps.Log.Error(err)
				close()
				return
			}
			if m, err := right.Write(buf[:n]); err != nil {
				eps.Log.Error(err)
				close()
				return
			} else if m != n {
				eps.Log.Errorf("cannot write all data")
				close()
				return
			}
		}
	}

	go pipe(internalConnection, proxyConnection)
	go pipe(proxyConnection, internalConnection)

	return nil
}

var IncomingConnectionForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "token",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MinLength: 32,
					MaxLength: 32,
				},
			},
		},
		{
			Name: "endpoint",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "domain",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "_client",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &epsForms.ClientInfoForm,
				},
			},
		},
	},
}

var GetPrivateAnnouncementsForm = forms.Form{
	Fields: []forms.Field{},
}

type GetPrivateAnnouncementsParams struct{}

func (c *PrivateServer) getAnnouncements(context *jsonrpc.Context, params *GetPrivateAnnouncementsParams) *jsonrpc.Response {
	relevantAnnouncements := make([]*PrivateAnnouncement, 0)
	for _, announcement := range c.announcements {
		if announcement.ExpiresAt != nil && announcement.ExpiresAt.Before(time.Now()) {
			continue
		}
		relevantAnnouncements = append(relevantAnnouncements, announcement)
	}
	return context.Result(relevantAnnouncements)
}

type IncomingConnectionParams struct {
	Domain   string          `json:"domain"`
	Endpoint string          `json:"endpoint"`
	Token    []byte          `json:"token"`
	Client   *eps.ClientInfo `json:"_client"`
}

func (c *PrivateServer) incomingConnection(context *jsonrpc.Context, params *IncomingConnectionParams) *jsonrpc.Response {

	eps.Log.Debugf("Incoming connection for domain '%s' from '%s', ID: %s", params.Domain, params.Client.Name, context.Request.ID)

	found := false
	for _, announcement := range c.announcements {
		if announcement.Proxy == params.Client.Name && announcement.Domain == params.Domain {
			// we make sure the announcement is not expired
			if announcement.ExpiresAt != nil && announcement.ExpiresAt.Before(time.Now()) {
				continue
			}
			found = true
			break
		}
	}

	if !found {
		eps.Log.Debugf("No matching announcement found, closing...")
		return context.Error(404, "no matching connection found", nil)
	}

	connection := MakeProxyConnection(params.Endpoint, params.Token, c.settings.InternalEndpoint, c.tlsConfig)

	go func() {
		if err := connection.Run(); err != nil {
			eps.Log.Error(err)
		}
	}()

	return context.Result(map[string]interface{}{"message": "ok"})
}

func MakePrivateServer(settings *PrivateServerSettings, definitions *eps.Definitions) (*PrivateServer, error) {

	dataStore, err := helpers.InitializeDatastore(settings.Datastore, definitions)

	if err != nil {
		return nil, err
	}

	server := &PrivateServer{
		stop:          make(chan bool),
		settings:      settings,
		dataStore:     dataStore,
		jsonrpcClient: jsonrpc.MakeClient(settings.JSONRPCClient),
	}

	if settings.InternalEndpoint.TLS != nil {
		if tlsConfig, err := epsTls.TLSServerConfig(settings.InternalEndpoint.TLS); err != nil {
			return nil, err
		} else {
			server.tlsConfig = tlsConfig
		}
	}

	methods := map[string]*jsonrpc.Method{
		"incomingConnection": {
			Form:    &IncomingConnectionForm,
			Handler: server.incomingConnection,
		},
		"announceConnection": {
			Form:    &PrivateAnnounceConnectionForm,
			Handler: server.announceConnection,
		},
		"getAnnouncements": {
			Form:    &GetPrivateAnnouncementsForm,
			Handler: server.getAnnouncements,
		},
	}

	if err := server.dataStore.Init(); err != nil {
		return nil, err
	}

	if err := server.update(); err != nil {
		return nil, err
	}

	handler, err := jsonrpc.MethodsHandler(methods)

	if err != nil {
		return nil, err
	}

	jsonrpcServer, err := jsonrpc.MakeJSONRPCServer(settings.JSONRPCServer, handler)

	if err != nil {
		return nil, err
	}

	server.jsonrpcServer = jsonrpcServer

	return server, nil

}

var PrivateAnnounceConnectionForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "expires_at",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsString{},
				forms.IsTime{
					Format: "rfc3339",
				},
				IsValidExpiresAtTime{},
			},
		},
		{
			Name: "proxy",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "domain",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "_client",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &epsForms.ClientInfoForm,
				},
			},
		},
	},
}

type PrivateAnnounceConnectionParams struct {
	ClientInfo *eps.ClientInfo `json:"_client"`
	ExpiresAt  *time.Time      `json:"expires_at"`
	Domain     string          `json:"domain"`
	Proxy      string          `json:"proxy"`
}

func (c *PrivateServer) announceConnection(context *jsonrpc.Context, params *PrivateAnnounceConnectionParams) *jsonrpc.Response {

	settings := params.ClientInfo.Entry.SettingsFor("proxy", c.settings.Name)

	if params.Proxy == c.settings.Name {
		return context.Error(400, "trying to announce with private proxy name", nil)
	}

	if settings == nil {
		return context.Error(403, "not authorized", nil)
	} else {
		directorySettings := &DirectorySettings{}

		if directoryParams, err := DirectorySettingsForm.Validate(settings.Settings); err != nil {
			return context.Error(500, "invalid directory settings", nil)
		} else if err := DirectorySettingsForm.Coerce(directorySettings, directoryParams); err != nil {
			return context.InternalError()
		} else {
			found := false
			for _, allowedDomain := range directorySettings.AllowedDomains {
				if strings.HasSuffix(params.Domain, allowedDomain) {
					found = true
					break
				}
			}
			if !found {
				return context.Error(403, "not allowed to forward this domain", directorySettings)
			}
		}
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	var newAnnouncement *PrivateAnnouncement
	changed := false
	for _, announcement := range c.announcements {
		if announcement.Domain == params.Domain && announcement.Proxy == params.Proxy {
			newAnnouncement = announcement
			if params.ExpiresAt != announcement.ExpiresAt || announcement.ExpiresAt != nil && params.ExpiresAt != nil && !announcement.ExpiresAt.Equal(*params.ExpiresAt) {
				changed = true
				// we update the expiration time
				announcement.ExpiresAt = params.ExpiresAt
			}
			break
		}
	}

	if newAnnouncement == nil {
		newAnnouncement = &PrivateAnnouncement{
			Domain:    params.Domain,
			ExpiresAt: params.ExpiresAt,
			Proxy:     params.Proxy,
		}
		c.announcements = append(c.announcements, newAnnouncement)
		changed = true
	}

	if changed {

		id, err := helpers.RandomID(16)

		if err != nil {
			eps.Log.Error(err)
			return context.InternalError()
		}

		rawData, err := json.Marshal(newAnnouncement)

		if err != nil {
			eps.Log.Error(err)
			return context.InternalError()
		}

		dataEntry := &eps.DataEntry{
			Type: PrivateAnnouncementType,
			ID:   id,
			Data: rawData,
		}

		if err := c.dataStore.Write(dataEntry); err != nil {
			eps.Log.Error(err)
			return context.InternalError()
		}

		if err := c.announceConnectionsRPC([]*PrivateAnnouncement{newAnnouncement}); err != nil {
			eps.Log.Error(err)
			return context.InternalError()
		}

	}

	return context.Acknowledge()
}

// this method requires that *all* announcements are for the same proxy!
func (s *PrivateServer) announceConnectionsRPC(announcements []*PrivateAnnouncement) error {
	if len(announcements) == 0 {
		return fmt.Errorf("expected at least one announcement")
	}
	proxy := announcements[0].Proxy
	for _, announcement := range announcements[1:] {
		if announcement.Proxy != proxy {
			return fmt.Errorf("expected all announcements for the same proxy")
		}
	}
	eps.Log.Debugf("Sending %d announcements for proxy '%s'", len(announcements), proxy)
	request := jsonrpc.MakeRequest(fmt.Sprintf("%s.announceConnections", proxy), "", map[string]interface{}{
		"connections": announcements,
	})
	result, err := s.jsonrpcClient.Call(request)

	if err != nil {
		return err
	}

	if result.Error != nil {
		eps.Log.Error(result.Error)
		return fmt.Errorf(result.Error.Message)
	}

	return nil

}

func (s *PrivateServer) announceConnections() {
loop:
	for {

		groupedAnnouncements := map[string][]*PrivateAnnouncement{}

		s.mutex.Lock()
		for _, announcement := range s.announcements {
			var announcements []*PrivateAnnouncement
			var ok bool
			if announcements, ok = groupedAnnouncements[announcement.Proxy]; !ok {
				announcements = []*PrivateAnnouncement{}
			}
			announcements = append(announcements, announcement)
			groupedAnnouncements[announcement.Proxy] = announcements

		}
		s.mutex.Unlock()

		for _, announcements := range groupedAnnouncements {
			if err := s.announceConnectionsRPC(announcements); err != nil {
				eps.Log.Error(err)
			}
		}

		select {
		case <-time.After(10 * time.Minute):
		case <-s.stop:
			s.stop <- true
			break loop
		}
	}
}

func (s *PrivateServer) Start() error {
	go s.announceConnections()
	return s.jsonrpcServer.Start()
}

func (s *PrivateServer) Stop() error {

	s.stop <- true
	select {
	case <-s.stop:
	case <-time.After(5 * time.Second):
		eps.Log.Error("timeout when closing announcements")
	}

	return s.jsonrpcServer.Stop()
}

func (s *PrivateServer) update() error {
	if entries, err := s.dataStore.Read(); err != nil {
		return err
	} else {
		announcements := make([]*PrivateAnnouncement, 0, len(entries))
		for _, entry := range entries {
			switch entry.Type {
			case PrivateAnnouncementType:
				eps.Log.Info(string(entry.Data))
				announcement := &PrivateAnnouncement{}
				if err := json.Unmarshal(entry.Data, &announcement); err != nil {
					return fmt.Errorf("invalid record format!")
				}
				announcements = append(announcements, announcement)
			default:
				return fmt.Errorf("unknown entry type found...")
			}
		}
		validAnnouncements := make([]*PrivateAnnouncement, 0)
		for _, announcement := range announcements {
			found := false
			for _, validAnnouncement := range validAnnouncements {
				if announcement.Domain == validAnnouncement.Domain && announcement.Proxy == validAnnouncement.Proxy {
					// we update an existing announcement
					validAnnouncement.ExpiresAt = announcement.ExpiresAt
					found = true
					break
				}
			}
			if !found {
				validAnnouncements = append(validAnnouncements, &PrivateAnnouncement{
					Domain:    announcement.Domain,
					Proxy:     announcement.Proxy,
					ExpiresAt: announcement.ExpiresAt,
				})
			}
		}
		s.announcements = validAnnouncements
		return nil
	}
}
