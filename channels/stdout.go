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

// This channel is only for debugging and testing purposes. It prints messages
// to stdout.

package channels

import (
	"github.com/iris-gateway/eps"
	"github.com/kiprotect/go-helpers/forms"
)

type StdoutSettings struct {
}

type StdoutChannel struct {
	eps.BaseChannel
	Settings StdoutSettings
}

var StdoutSettingsForm = forms.Form{
	Fields: []forms.Field{},
}

func StdoutSettingsValidator(settings map[string]interface{}) (interface{}, error) {
	if params, err := StdoutSettingsForm.Validate(settings); err != nil {
		return nil, err
	} else {
		validatedSettings := &StdoutSettings{}
		if err := StdoutSettingsForm.Coerce(validatedSettings, params); err != nil {
			return nil, err
		}
		return validatedSettings, nil
	}
}

func MakeStdoutChannel(broker eps.MessageBroker, settings interface{}) (eps.Channel, error) {
	return &StdoutChannel{
		BaseChannel: eps.BaseChannel{Broker: broker},
		Settings:    settings.(StdoutSettings),
	}, nil
}

func (c *StdoutChannel) Open() error {
	return nil
}

func (c *StdoutChannel) Close() error {
	return nil
}

func (c *StdoutChannel) Deliver(message *eps.Message) (*eps.Message, error) {
	return nil, nil
}

func (c *StdoutChannel) CanDeliver(message *eps.Message) bool {
	return false
}
func (c *StdoutChannel) CanHandle(message *eps.Message) bool {
	return false
}
