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
)

type MessageBroker interface {
	AddChannel(Channel) error
	Channels() []Channel
	DeliverRequest(*Request) (*Response, error)
}

type BasicMessageBroker struct {
	channels []Channel
}

func MakeBasicMessageBroker() (*BasicMessageBroker, error) {
	return &BasicMessageBroker{
		channels: make([]Channel, 0),
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

func (b *BasicMessageBroker) DeliverRequest(request *Request) (*Response, error) {

	Log.Info(request.ID)

	address, err := GetAddress(request.ID)

	if err != nil {
		Log.Error("addressing errro...")
		return nil, err
	}

	for _, channel := range b.channels {
		if !channel.CanDeliverTo(address) {
			continue
		}
		Log.Infof("found channel that can deliver!")
		return channel.DeliverRequest(request)
	}
	return nil, fmt.Errorf("no channel can deliver this request")
}

func (b *BasicMessageBroker) Channels() []Channel {
	return b.channels
}
