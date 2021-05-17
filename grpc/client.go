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
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/helpers"
	"github.com/iris-gateway/eps/protobuf"
	"github.com/iris-gateway/eps/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/structpb"
	"io"
	"net"
	"sync"
)

type Client struct {
	directory  eps.Directory
	connection *grpc.ClientConn
	clientInfo *eps.ClientInfo
	settings   *GRPCClientSettings
	mutex      sync.Mutex
}

type VerifyCredentials struct {
	credentials.TransportCredentials
	directory  eps.Directory
	ClientInfo *eps.ClientInfo
}

type ClientInfoAuthInfo struct {
	credentials.AuthInfo
	ClientInfo *eps.ClientInfo
}

func (c VerifyCredentials) checkFingerprint(cert *x509.Certificate, name string, clientInfo *eps.ClientInfo) (bool, error) {
	if entry, err := c.directory.EntryFor(name); err != nil {
		eps.Log.Error("can't verify entry...")
		return false, err
	} else {
		clientInfo.Entry = entry
		// we go through all certificates for the entry
		for _, directoryCert := range entry.Certificates {
			// we make sure the certificate is good for encryption
			if directoryCert.KeyUsage != "encryption" {
				continue
			}
			// we check if this is a valid certificate for this operator
			if helpers.VerifyFingerprint(cert, directoryCert.Fingerprint) {
				return true, nil
			}
		}
		return false, nil
	}
}

func (c VerifyCredentials) handshake(conn net.Conn, authInfo credentials.AuthInfo) (net.Conn, credentials.AuthInfo, error) {

	tlsInfo := authInfo.(credentials.TLSInfo)
	cert := tlsInfo.State.PeerCertificates[0]
	name := cert.Subject.CommonName

	if ok, err := c.checkFingerprint(cert, name, c.ClientInfo); err != nil {
		return conn, authInfo, err
	} else if !ok {
		return conn, authInfo, fmt.Errorf("invalid certificate")
	}

	c.ClientInfo.Name = name
	clientInfoAuthInfo := &ClientInfoAuthInfo{authInfo, c.ClientInfo}

	return conn, clientInfoAuthInfo, nil

}

func (c VerifyCredentials) ServerHandshake(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	conn, authInfo, err := c.TransportCredentials.ServerHandshake(conn)

	if err != nil {
		return conn, authInfo, err
	}

	return c.handshake(conn, authInfo)

}

func (c VerifyCredentials) ClientHandshake(ctx context.Context, endpoint string, conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	conn, authInfo, err := c.TransportCredentials.ClientHandshake(ctx, endpoint, conn)

	if err != nil {
		return conn, authInfo, err
	}

	return c.handshake(conn, authInfo)
}

func (c *Client) Connect(address, serverName string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var err error
	var opts []grpc.DialOption

	tlsConfig, err := tls.TLSClientConfig(c.settings.TLS, serverName)

	if err != nil {
		return err
	}

	vc := &VerifyCredentials{directory: c.directory, ClientInfo: &eps.ClientInfo{}, TransportCredentials: credentials.NewTLS(tlsConfig)}
	opts = append(opts, grpc.WithTransportCredentials(vc))

	c.connection, err = grpc.Dial(address, opts...)

	if err != nil {
		return err
	}

	c.clientInfo = vc.ClientInfo

	return nil

}

func (c *Client) Close() error {

	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.connection.Close()
	c.connection = nil
	c.clientInfo = nil
	return err
}

func (c *Client) ServerCall(handler Handler, stop chan bool) error {

	client := protobuf.NewEPSClient(c.connection)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.ServerCall(ctx)

	if err != nil {
		return err
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
			return err
		}

		request := &eps.Request{
			ID:     pbRequest.Id,
			Params: pbRequest.Params.AsMap(),
			Method: pbRequest.Method,
		}

		response, err := handler.HandleRequest(request, c.clientInfo)

		pbResponse := &protobuf.Response{
			Id: pbRequest.Id,
		}

		if err != nil {
			pbResponse.Error = &protobuf.Error{
				Code:    -100,
				Message: err.Error(),
			}
		} else {
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
		return nil, err
	}

	pbRequest := &protobuf.Request{
		Params: paramsStruct,
		Method: request.Method,
		Id:     request.ID,
	}

	pbResponse, err := client.Call(ctx, pbRequest)

	if err != nil {
		eps.Log.Error(err)
		return nil, err
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
