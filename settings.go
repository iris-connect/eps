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

type ChannelSettings struct {
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Services []string    `json:"services"`
	Settings interface{} `json:"settings"`
}

type DirectorySettings struct {
	Type     string      `json:"type"`
	Settings interface{} `json:"settings"`
}

type Settings struct {
	Definitions *Definitions       `json:"definitions"`
	Channels    []*ChannelSettings `json:"channels"`
	Directory   *DirectorySettings `json:"directory"`
	Name        string             `json:"name"`
}

type SettingsValidator func(settings map[string]interface{}) (interface{}, error)
