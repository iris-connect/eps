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
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/structpb"
	"io"
)

type EPSServer struct {
	protobuf.UnimplementedEPSServer
	handler Handler
}

type Handler func(*eps.Request) (*eps.Response, error)

func (s *EPSServer) Call(context context.Context, pbRequest *protobuf.Request) (*protobuf.Response, error) {

	peer, ok := peer.FromContext(context)
	if ok {
		tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
		v := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
		fmt.Printf("%v - %v\n", peer.Addr.String(), v)
	}

	request := &eps.Request{
		ID:     pbRequest.Id,
		Params: pbRequest.Params.AsMap(),
		Method: pbRequest.Method,
	}

	if response, err := s.handler(request); err != nil {
		return nil, err
	} else {

		pbResponse := &protobuf.Response{
			Id: pbRequest.Id,
		}
		if response.Result != nil {
			resultStruct, err := structpb.NewStruct(response.Result)
			if err != nil {
				return nil, err
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
					return nil, err
				}
				pbResponse.Error.Data = errorStruct
			}

		}

		return pbResponse, nil

	}

}

func (s *EPSServer) Stream(stream protobuf.EPS_AsyncUpstreamServer) error {

	// this is a bidirectional message stream

	peer, ok := peer.FromContext(stream.Context())
	if ok {
		tlsInfo := peer.AuthInfo.(credentials.TLSInfo)
		v := tlsInfo.State.VerifiedChains[0][0].Subject.CommonName
		fmt.Printf("%v - %v\n", peer.Addr.String(), v)
	}

	for {
		_, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func MakeEPSServer(handler Handler) *EPSServer {
	return &EPSServer{
		handler: handler,
	}
}
