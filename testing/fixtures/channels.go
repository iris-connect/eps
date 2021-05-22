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

package fixtures

import (
	"fmt"
	"github.com/iris-connect/eps"
	"github.com/iris-connect/eps/helpers"
)

type Channels struct {
}

func (c Channels) Setup(fixtures map[string]interface{}) (interface{}, error) {
	settings, ok := fixtures["settings"].(*eps.Settings)

	if !ok {
		return nil, fmt.Errorf("settings missing")
	}

	broker, ok := fixtures["broker"].(eps.MessageBroker)

	if !ok {
		return nil, fmt.Errorf("message broker missing")
	}

	directory, ok := fixtures["directory"].(eps.Directory)

	if !ok {
		return nil, fmt.Errorf("directory missing")
	}

	return helpers.InitializeChannels(broker, directory, settings)
}

func (c Channels) Teardown(fixture interface{}) error {
	return nil
}
