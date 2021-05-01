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
type SettingsValidator func(definitions *Definitions, settings map[string]interface{}) (interface{}, error)
type ChannelMaker func(definitions *Definitions, settings interface{}) (Channel, error)

// A channel can deliver and accept message
type Channel interface {
	MessageBroker() MessageBroker
	SetMessageBroker(MessageBroker) error
	CanDeliver(*Message) bool
	CanHandle(*Message) bool
	Deliver(*Message) (*Message, error)
	Close() error
	Open() error
}

type BaseChannel struct {
	messageBroker MessageBroker
}

func (b *BaseChannel) MessageBroker() MessageBroker {
	return b.messageBroker
}

func (b *BaseChannel) SetMessageBroker(messageBroker MessageBroker) error {
	b.messageBroker = messageBroker
	return nil
}
