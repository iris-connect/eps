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
	"encoding/json"
	"fmt"
	"github.com/iris-gateway/eps"
	epsForms "github.com/iris-gateway/eps/forms"
	"github.com/iris-gateway/eps/helpers"
	"github.com/iris-gateway/eps/jsonrpc"
	"github.com/kiprotect/go-helpers/forms"
	"net"
	"strings"
	"sync"
	"time"
)

type PrivateServer struct {
	dataStore     helpers.DataStore
	settings      *PrivateServerSettings
	announcements []*PrivateAnnouncement
	jsonrpcServer *jsonrpc.JSONRPCServer
	jsonrpcClient *jsonrpc.Client
	l             net.Listener
	mutex         sync.Mutex
}

type ProxyConnection struct {
	proxyEndpoint    string
	internalEndpoint string
	token            []byte
}

func MakeProxyConnection(proxyEndpoint, internalEndpoint string, token []byte) *ProxyConnection {
	return &ProxyConnection{
		proxyEndpoint:    proxyEndpoint,
		internalEndpoint: internalEndpoint,
		token:            token,
	}
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

	internalConnection, err := net.Dial("tcp", p.internalEndpoint)

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
	return context.Result(c.announcements)
}

type IncomingConnectionParams struct {
	Endpoint string          `json:"endpoint"`
	Token    []byte          `json:"token"`
	Client   *eps.ClientInfo `json:"_client"`
}

func (c *PrivateServer) incomingConnection(context *jsonrpc.Context, params *IncomingConnectionParams) *jsonrpc.Response {

	data, err := json.Marshal(params.Client)

	if err != nil {
		return context.InternalError()
	}
	eps.Log.Info(string(data))
	connection := MakeProxyConnection(params.Endpoint, c.settings.InternalEndpoint, params.Token)

	go func() {
		if err := connection.Run(); err != nil {
			eps.Log.Error(err)
		}
	}()

	return context.Result(map[string]interface{}{"message": "ok"})
}

func MakePrivateServer(settings *PrivateServerSettings) (*PrivateServer, error) {

	server := &PrivateServer{
		settings:      settings,
		dataStore:     helpers.MakeFileDataStore(settings.DatabaseFile),
		jsonrpcClient: jsonrpc.MakeClient(settings.JSONRPCClient),
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
	Proxy      string
}

func (c *PrivateServer) announceConnection(context *jsonrpc.Context, params *PrivateAnnounceConnectionParams) *jsonrpc.Response {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	settings := params.ClientInfo.Entry.SettingsFor("proxy", c.settings.Name)

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

	var newAnnouncement *PrivateAnnouncement
	changed := false
	for _, announcement := range c.announcements {
		if announcement.Domain == params.Domain {
			if announcement.Proxy == params.Proxy {
				return context.Error(400, "already taken", announcement)
			}
			newAnnouncement = announcement
			if newAnnouncement.ExpiresAt != announcement.ExpiresAt || announcement.ExpiresAt != nil && newAnnouncement.ExpiresAt != nil && !params.ExpiresAt.Equal(*announcement.ExpiresAt) {
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

		dataEntry := &helpers.DataEntry{
			Type: PrivateAnnouncementType,
			ID:   id,
			Data: rawData,
		}

		if err := c.dataStore.Write(dataEntry); err != nil {
			eps.Log.Error(err)
			return context.InternalError()
		}

	}

	return context.Acknowledge()
}

func (s *PrivateServer) announceConnectionRPC(domain string, proxy string) error {
	eps.Log.Infof("Sending announcement to %s", proxy)
	request := jsonrpc.MakeRequest(fmt.Sprintf("%s.announceConnection", proxy), "", map[string]interface{}{
		"operator": s.settings.Name,
		"domain":   domain,
	})
	response, err := s.jsonrpcClient.Call(request)
	eps.Log.Info(response.Error)
	return err

}

func (s *PrivateServer) Start() error {
	for _, announcement := range s.announcements {
		if err := s.announceConnectionRPC(announcement.Domain, announcement.Proxy); err != nil {
			eps.Log.Error(err)
		}
	}
	return s.jsonrpcServer.Start()
}

func (s *PrivateServer) Stop() error {
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
			if announcement.Revoked {
				newAnnouncements := make([]*PrivateAnnouncement, 0)
				for _, validAnnouncement := range validAnnouncements {
					if validAnnouncement.Domain == announcement.Domain && validAnnouncement.Proxy == announcement.Proxy {
						continue
					}
					newAnnouncements = append(newAnnouncements, validAnnouncement)
				}
				validAnnouncements = newAnnouncements
			} else {
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
		}
		s.announcements = validAnnouncements
		return nil
	}
}
