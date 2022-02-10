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
	epsNet "github.com/iris-connect/eps/net"
	"net"
	"sync"
	"time"
)

type GRPCServerChannel struct {
	eps.BaseChannel
	server        *grpc.Server
	proxyListener *ProxyListener
	Settings      grpc.GRPCServerSettings
}

func GRPCServerSettingsValidator(settings map[string]interface{}) (interface{}, error) {
	if params, err := grpc.GRPCServerSettingsForm.Validate(settings); err != nil {
		return nil, err
	} else {
		validatedSettings := &grpc.GRPCServerSettings{}
		if err := grpc.GRPCServerSettingsForm.Coerce(validatedSettings, params); err != nil {
			return nil, err
		}
		return validatedSettings, nil
	}
}

func MakeGRPCServerChannel(settings interface{}) (eps.Channel, error) {
	return &GRPCServerChannel{
		Settings: settings.(grpc.GRPCServerSettings),
	}, nil
}

func (c *GRPCServerChannel) HandleConnectionRequest(address *eps.Address, request *eps.Request) (*eps.Response, error) {
	if connectionRequest, err := parseConnectionRequest(request); err != nil {
		return nil, err
	} else if connectionRequest.Channel != c.Type() {
		// this request is not for us
		return nil, nil
	} else {

		proxyConnection, err := net.Dial("tcp", connectionRequest.Endpoint)

		if err != nil {
			return nil, err
		}

		if n, err := proxyConnection.Write(connectionRequest.Token); err != nil {
			proxyConnection.Close()
			return nil, err
		} else if n != len(connectionRequest.Token) {
			proxyConnection.Close()
			return nil, fmt.Errorf("could not write token")
		}
		eps.Log.Tracef("Successfully established gRPC server connection to proxy %s", connectionRequest.Endpoint)

		if err := c.proxyListener.Inject(proxyConnection); err != nil {
			return nil, err
		}

	}
	return &eps.Response{}, nil
}

func (c *GRPCServerChannel) HandleRequest(request *eps.Request, clientInfo *eps.ClientInfo) (*eps.Response, error) {
	return c.MessageBroker().DeliverRequest(request, clientInfo)
}

type ProxyListener struct {
	net.Listener
	listener net.Listener
	channel  chan interface{}
	mutex    sync.Mutex
}

func (l *ProxyListener) Start() {
	go func() {
		for {
			if conn, err := l.Listener.Accept(); err != nil {
				l.Inject(err)
				return
			} else if err := l.Inject(conn); err != nil {
				return
			}
		}
	}()
}

// Accept a connection, ensuring that rate limits are enforced
func (l *ProxyListener) Accept() (net.Conn, error) {
	val := <-l.channel
	switch v := val.(type) {
	case error:
		return nil, v
	case net.Conn:
		return v, nil
	default:
		return nil, fmt.Errorf("unknown type")
	}
}

func (l *ProxyListener) Inject(value interface{}) error {
	select {
	case l.channel <- value:
		return nil
	case <-time.After(2 * time.Second):
		return fmt.Errorf("timeout")
	}
}

func MakeProxyListener(listener net.Listener) *ProxyListener {
	pl := &ProxyListener{
		Listener: listener,
		channel:  make(chan interface{}, 1),
	}
	pl.Start()
	return pl
}

func (c *GRPCServerChannel) Type() string {
	return "grpc_server"
}

func (c *GRPCServerChannel) Open() error {
	var err error

	lis, err := net.Listen("tcp", c.Settings.BindAddress)

	if err != nil {
		return fmt.Errorf("error binding to address '%s': %w", c.Settings.BindAddress, err)
	}

	if c.Settings.TCPRateLimits != nil {
		lis = epsNet.MakeRateLimitedListener(lis, c.Settings.TCPRateLimits)
	}

	c.proxyListener = MakeProxyListener(lis)

	if c.server, err = grpc.MakeServer(&c.Settings, c, c.proxyListener, c.Directory()); err != nil {
		return err
	}
	return c.server.Start()
}

func (c *GRPCServerChannel) Close() error {
	return nil
}

func (c *GRPCServerChannel) DeliverRequest(request *eps.Request) (*eps.Response, error) {
	return c.server.DeliverRequest(request)
}

func (c *GRPCServerChannel) CanDeliverTo(address *eps.Address) bool {

	// we'll never deliver to ourselves...
	if address.Operator == c.Directory().Name() {
		return false
	}

	return c.server.CanDeliverTo(address)
}
