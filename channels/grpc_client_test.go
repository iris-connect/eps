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

package channels_test

import (
	"github.com/iris-connect/eps"
	"github.com/iris-connect/eps/channels"
	th "github.com/iris-connect/eps/testing"
	"github.com/iris-connect/eps/testing/fixtures"
	"testing"
)

var clientFixtures = []th.FC{
	{fixtures.Settings{Paths: []string{"", "roles/op-1"}}, "settings"},
	{fixtures.Directory{}, "directory"},
	{fixtures.MessageBroker{}, "broker"},
	{fixtures.Channel{"test gRPC client"}, "client"},
}

var serverFixtures = []th.FC{
	{fixtures.Settings{Paths: []string{"", "roles/hd-1"}}, "settings"},
	{fixtures.Directory{}, "directory"},
	{fixtures.MessageBroker{}, "broker"},
	{fixtures.Channels{Open: true}, "channels"},
}

func TestGRPCClientConnection(t *testing.T) {

	cf, err := th.SetupFixtures(clientFixtures)

	if err != nil {
		t.Fatal(err)
	}

	defer th.TeardownFixtures(clientFixtures, cf)

	sf, err := th.SetupFixtures(serverFixtures)

	if err != nil {
		t.Fatal(err)
	}

	defer th.TeardownFixtures(serverFixtures, sf)

	client := cf["client"].(*channels.GRPCClientChannel)

	request := &eps.Request{
		ID: "hd-1.add(1)",
	}

	if _, err := client.DeliverRequest(request); err != nil {
		t.Fatal(err)
	}

}
