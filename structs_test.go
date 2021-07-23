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

import (
	"testing"
)

func TestAddressRegexp(t *testing.T) {

	if groups := IDAddressRegexp.FindStringSubmatch("normal-name.f(acdds)"); groups == nil {
		t.Fatal("invalid match")
	} else if groups[1] != "normal-name" {
		t.Fatal("invalid operator name")
	} else if groups[2] != "f" {
		t.Fatalf("invalid method name: '%s'", groups[2])
	} else if groups[3] != "acdds" {
		t.Fatal("invalid ID")
	}

	if groups := IDAddressRegexp.FindStringSubmatch("a.b.c.d.e.f(acdds)"); groups == nil {
		t.Fatal("invalid match")
	} else if groups[1] != "a.b.c.d.e" {
		t.Fatal("invalid operator name")
	} else if groups[2] != "f" {
		t.Fatalf("invalid method name: '%s'", groups[2])
	} else if groups[3] != "acdds" {
		t.Fatal("invalid ID")
	}
}
