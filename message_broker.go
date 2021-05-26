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

package eps

import (
	"fmt"
	"sync"
)

type MessageBroker interface {
	AddChannel(Channel) error
	Channels() []Channel
	DeliverRequest(*Request, *ClientInfo) (*Response, error)
}

type BasicMessageBroker struct {
	channels          []Channel
	directory         Directory
	mutex             sync.Mutex
	requestsInTransit map[string]bool
}

func MakeBasicMessageBroker(directory Directory) (*BasicMessageBroker, error) {
	return &BasicMessageBroker{
		channels:          make([]Channel, 0),
		requestsInTransit: make(map[string]bool),
		directory:         directory,
	}, nil
}

func (b *BasicMessageBroker) AddChannel(channel Channel) error {
	b.channels = append(b.channels, channel)
	// we tell the channel that it's part of the message broker
	if err := channel.SetMessageBroker(b); err != nil {
		b.channels = b.channels[:len(b.channels)-1]
		return err
	}
	return nil
}

func (b *BasicMessageBroker) DeliverRequest(request *Request, clientInfo *ClientInfo) (*Response, error) {

	b.mutex.Lock()

	if inTransit, ok := b.requestsInTransit[request.ID]; ok && inTransit {
		b.mutex.Unlock()
		return nil, fmt.Errorf("request %s is already being processed (maybe a delivery loop)", request.ID)
	} else {
		b.requestsInTransit[request.ID] = true
		defer func() {
			b.mutex.Lock()
			delete(b.requestsInTransit, request.ID)
			b.mutex.Unlock()
		}()
	}

	b.mutex.Unlock()

	if clientInfo == nil {
		return nil, fmt.Errorf("client info missing")
	}

	Log.Debug("Checking request details...")

	var ownEntry, remoteEntry *DirectoryEntry
	var err error

	// we always update the directory entry of the client info struct
	if remoteEntry, err = b.directory.EntryFor(clientInfo.Name); err != nil {
		return nil, err
	} else {
		clientInfo.Entry = remoteEntry
	}

	if ownEntry, err = b.directory.OwnEntry(); err != nil {
		return nil, err
	}

	// we always add the client information to the request
	if request.Params != nil {

		if clientInfoStruct, err := clientInfo.AsStruct(); err != nil {
			return nil, err
		} else {
			request.Params["_client"] = clientInfoStruct
		}
	}

	address, err := GetAddress(request.ID)

	if err != nil {
		return nil, err
	}

	// if the remote entry isn't identical to the local one we check if the
	// remote endpoint actually has the right to call the given service on
	// this endpoint
	if ownEntry.Name != remoteEntry.Name {
		if !CanCall(remoteEntry, ownEntry, address.Method) {
			Log.Warningf("Permission denied for method '%s' and client '%s'", address.Method, clientInfo.Name)
			return PermissionDenied(&request.ID, nil), nil
		}
	}

	Log.Debug("Checking channels...")

	// To do: Check if a client can actually call the service method of the
	// given operator, reject the request if that's not the case.

	for i, channel := range b.channels {
		Log.Debugf("Checking whether channel %d can deliver message with method '%s' to '%s'...", i, address.Method, address.Operator)
		if !channel.CanDeliverTo(address) {
			continue
		}
		Log.Debug("Trying to deliver message...")
		if response, err := channel.DeliverRequest(request); err != nil {
			Log.Errorf("Channel %d encountered an error delivering message", i)
			continue
		} else {
			return response, nil
		}
	}

	Log.Debug("Done checking channels...")

	return nil, fmt.Errorf("no channel can deliver this request")
}

func (b *BasicMessageBroker) Channels() []Channel {
	return b.channels
}
