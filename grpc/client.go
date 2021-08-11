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
	"crypto/x509"
	"fmt"
	"github.com/iris-connect/eps"
	"github.com/iris-connect/eps/helpers"
	"github.com/iris-connect/eps/protobuf"
	"github.com/iris-connect/eps/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/types/known/structpb"
	"io"
	"net"
	"sync"
	"time"
)

type Client struct {
	directory   eps.Directory
	connection  *grpc.ClientConn
	clientInfos *ClientInfos
	settings    *GRPCClientSettings
	mutex       sync.Mutex
}

type ClientInfos struct {
	Infos []*eps.ClientInfo
}

func (c *ClientInfos) PrimaryName() string {
	if len(c.Infos) == 0 {
		return ""
	}
	return c.Infos[0].Name
}

func (c *ClientInfos) ClientInfo(name string) *eps.ClientInfo {
	for _, info := range c.Infos {
		if info.Name == name || name == "" {
			return info
		}
	}
	return nil
}

func (c *ClientInfos) HasName(name string) bool {
	for _, info := range c.Infos {
		if info.Name == name {
			return true
		}
	}
	return false
}

func MakeClientInfos() *ClientInfos {
	return &ClientInfos{
		Infos: []*eps.ClientInfo{},
	}
}

type VerifyCredentials struct {
	credentials.TransportCredentials
	directory   eps.Directory
	ClientInfos *ClientInfos
}

type ClientInfoAuthInfo struct {
	credentials.AuthInfo
	ClientInfos *ClientInfos
}

func (c *VerifyCredentials) checkFingerprint(cert *x509.Certificate, name string) (*eps.DirectoryEntry, bool, error) {
	if entry, err := c.directory.EntryFor(name); err != nil {
		if err == eps.NoEntryFound {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("error retrieving directory entry for '%s' for fingerprint check: %w", name, err)
	} else {
		// we go through all certificates for the entry
		for _, directoryCert := range entry.Certificates {
			// we make sure the certificate is good for encryption
			if directoryCert.KeyUsage != "encryption" {
				continue
			}
			// we check if this is a valid certificate for this operator
			if helpers.VerifyFingerprint(cert, directoryCert.Fingerprint) {
				return entry, true, nil
			}
		}
		return nil, false, nil
	}
}

func (c *VerifyCredentials) handshake(conn net.Conn, authInfo credentials.AuthInfo, clientInfos *ClientInfos) (net.Conn, credentials.AuthInfo, error) {

	tlsInfo := authInfo.(credentials.TLSInfo)

	if len(tlsInfo.State.PeerCertificates) == 0 {
		return conn, authInfo, fmt.Errorf("certificate missing")
	}

	cert := tlsInfo.State.PeerCertificates[0]

	names := []string{cert.Subject.CommonName}
	names = append(names, cert.DNSNames...)

	if clientInfos == nil {
		// we create a new client info object
		clientInfos = MakeClientInfos()
	} else {
		// we reset the infos
		clientInfos.Infos = []*eps.ClientInfo{}
	}

	for _, name := range names {

		if entry, ok, err := c.checkFingerprint(cert, name); err != nil {
			return conn, authInfo, err
		} else if !ok {
			continue
		} else {
			clientInfo := &eps.ClientInfo{
				Name:  name,
				Entry: entry,
			}
			clientInfos.Infos = append(clientInfos.Infos, clientInfo)
		}
	}

	if len(clientInfos.Infos) == 0 {
		return conn, authInfo, fmt.Errorf("no name matched")
	}

	return conn, &ClientInfoAuthInfo{authInfo, clientInfos}, nil

}

func (c VerifyCredentials) ServerHandshake(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	conn, authInfo, err := c.TransportCredentials.ServerHandshake(conn)

	if err != nil {
		return conn, authInfo, fmt.Errorf("error performing server handshake: %w", err)
	}
	// for the server we do not pass a client info object but create a new
	// one for every handshake, as the client info will change...
	return c.handshake(conn, authInfo, nil)

}

func (c VerifyCredentials) ClientHandshake(ctx context.Context, endpoint string, conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	conn, authInfo, err := c.TransportCredentials.ClientHandshake(ctx, endpoint, conn)

	if err != nil {
		return conn, authInfo, fmt.Errorf("error performing client handshake: %w", err)
	}
	// for the client we pass the existing client info object as it will be
	// used only once...
	return c.handshake(conn, authInfo, c.ClientInfos)
}

func (c *Client) Connect(address, serverName string) error {

	c.mutex.Lock()
	defer c.mutex.Unlock()

	var err error
	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             20 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	tlsConfig, err := tls.TLSClientConfig(c.settings.TLS)

	if err != nil {
		return fmt.Errorf("error retrieving gRPC client TLS config: %w", err)
	}

	tlsConfig.ServerName = serverName

	c.clientInfos = MakeClientInfos()

	vc := &VerifyCredentials{directory: c.directory, ClientInfos: c.clientInfos, TransportCredentials: credentials.NewTLS(tlsConfig)}
	opts = append(opts, grpc.WithTransportCredentials(vc))

	c.connection, err = grpc.Dial(address, opts...)

	return err

}

func (c *Client) Close() error {

	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.connection.Close()
	c.connection = nil
	c.clientInfos = nil
	return err
}

func (c *Client) ServerCall(handler Handler, stop chan bool) error {

	client := protobuf.NewEPSClient(c.connection)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.ServerCall(ctx)

	if err != nil {
		return fmt.Errorf("error performing server call: %w", err)
	}

	announcementStruct, err := structpb.NewStruct(map[string]interface{}{"name": c.directory.Name()})

	if err != nil {
		return fmt.Errorf("error serializing directory entry for gRPC: %w", err)
	}

	pbResponse := &protobuf.Response{
		Id: "1",
	}

	pbResponse.Result = announcementStruct

	// we announce the client name to the server
	if err := stream.Send(pbResponse); err != nil {
		eps.Log.Error(err)
	}

	for {

		done := make(chan bool, 1)

		var pbRequest *protobuf.Request
		var err error

		go func() {
			pbRequest, err = stream.Recv()
			done <- true
		}()

		select {
		case <-stop:
			// we were asked to stop
			stop <- true
			return nil
		case <-done:
		}

		if err == io.EOF {
			continue
		}

		if err != nil {
			return fmt.Errorf("error receiving gRPC request: %w", err)
		}

		request := &eps.Request{
			ID:     pbRequest.Id,
			Params: pbRequest.Params.AsMap(),
			Method: pbRequest.Method,
		}

		clientInfo := c.clientInfos.ClientInfo(pbRequest.ClientName)

		if clientInfo == nil {

			pbResponse := &protobuf.Response{
				Id: pbRequest.Id,
			}

			pbResponse.Error = &protobuf.Error{
				Code:    404,
				Message: "no matching client found",
			}

			if err := stream.Send(pbResponse); err != nil {
				eps.Log.Error(err)
			}

			continue
		}

		response, err := handler.HandleRequest(request, clientInfo)

		pbResponse := &protobuf.Response{
			Id: pbRequest.Id,
		}

		if err != nil {
			pbResponse.Error = &protobuf.Error{
				Code:    -100,
				Message: err.Error(),
			}
		} else if response != nil {
			if response.Result != nil {
				resultStruct, err := structpb.NewStruct(response.Result)
				if err != nil {
					eps.Log.Error(err)
				}
				pbResponse.Result = resultStruct
			}
			if response.Error != nil {
				pbResponse.Error = &protobuf.Error{
					Code:    int32(response.Error.Code),
					Message: response.Error.Message,
				}

				if response.Error.Data != nil {
					errorStruct, err := structpb.NewStruct(response.Error.Data)
					if err != nil {
						eps.Log.Error(err)
					}
					pbResponse.Error.Data = errorStruct
				}
			}
		}

		if err := stream.Send(pbResponse); err != nil {
			eps.Log.Error(err)
		}

	}

}

func (c *Client) SendRequest(request *eps.Request) (*eps.Response, error) {

	client := protobuf.NewEPSClient(c.connection)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	paramsStruct, err := structpb.NewStruct(request.Params)

	if err != nil {
		eps.Log.Error(err)
		return nil, fmt.Errorf("error serializing params for gRPC: %w", err)
	}

	pbRequest := &protobuf.Request{
		ClientName: c.directory.Name(),
		Params:     paramsStruct,
		Method:     request.Method,
		Id:         request.ID,
	}

	pbResponse, err := client.Call(ctx, pbRequest)

	if err != nil {
		eps.Log.Error(err)
		return nil, fmt.Errorf("error performing gRPC call: %w", err)
	}

	var responseError *eps.Error

	if pbResponse.Error != nil {
		responseError = &eps.Error{
			Code:    int(pbResponse.Error.Code),
			Data:    pbResponse.Error.Data.AsMap(),
			Message: pbResponse.Error.Message,
		}
	}

	response := &eps.Response{
		Result: pbResponse.Result.AsMap(),
		ID:     &pbResponse.Id,
		Error:  responseError,
	}

	return response, nil

}

func MakeClient(settings *GRPCClientSettings, directory eps.Directory) (*Client, error) {

	return &Client{
		settings:  settings,
		directory: directory,
	}, nil

}
