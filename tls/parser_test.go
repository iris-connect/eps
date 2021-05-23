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

package tls

import (
	"encoding/hex"
	"testing"
)

func TestCorrectHelloClient(t *testing.T) {
	data, err := hex.DecodeString("1603010200010001fc0303b66b3f8d6b1c7fbc6def4cf61a86eb5e1ade3dfbb6e1996801539e51efb6a0a620f84e5e8df6aa1bbf29dc5014996049f7774904aa22d44d1df1d4d124c0365e98003e130213031301c02cc030009fcca9cca8ccaac02bc02f009ec024c028006bc023c0270067c00ac0140039c009c0130033009d009c003d003c0035002f00ff010001750000000e000c0000096c6f63616c686f7374000b000403000102000a000c000a001d0017001e00190018337400000010000e000c02683208687474702f312e31001600000017000000310000000d002a0028040305030603080708080809080a080b080408050806040105010601030303010302040205020602002b00050403040303002d00020101003300260024001d0020094e8ee13e1d3fad26d4966d305b28dc81b3df5317820b338dcba59b77d72c41001500be00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")

	if err != nil {
		t.Fatal(err)
	}

	clientHello, err := ParseClientHello(data)

	if err != nil {
		t.Fatal(err)
	}

	if len(clientHello.Extensions) != 13 {
		t.Fatalf("expected 13 extensions, got %d", len(clientHello.Extensions))
	}

	if clientHello.Extensions[0].Type != ServerNameExtension {
		t.Fatalf("Expdcted a server name extension")
	}

	if string(clientHello.Extensions[0].Struct.(*ServerNameList).ServerNames[0].HostName) != "localhost" {
		t.Fatalf("Expected the hostname to be 'localhost'")
	}

}
