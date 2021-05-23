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
	"encoding/binary"
	"fmt"
	"regexp"
)

type ProtocolVersion struct {
	Minor uint8
	Major uint8
}

type ClientHello struct {
	ProtocolVersion    ProtocolVersion `json:"protocol_version"`
	Random             Random          `json:"random"`
	SessionID          []byte          `json:"session_id"`
	CipherSuites       [][2]uint8      `json:"cipher_suites"`
	CompressionMethods []uint8         `json:"compression_methods"`
	Extensions         []Extension     `json:"extensions"`
}

func (c *ClientHello) ServerNameList() *ServerNameList {
	for _, extension := range c.Extensions {
		if extension.Type == ServerNameExtension {
			return extension.Struct.(*ServerNameList)
		}
	}
	return nil
}

type Extension struct {
	Type   ExtensionType `json:"type"`
	Data   []byte        `json:"data"`
	Struct interface{}   `json:"struct"`
}

type ServerNameList struct {
	ServerNames []ServerName `json:"server_names"`
}

func (s *ServerNameList) HostName() string {
	for _, serverName := range s.ServerNames {
		if serverName.NameType == HostNameType {
			return serverName.HostName
		}
	}
	return ""
}

type ServerName struct {
	NameType ServerNameType `json:"name_type"`
	HostName string         `json:"host_name"`
}

type ServerNameType uint8

const (
	HostNameType ServerNameType = 0 // the only name type we're interested in....
)

type ExtensionType uint16

const (
	ServerNameExtension ExtensionType = 0 // the only extension type we're interested in...
)

type Random struct {
	GMTUnixTime uint32   `json:"gmt_unix_time"`
	RandomBytes [28]byte `json:"random_bytes"`
}

// SNI hostnames do not include the trailing dot.
var HostNameRegexp = regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9-]{0,62}\.)*([a-zA-Z0-9][a-zA-Z0-9-]{0,62})$`)

func ParseClientHello(data []byte) (*ClientHello, error) {

	clientHello := &ClientHello{}

	c := data

	if len(c) < 1 {
		return nil, fmt.Errorf("out of data when parsing record type")
	}

	recordType := uint8(c[0])

	if recordType != 22 {
		return nil, fmt.Errorf("expected type '22' but received %d", recordType)
	}

	c = c[1:]

	if len(c) < 2 { // parse minor and major protocol version
		return nil, fmt.Errorf("out of data when parsing protocol version")
	}

	clientHello.ProtocolVersion = ProtocolVersion{
		Minor: uint8(c[0]),
		Major: uint8(c[1]),
	}

	c = c[2:]

	if len(c) < 2 { // parse length
		return nil, fmt.Errorf("out of data when parsing length")
	}

	length := binary.BigEndian.Uint16(c[:2])

	c = c[2:]

	if int(length) > len(c) { // length of remaining record
		return nil, fmt.Errorf("incomplete record")
	}

	if len(c) < 1 { // parse message type
		return nil, fmt.Errorf("out of data when passing message type")
	}

	messageType := uint8(c[0])

	if messageType != 1 {
		return nil, fmt.Errorf("expected client hello")
	}

	c = c[1:]

	if len(c) < 3 {
		return nil, fmt.Errorf("out of data when parsing message length")
	}

	var d [4]byte

	copy(d[1:4], c[:3])

	messageLength := binary.BigEndian.Uint32(d[:])

	c = c[3:]

	if int(messageLength) > len(c) {
		return nil, fmt.Errorf("incomplete record")
	}

	if len(c) < 2 {
		return nil, fmt.Errorf("out of data when parsing protocol version")
	}

	majorClientVersion := uint8(c[0])
	minorClientVersion := uint8(c[1])

	if majorClientVersion != 3 || minorClientVersion != 3 {
		return nil, fmt.Errorf("expected 3/3")
	}

	c = c[2:]

	if len(c) < 4 { // parse unix time
		return nil, fmt.Errorf("out of data when parsing random unix time")
	}

	// actually this is just random data in TLS 1.2...
	// so no worries if the time doesn't match
	clientHello.Random.GMTUnixTime = binary.BigEndian.Uint32(c[:4])

	c = c[4:]

	if len(c) < 28 { // random data length
		return nil, fmt.Errorf("out of data when parsing random data")
	}

	copy(clientHello.Random.RandomBytes[:], c[:28])

	c = c[28:]

	if len(c) < 1 { // session ID length
		return nil, fmt.Errorf("out of data when parsing session ID length")
	}

	sessionIDLength := uint8(c[0])

	if sessionIDLength > 32 {
		return nil, fmt.Errorf("invalid sessionID length (max 32 bytes, but was %d)", sessionIDLength)
	}

	c = c[1:]

	if len(c) < int(sessionIDLength) { // session ID
		return nil, fmt.Errorf("out of data when parsing session ID")
	}

	clientHello.SessionID = c[:sessionIDLength]

	c = c[sessionIDLength:]

	if len(c) < 2 {
		return nil, fmt.Errorf("out of data when parsing ciphers length")
	}

	ciphersLength := binary.BigEndian.Uint16(c[:2])

	c = c[2:]

	if int(ciphersLength) > len(c) {
		return nil, fmt.Errorf("incomplete record (ciphers: %d)", ciphersLength)
	}

	if int(ciphersLength)%2 != 0 {
		return nil, fmt.Errorf("expected an even number")
	}

	ciphers := int(ciphersLength) / 2

	clientHello.CipherSuites = make([][2]byte, ciphers)

	for i := 0; i < ciphers; i++ {
		clientHello.CipherSuites[i][0] = c[i*2]
		clientHello.CipherSuites[i][0] = c[i*2+1]
	}

	c = c[ciphersLength:]

	if len(c) < 1 {
		return nil, fmt.Errorf("out of data when parsing compression methods length")
	}

	compressionLength := uint8(c[0])

	c = c[1:]

	if len(c) < int(compressionLength) {
		return nil, fmt.Errorf("out of data when parsing compression methods")
	}

	clientHello.CompressionMethods = c[:compressionLength]

	c = c[compressionLength:]

	if len(c) < 2 {
		return nil, fmt.Errorf("out of data when parsing extensions length")
	}

	extensionsLength := binary.BigEndian.Uint16(c[:2])

	c = c[2:]

	if int(extensionsLength) > len(c) {
		return nil, fmt.Errorf("error when parsing extensions")
	}

	// if there's extraneous data at the end we ignore it (e.g. in case the
	// client tries to fool us or sends additional TCP data, which it shouldn't)
	c = c[:extensionsLength]

	extensions := make([]Extension, 0)

	for {
		if len(c) == 0 {
			break
		}

		if len(c) < 2 {
			return nil, fmt.Errorf("out of data when parsing extension type")
		}

		extensionType := binary.BigEndian.Uint16(c[:2])

		c = c[2:]

		if len(c) < 2 {
			return nil, fmt.Errorf("out of data when parsing extension length")
		}

		extensionLength := binary.BigEndian.Uint16(c[:2])

		c = c[2:]

		if len(c) < int(extensionLength) {
			return nil, fmt.Errorf("out of data when parsing extension data")
		}

		extensionData := c[:extensionLength]

		c = c[extensionLength:]

		extension := Extension{
			Type: ExtensionType(extensionType),
			Data: extensionData,
		}

		d := extensionData

		switch ExtensionType(extensionType) {
		case ServerNameExtension:
			extensionStruct := &ServerNameList{
				ServerNames: make([]ServerName, 0),
			}
			extension.Struct = extensionStruct
			if len(d) < 2 {
				return nil, fmt.Errorf("out of data when parsing server name list length")
			}

			serverNameListLength := binary.BigEndian.Uint16(d[:2])

			d = d[2:]

			if int(serverNameListLength) < len(d) {
				return nil, fmt.Errorf("out of data when parsing server name list")
			}

			if len(d) < 1 {
				return nil, fmt.Errorf("out of data when parsing server name type")
			}

			serverNameType := uint8(d[0])

			d = d[1:]

			switch ServerNameType(serverNameType) {
			case HostNameType:
				if len(d) < 2 {
					return nil, fmt.Errorf("out of data when parsing host name length")
				}

				hostNameLength := binary.BigEndian.Uint16(d[:2])

				d = d[2:]

				if int(hostNameLength) != len(d) {
					return nil, fmt.Errorf("unexpected data length when parsing host name")
				}

				hostName := d[:hostNameLength]

				if len(hostName) > 253 {
					return nil, fmt.Errorf("hostname is too long")
				}

				if !HostNameRegexp.Match(hostName) {
					return nil, fmt.Errorf("invalid hostname detected")
				}

				extensionStruct.ServerNames = append(extensionStruct.ServerNames, ServerName{
					NameType: HostNameType,
					HostName: string(hostName),
				})

				d = d[hostNameLength:]

			}
		}

		extensions = append(extensions, extension)

	}

	clientHello.Extensions = extensions

	return clientHello, nil
}

/*
// https://datatracker.ietf.org/doc/html/rfc5246

The structure of the ClientHello record is described here:

struct {
   ProtocolVersion client_version;
   Random random;
   SessionID session_id;
   CipherSuite cipher_suites<2..2^16-2>;
   CompressionMethod compression_methods<1..2^8-1>;
   select (extensions_present) {
       case false:
           struct {};
       case true:
           Extension extensions<0..2^16-1>;
   };
} ClientHello;

struct {
  uint8 major;
  uint8 minor;
} ProtocolVersion;

 struct {
     uint32 gmt_unix_time;
     opaque random_bytes[28];
 } Random;

opaque SessionID<0..32>;

uint8 CipherSuite[2];

enum { null(0), (255) } CompressionMethod;

struct {
   ExtensionType extension_type;
   opaque extension_data<0..2^16-1>;
} Extension;

// https://datatracker.ietf.org/doc/html/rfc6066#section-3

// Server Name Indication Extension Record Content is a "ServerNameList"
// as defined below:

struct {
  NameType name_type;
  select (name_type) {
      case host_name: HostName;
  } name;
} ServerName;

enum {
  host_name(0), (255)
} NameType;

opaque HostName<1..2^16-1>;

struct {
  ServerName server_name_list<1..2^16-1>
} ServerNameList;

*/
