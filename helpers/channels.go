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

package helpers

import (
	"fmt"
	"github.com/iris-gateway/eps"
)

func GetChannelSettingsAndDefinition(settings *eps.Settings, name string) (*eps.ChannelSettings, *eps.ChannelDefinition, error) {
	for _, channel := range settings.Channels {
		if channel.Name == name {
			def := settings.Definitions.ChannelDefinitions[channel.Type]
			return channel, &def, nil
		}
	}
	return nil, nil, fmt.Errorf("channel not found")
}

func InitializeChannels(broker eps.MessageBroker, directory eps.Directory, settings *eps.Settings) ([]eps.Channel, error) {
	channels := make([]eps.Channel, 0)
	for _, channel := range settings.Channels {
		eps.Log.Debugf("Initializing channel '%s' of type '%s'", channel.Name, channel.Type)
		definition := settings.Definitions.ChannelDefinitions[channel.Type]
		if channel, err := definition.Maker(channel.Settings); err != nil {
			return nil, err
		} else {
			if err := broker.AddChannel(channel); err != nil {
				return nil, err
			}
			if err := channel.SetDirectory(directory); err != nil {
				return nil, err
			}
			channels = append(channels, channel)
		}
	}
	return channels, nil
}
