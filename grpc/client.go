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

package grpc

import (
	"context"
	"fmt"
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/protobuf"
	"github.com/iris-gateway/eps/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"time"
)

type Client struct {
	connection *grpc.ClientConn
	settings   *GRPCClientSettings
}

func (c *Client) Connect(address, serverName string) error {
	var err error
	var opts []grpc.DialOption

	tlsConfig, err := tls.TLSClientConfig(c.settings.TLS, serverName)

	if err != nil {
		return err
	}

	opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))

	c.connection, err = grpc.Dial(address, opts...)

	if err != nil {
		return err
	}

	return nil

}

func (c *Client) Close() error {
	return c.connection.Close()
}

func (c *Client) SendMessage(message *eps.Message) error {

	client := protobuf.NewEPSClient(c.connection)

	pbMessage := &protobuf.Message{}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	streamClient, err := client.Stream(ctx)

	if err != nil {
		return err
	}

	peer, ok := peer.FromContext(streamClient.Context())
	if ok {
		tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
		v := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
		fmt.Printf("%v - %v\n", peer.Addr.String(), v)
	}

	return streamClient.Send(pbMessage)

}

func MakeClient(settings *GRPCClientSettings) (*Client, error) {

	return &Client{
		settings: settings,
	}, nil

}
