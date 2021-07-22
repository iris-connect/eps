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
	"fmt"
	"github.com/iris-connect/eps"
	"github.com/iris-connect/eps/grpc"
	"github.com/kiprotect/go-helpers/forms"
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
			eps.Log.Tracef("Connection still good, doing nothing...")
			// we're already connected and nothing changed
			return nil
		}
		eps.Log.Tracef("Connection details changed!")
		// some connection details changed, we reestablish the connection
		if err := c.Close(); err != nil {
			return err
		}
	} else if c.connecting {
		return nil
	} else {
		eps.Log.Tracef("Opening a new gRPC client connection")
	}
	c.connecting = true
	defer func() { c.connecting = false }()
	if client, err := grpc.MakeClient(&c.channel.Settings, c.channel.Directory()); err != nil {
		return err
	} else if err := client.Connect(c.Address, c.Name); err != nil {
		return err
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
						eps.Log.Errorf("Server call failed: %v", err)
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
				// in case of an error we try to reconnect after 1 second
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
			delete(c.connections, key)
		}
	}
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
		return err
	} else if ownEntry, err := c.Directory().OwnEntry(); err != nil {
		return err
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
				} else if settings.Internal == false {
					eps.Log.Debugf("Skipping gRPC client connection from '%s' to '%s' as the latter has a gRPC client and the former a gRPC server", ownEntry.Name, entry.Name)
					continue
				}
			}

			if channel := entry.Channel("grpc_server"); channel == nil {
				return fmt.Errorf("this should not happen: no grpc_server channel found")
			} else if settings, err := getEntrySettings(channel.Settings); err != nil {
				return err
			} else {
				if c.Directory().Name() == entry.Name {
					// we won't open a channel to ourselves...
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

func (c *GRPCClientChannel) DeliverRequest(request *eps.Request) (*eps.Response, error) {

	address, err := eps.GetAddress(request.ID)

	if err != nil {
		return nil, err
	}

	entry, err := c.DirectoryEntry(address, "grpc_server")

	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, fmt.Errorf("cannot deliver gRPC request")
	}

	if len(entry.Channels) == 0 {
		return nil, fmt.Errorf("cannot find channel")
	}

	settings, err := getEntrySettings(entry.Channel("grpc_server").Settings)

	if err != nil {
		return nil, err
	}

	if client, err := grpc.MakeClient(&c.Settings, c.Directory()); err != nil {
		return nil, err
	} else if err := client.Connect(settings.Address, entry.Name); err != nil {
		return nil, err
	} else {
		response, err := client.SendRequest(request)
		if err != nil {
			return nil, err
		}

		if err := client.Close(); err != nil {
			eps.Log.Error(err)
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
		return true
	} else if err != nil {
		// we log this error
		eps.Log.Error(err)
	}

	return false
}
