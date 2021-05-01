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

/*
Message flow through the system:

- We initialize a message broker and a message store
- Messages are passed through channels
- A message can e.g. come into the broker via the `Deliver` method.
- Depending on the message type (synchronous, asynchronous) the `Deliver` call
  will directly return a message response or just put the message into the system.
- When receiving a message, the broker goes through all the channels and asks them
  if they can handle the message. If a channel replies with yes, it asks it
  whether it can deliver the message now. If yes, it calls the `Deliver` function
  of the channel. Otherwise, if the message is synchronous it returns an error.
  If the message is asynchronous, it stores it in the MessageStore and schedules
  it for redelivery later.
*/

type MessageBroker interface {
	MessageStore() MessageStore
	SetMessageStore(MessageStore) error
	AddChannel(Channel) error
	RemoveChannel(Channel) error
	Channels() []Channel
	Deliver(*Message) (*Message, error)
}
