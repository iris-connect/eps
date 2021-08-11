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
	"github.com/iris-connect/eps"
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
		if channelObj, err := definition.Maker(channel.Settings); err != nil {
			return nil, fmt.Errorf("error initializing channel '%s': %w", channel.Name, err)
		} else {
			if err := broker.AddChannel(channelObj); err != nil {
				return nil, fmt.Errorf("error adding channel '%s': %w", channel.Name, err)
			}
			if err := channelObj.SetDirectory(directory); err != nil {
				return nil, fmt.Errorf("error setting directory for channel '%s': %w", channel.Name, err)
			}
			channels = append(channels, channelObj)
		}
	}
	return channels, nil
}

func OpenChannels(broker eps.MessageBroker, directory eps.Directory, settings *eps.Settings) ([]eps.Channel, error) {

	channels, err := InitializeChannels(broker, directory, settings)

	if err != nil {
		return nil, fmt.Errorf("error initializing channels: %w", err)
	} else {
		for i, channel := range channels {
			if err := channel.Open(); err != nil {
				return nil, fmt.Errorf("error opening channel %d: %w", i, err)
			}
		}
	}
	return channels, nil
}

func CloseChannels(channels []eps.Channel) error {
	var lastErr error
	for i, channel := range channels {
		if err := channel.Close(); err != nil {
			lastErr = fmt.Errorf("error closing channel %d: %w", i, err)
			eps.Log.Error(lastErr)
		}
	}
	return lastErr
}
