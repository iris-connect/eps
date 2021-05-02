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

import ()

type ChannelDefinition struct {
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	Maker             ChannelMaker      `json:"-"`
	SettingsValidator SettingsValidator `json:"-"`
}

type ChannelDefinitions map[string]ChannelDefinition
type ChannelMaker func(settings interface{}) (Channel, error)

// A channel can deliver and accept message
type Channel interface {
	MessageBroker() MessageBroker
	SetMessageBroker(MessageBroker) error
	CanDeliverTo(*Address) bool
	DeliverRequest(*Request) (*Response, error)
	DeliverResponse(*Response) error
	SetDirectory(Directory) error
	Directory() Directory
	Close() error
	Open() error
}

type BaseChannel struct {
	broker    MessageBroker
	directory Directory
}

func (b *BaseChannel) GetDirectoryEntry(address *Address) *DirectoryEntry {
	entries := b.Directory().Entries(&DirectoryQuery{
		Operator: address.Operator,
		Channels: []string{"grpc_client"},
	})

	if len(entries) > 0 {
		return entries[0]
	}

	return nil

}

func (b *BaseChannel) Directory() Directory {
	return b.directory
}

func (b *BaseChannel) SetDirectory(directory Directory) error {
	b.directory = directory
	return nil
}

func (b *BaseChannel) MessageBroker() MessageBroker {
	return b.broker
}

func (b *BaseChannel) SetMessageBroker(broker MessageBroker) error {
	b.broker = broker
	return nil
}
