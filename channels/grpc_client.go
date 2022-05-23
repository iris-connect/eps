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

package channels

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/iris-connect/eps"
	"github.com/iris-connect/eps/grpc"
	"github.com/iris-connect/eps/helpers"
	"github.com/kiprotect/go-helpers/forms"
	"net"
	"sync"
	"time"
)

type GRPCClientChannel struct {
	eps.BaseChannel
	Settings    grpc.GRPCClientSettings
	connections map[string]*GRPCServerConnection
	stop        chan bool
	mutex       sync.Mutex
}

type GRPCServerConnection struct {
	Name               string
	Address            string
	Stale              bool
	establishedName    string
	establishedAddress string
	client             *grpc.Client
	channel            *GRPCClientChannel
	connected          bool
	connecting         bool
	mutex              sync.Mutex
	stop               chan bool
}

func (c *GRPCServerConnection) Open() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.connected {
		if c.establishedAddress == c.Address && c.establishedName == c.Name {
			eps.Log.Tracef("connection to server '%s' still good...", c.Name)
			// we're already connected and nothing changed
			return nil
		}
		// some connection details changed, we reestablish the connection
		if err := c.Close(); err != nil {
			return fmt.Errorf("error closing connection: %w", err)
		}
	} else if c.connecting {
		return nil
	} else {
		eps.Log.Tracef("opening a new gRPC client connection to server '%s'", c.Name)
	}
	c.connecting = true
	defer func() { c.connecting = false }()
	if client, err := grpc.MakeClient(&c.channel.Settings, nil, c.channel.Directory()); err != nil {
		return fmt.Errorf("error creating gRPC client: %w", err)
	} else if err := client.Connect(c.Address, c.Name); err != nil {
		return fmt.Errorf("error connecting gRPC client: %w", err)
	} else {
		c.client = client
		c.connected = true
		c.establishedName = c.Name
		c.establishedAddress = c.Address
		// we open the server call in another goroutine
		go func() {
			stopping := false
		loop:
			for {
				if !stopping {
					if err := client.ServerCall(c.channel, c.stop); err != nil {
						eps.Log.Errorf("server call failed: %v", err)
						stopping = true
						// there was a connection error, we stop the loop
						go func() {
							if err := c.Close(); err != nil {
								eps.Log.Error(err)
							}
						}()
					} else {
						// the call stopped because it was requested to
						break loop
					}
				}
				select {
				case <-time.After(60 * time.Second):
				case <-c.stop:
					c.stop <- true
					break loop
				}
			}

		}()
		return nil
	}

}

func (c *GRPCServerConnection) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.connected {
		return nil
	}
	if c.client != nil {
		if err := c.client.Close(); err != nil {
			eps.Log.Error(err)
		}
		c.client = nil
	}
	c.stop <- true
	select {
	case <-c.stop:
		c.connected = false
		return nil
	case <-time.After(5 * time.Second):
		c.connected = false
		return fmt.Errorf("timeout when closing channel")
	}
}

type GRPCServerEntrySettings struct {
	Address  string `json:"address"`
	Internal bool   `json:"internal"`
	Proxy    string `json:"proxy"`
}

var GRPCServerEntrySettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "address",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "internal",
			Validators: []forms.Validator{
				forms.IsOptional{Default: false},
				forms.IsBoolean{},
			},
		},
		{
			Name: "proxy",
			Validators: []forms.Validator{
				forms.IsOptional{Default: ""},
				forms.IsString{},
			},
		},
	},
}

func getEntrySettings(settings map[string]interface{}) (*GRPCServerEntrySettings, error) {
	if params, err := GRPCServerEntrySettingsForm.Validate(settings); err != nil {
		return nil, err
	} else {
		validatedSettings := &GRPCServerEntrySettings{}
		if err := GRPCServerEntrySettingsForm.Coerce(validatedSettings, params); err != nil {
			return nil, err
		}
		return validatedSettings, nil
	}
}

func GRPCClientSettingsValidator(settings map[string]interface{}) (interface{}, error) {
	if params, err := grpc.GRPCClientSettingsForm.Validate(settings); err != nil {
		return nil, err
	} else {
		validatedSettings := &grpc.GRPCClientSettings{}
		if err := grpc.GRPCClientSettingsForm.Coerce(validatedSettings, params); err != nil {
			return nil, err
		}
		return validatedSettings, nil
	}
}

func MakeGRPCClientChannel(settings interface{}) (eps.Channel, error) {
	return &GRPCClientChannel{
		Settings:    settings.(grpc.GRPCClientSettings),
		connections: make(map[string]*GRPCServerConnection),
		stop:        make(chan bool),
	}, nil
}

func (c *GRPCClientChannel) Open() error {
	// we start the background task
	go c.backgroundTask()
	if err := c.openConnections(); err != nil {
		eps.Log.Error(err)
	}
	return nil
}

func (c *GRPCClientChannel) Close() error {
	c.stop <- true
	<-c.stop
	return c.closeConnections()
}

func (c *GRPCClientChannel) markConnectionsStale() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, connection := range c.connections {
		connection.Stale = true
	}
}

func (c *GRPCClientChannel) clearStaleConnections() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	var lastErr error
	for key, connection := range c.connections {
		if connection.Stale {
			// this connection is stale, we remove it
			if err := connection.Close(); err != nil {
				eps.Log.Error(err)
				lastErr = err
			}
			eps.Log.Tracef("Removing stale connection with name '%s' and address '%s'...", connection.Name, connection.Address)
			delete(c.connections, key)
		} else {
			eps.Log.Tracef("Keeping connection with name '%s' and address '%s' open...", connection.Name, connection.Address)
		}
	}
	eps.Log.Tracef("%d open gRPC server-client connections in total...", len(c.connections))
	return lastErr
}

func (c *GRPCClientChannel) getConnection(name string) *GRPCServerConnection {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	conn, _ := c.connections[name]
	return conn
}

func (c *GRPCClientChannel) setConnection(name string, conn *GRPCServerConnection) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.connections[name] = conn
}

func (c *GRPCClientChannel) closeConnections() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	var lastErr error
	for _, connection := range c.connections {
		if err := connection.Close(); err != nil {
			lastErr = err
			eps.Log.Error(err)
		}
	}
	return lastErr
}

func (c *GRPCClientChannel) Type() string {
	return "grpc_client"
}

func (c *GRPCClientChannel) backgroundTask() {
	for {
		// we continuously watch for changes in the service directory and
		// adapt our outgoing connections to that...
		select {
		case <-c.stop:
			eps.Log.Debug("Stopping gRPC client background task")
			c.stop <- true
			return
		case <-time.After(60 * time.Second):
			if err := c.openConnections(); err != nil {
				eps.Log.Error(err)
			}
		}
	}
}

func (c *GRPCClientChannel) openConnections() error {
	if entries, err := c.Directory().Entries(&eps.DirectoryQuery{
		Channels: []string{"grpc_server"},
	}); err != nil {
		return fmt.Errorf("error retrieving directory entries: %w", err)
	} else if ownEntry, err := c.Directory().OwnEntry(); err != nil {
		return fmt.Errorf("error retrieving own entry: %w", err)
	} else {
		// we only connect to entries that can actually call services on this
		// endpoint (incoming requests only, as outgoing ones go through the
		// regular client channel and do not require an open connection)
		peerEntries := eps.GetPeers(ownEntry, entries, true)
		// we mark all connections as stale
		c.markConnectionsStale()
		for _, entry := range peerEntries {
			// we skip this connection if the other endpoint has a gRPC client of its own and the current
			// endpoint has a gRPC server, as the other endpoint can then just connect to this gRPC
			// server via its own client...
			if ownEntry.Channel("grpc_server") != nil && entry.Channel("grpc_client") != nil {
				if settings, err := getEntrySettings(ownEntry.Channel("grpc_server").Settings); err != nil {
					eps.Log.Trace(err)
					continue
				} else if settings.Internal == false && settings.Proxy == "" {
					eps.Log.Debugf("Skipping gRPC client connection from '%s' to '%s' as the latter has a gRPC client and the former a publicly available gRPC server", ownEntry.Name, entry.Name)
					continue
				}
			}

			if channel := entry.Channel("grpc_server"); channel == nil {
				return fmt.Errorf("this should not happen: no grpc_server channel found")
			} else if settings, err := getEntrySettings(channel.Settings); err != nil {
				return fmt.Errorf("error retrieving entry settings: %w", err)
			} else {
				if c.Directory().Name() == entry.Name {
					// we won't open a channel to ourselves...
					continue
				}
				if settings.Internal == true || settings.Proxy != "" {
					// we won't try to connect to an internal gRPC server or a gRPC server
					// only reachable through a proxy...
					continue
				}
				eps.Log.Tracef("Maintaining connection to %s at %s", entry.Name, settings.Address)
				if err := c.openConnection(settings.Address, entry.Name); err != nil {
					// we only log this as tracing errors
					eps.Log.Trace(err)
				}
			}
		}
		return c.clearStaleConnections()
	}
}

func (c *GRPCClientChannel) openConnection(address, name string) error {

	eps.Log.Tracef("Opening gRPC client connection to name '%s' and address '%s'...", name, address)

	conn := c.getConnection(name)

	if conn == nil {
		conn = &GRPCServerConnection{
			channel: c,
			stop:    make(chan bool),
		}
	}

	conn.Address = address
	conn.Name = name
	conn.Stale = false

	c.setConnection(name, conn)

	return conn.Open()

}

func (c *GRPCClientChannel) HandleRequest(request *eps.Request, clientInfo *eps.ClientInfo) (*eps.Response, error) {
	return c.MessageBroker().DeliverRequest(request, clientInfo)
}

type RequestConnectionResponse struct {
	Token    []byte `json:"token"`
	Endpoint string `json:"endpoint"`
}

var RequestConnectionResponseForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "token",
			Validators: []forms.Validator{
				forms.IsBytes{Encoding: "base64"},
			},
		},
		{
			Name: "endpoint",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
	},
}

func (c *GRPCClientChannel) DeliverRequest(request *eps.Request) (*eps.Response, error) {

	address, err := eps.GetAddress(request.ID)

	if err != nil {
		return nil, fmt.Errorf("error parsing address: %w", err)
	}

	entry, err := c.DirectoryEntry(address, "grpc_server")

	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, fmt.Errorf("cannot deliver gRPC request: recipient does not have a gRPC server")
	}

	if len(entry.Channels) == 0 {
		return nil, fmt.Errorf("cannot find channel")
	}

	settings, err := getEntrySettings(entry.Channel("grpc_server").Settings)

	if err != nil {
		return nil, fmt.Errorf("error retrieving entry settings: %w", err)
	}

	var dialer grpc.Dialer

	if settings.Proxy != "" {

		eps.Log.Tracef("Destination is only reachable via proxy '%s'...", settings.Proxy)

		if !c.Settings.UseProxy {
			return nil, fmt.Errorf("destination is only reachable via proxy but proxying is disabled")
		}

		dialer = func(context context.Context, addr string) (net.Conn, error) {
			eps.Log.Tracef("Dialing operator '%s' through proxy...", address.Operator)

			// this request comes from ourselves
			clientInfo := &eps.ClientInfo{
				Name: c.Directory().Name(),
			}

			if entry, err := c.Directory().OwnEntry(); err != nil {
				return nil, err
			} else {
				clientInfo.Entry = entry
			}

			method := fmt.Sprintf("%s.requestConnection", settings.Proxy)

			// does not need to be secure just unique for this EPS server...
			id, err := helpers.RandomID(8)

			if err != nil {
				return nil, err
			}

			request := &eps.Request{
				Method: method,
				ID:     fmt.Sprintf("%s(%s)", method, hex.EncodeToString(id)),
				Params: map[string]interface{}{
					"to":      address.Operator,
					"channel": "grpc_server",
				},
			}

			if response, err := c.MessageBroker().DeliverRequest(request, clientInfo); err != nil {
				return nil, err
			} else if response.Error != nil {
				return nil, fmt.Errorf(response.Error.Message)
			} else if requestConnectionResponse, err := parseRequestConnectionResponse(response.Result); err != nil {
				return nil, err
			} else {

				proxyConnection, err := net.Dial("tcp", requestConnectionResponse.Endpoint)

				if err != nil {
					return nil, err
				}

				if n, err := proxyConnection.Write(requestConnectionResponse.Token); err != nil {
					proxyConnection.Close()
					return nil, err
				} else if n != len(requestConnectionResponse.Token) {
					proxyConnection.Close()
					return nil, fmt.Errorf("could not write token")
				}

				eps.Log.Infof("Successfully established gRPC client connection to proxy %s", requestConnectionResponse.Endpoint)

				return proxyConnection, nil
			}
		}
	}

	if client, err := grpc.MakeClient(&c.Settings, dialer, c.Directory()); err != nil {
		return nil, fmt.Errorf("error creating gRPC client: %w", err)
	} else if err := client.Connect(settings.Address, entry.Name); err != nil {
		return nil, fmt.Errorf("error connecting gRPC client: %w", err)
	} else {

		// we ensure the client will always be closed
		defer func() {
			if err := client.Close(); err != nil {
				eps.Log.Error(err)
			}
		}()

		response, err := client.SendRequest(request)
		if err != nil {
			return nil, fmt.Errorf("error sending request: %w", err)
		}

		return response, nil

	}
}

func (c *GRPCClientChannel) CanDeliverTo(address *eps.Address) bool {

	// we'll never deliver to ourselves...
	if address.Operator == c.Directory().Name() {
		return false
	}

	// we check if the requested service offers a gRPC server
	if entry, err := c.DirectoryEntry(address, "grpc_server"); entry != nil {
		if settings, err := getEntrySettings(entry.Channel("grpc_server").Settings); err != nil {
			eps.Log.Error(err)
			return false
		} else if settings.Internal {
			return false
		} else if settings.Proxy == c.Directory().Name() {
			return false
		}
		return true
	} else if err != nil {
		// we log this error
		eps.Log.Error(err)
	}

	return false
}
