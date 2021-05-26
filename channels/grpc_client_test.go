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

func TestGRPCClientConnection(t *testing.T) {

	fixtures := []th.FC{
		{fixtures.Settings{}, "settings"},
		{fixtures.Directory{}, "directory"},
		{fixtures.MessageBroker{}, "broker"},
		{fixtures.Channel{"test gRPC client"}, "client"},
		{fixtures.Channel{"test gRPC server"}, "server"},
	}

	fc, err := th.SetupFixtures(fixtures)

	if err != nil {
		t.Fatal(err)
	}

	defer th.TeardownFixtures(fixtures, fc)

	client := fc["client"].(*channels.GRPCClientChannel)
	server := fc["server"].(*channels.GRPCServerChannel)

	if err := server.Open(); err != nil {
		t.Fatal(err)
	}

	if err := client.Open(); err != nil {
		t.Fatal(err)
	}

	request := &eps.Request{
		ID: "hd-1.add(1)",
	}

	if _, err := client.DeliverRequest(request); err == nil {
		t.Fatalf("expeceted an error")
	}

}
