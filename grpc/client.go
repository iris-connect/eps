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
	"github.com/iris-gateway/eps"
	"github.com/iris-gateway/eps/protobuf"
	"github.com/iris-gateway/eps/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/structpb"
	"io"
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

func (c *Client) ServerCall(messageBroker eps.MessageBroker, stop chan bool) error {

	client := protobuf.NewEPSClient(c.connection)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := client.ServerCall(ctx)

	if err != nil {
		eps.Log.Error("Setup:", err)
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
			eps.Log.Error("Call err:", err)
			return err
		}

		request := &eps.Request{
			ID:     pbRequest.Id,
			Params: pbRequest.Params.AsMap(),
			Method: pbRequest.Method,
		}

		response, err := messageBroker.DeliverRequest(request)

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

func MakeClient(settings *GRPCClientSettings) (*Client, error) {

	return &Client{
		settings: settings,
	}, nil

}
