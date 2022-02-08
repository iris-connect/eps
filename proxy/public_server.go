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

/*
The public proxy accepts incoming TLS connections (using a TCP connection),
parses the `HelloClient` packet and forwards the connection to the internal
proxy via a separate TCP channel.
*/

package proxy

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/iris-connect/eps"
	epsForms "github.com/iris-connect/eps/forms"
	"github.com/iris-connect/eps/helpers"
	"github.com/iris-connect/eps/jsonrpc"
	epsNet "github.com/iris-connect/eps/net"
	"github.com/iris-connect/eps/tls"
	"github.com/kiprotect/go-helpers/forms"
	"net"
	"strings"
	"sync"
	"time"
)

type PublicServer struct {
	dataStore        eps.Datastore
	settings         *PublicServerSettings
	jsonrpcServer    *jsonrpc.JSONRPCServer
	jsonrpcClient    *jsonrpc.Client
	tlsListener      net.Listener
	internalListener net.Listener
	tlsConnections   map[string]net.Conn
	epsConnections   map[string]net.Conn
	announcements    []*PublicAnnouncement
	tlsHellos        map[string][]byte
	mutex            sync.Mutex
}

var PublicRequestConnectionForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "_client",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &epsForms.ClientInfoForm,
				},
			},
		},
		{
			Name: "to",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "channel",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
	},
}

var PublicAnnounceConnectionsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "_client",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &epsForms.ClientInfoForm,
				},
			},
		},
		{
			Name: "connections",
			Validators: []forms.Validator{
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{
							Form: &PublicConnectionForm,
						},
					},
				},
			},
		},
	},
}

var PublicConnectionForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "expires_at",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsString{},
				forms.IsTime{
					Format: "rfc3339",
				},
				IsValidExpiresAtTime{},
			},
		},
		{
			Name: "domain",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
	},
}

type PublicRequestConnectionParams struct {
	To         string          `json:"to"`
	Channel    string          `json:"channel"`
	ClientInfo *eps.ClientInfo `json:"_client"`
}

type PublicAnnounceConnectionsParams struct {
	Connections []*PublicProxyConnection
	ClientInfo  *eps.ClientInfo `json:"_client"`
}

type PublicProxyConnection struct {
	Domain    string     `json:"domain"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func (c *PublicServer) requestConnection(context *jsonrpc.Context, params *PublicRequestConnectionParams) *jsonrpc.Response {

	randomBytes, err := helpers.RandomBytes(32)

	if err != nil {
		eps.Log.Error(err)
		return context.InternalError()
	}

	randomStr := base64.StdEncoding.EncodeToString(randomBytes)

	// we tell the target EPS about the connection request
	request := jsonrpc.MakeRequest(fmt.Sprintf("%s._connectionRequest", params.To), "", map[string]interface{}{
		"client":  params.ClientInfo,
		"channel": params.Channel,
		"token":   randomStr,
	})

	if result, err := c.jsonrpcClient.Call(request); err != nil {
		eps.Log.Errorf("RPC error when announcing connection request: %v", err)
		return context.InternalError()
	} else {
		if result.Error != nil {
			eps.Log.Errorf("Error when requesting connection: %v (%v)", result.Error.Message, result.Error.Data)
			return context.Error(result.Error.Code, result.Error.Message, result.Error.Data)
		}
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// we initialize the connection with a nil value
	c.epsConnections[randomStr] = nil

	go func() {
		time.Sleep(time.Duration(c.settings.AcceptTimeout) * time.Second)
		c.mutex.Lock()
		defer c.mutex.Unlock()
		// connection still waiting, we close it
		if conn, ok := c.epsConnections[randomStr]; ok && conn == nil {
			eps.Log.Warningf("TLS connection not initiated by a single party, closing it...")
			delete(c.epsConnections, randomStr)
		}
	}()

	return context.Result(map[string]interface{}{"token": randomStr})
}

func (c *PublicServer) announceConnections(context *jsonrpc.Context, params *PublicAnnounceConnectionsParams) *jsonrpc.Response {

	results := []interface{}{}

	settings := params.ClientInfo.Entry.SettingsFor("proxy", c.settings.Name)

	if settings == nil {
		return context.Error(403, "not authorized", nil)
	}

	directorySettings := &DirectorySettings{}

	if directoryParams, err := DirectorySettingsForm.Validate(settings.Settings); err != nil {
		return context.Error(500, "invalid directory settings", nil)
	} else if err := DirectorySettingsForm.Coerce(directorySettings, directoryParams); err != nil {
		return context.InternalError()
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

connections:
	for _, connection := range params.Connections {
		eps.Log.Debugf("Received announcement for domain '%s' from operator '%s'", connection.Domain, params.ClientInfo.Name)

		found := false
		for _, allowedDomain := range directorySettings.AllowedDomains {
			if strings.HasSuffix(connection.Domain, allowedDomain) {
				found = true
				break
			}
		}
		if !found {
			results = append(results, jsonrpc.MakeError(403, "not allowed", nil))
			continue connections
		}

		var newAnnouncement *PublicAnnouncement
		changed := false
		for _, announcement := range c.announcements {
			if announcement.Domain == connection.Domain {
				if announcement.Operator != params.ClientInfo.Name {
					results = append(results, jsonrpc.MakeError(409, "already taken", nil))
					continue connections
				}
				newAnnouncement = announcement
				if (announcement.ExpiresAt != nil && connection.ExpiresAt != nil && !connection.ExpiresAt.Equal(*announcement.ExpiresAt)) || (announcement.ExpiresAt == nil && connection.ExpiresAt != nil) {
					changed = true
					// we update the expiration time
					announcement.ExpiresAt = connection.ExpiresAt
				} else if connection.ExpiresAt == nil && announcement.ExpiresAt != nil {
					// we remove the expiration time
					changed = true
					announcement.ExpiresAt = nil
				}
				break
			}
		}

		if newAnnouncement == nil {
			newAnnouncement = &PublicAnnouncement{
				Domain:    connection.Domain,
				ExpiresAt: connection.ExpiresAt,
				Operator:  params.ClientInfo.Name,
			}
			c.announcements = append(c.announcements, newAnnouncement)
			changed = true
		}

		if changed {

			eps.Log.Debugf("An announcement was added or modified.")

			id, err := helpers.RandomID(16)

			if err != nil {
				eps.Log.Error(err)
				return context.InternalError()
			}

			rawData, err := json.Marshal(newAnnouncement)

			if err != nil {
				eps.Log.Error(err)
				return context.InternalError()
			}

			dataEntry := &eps.DataEntry{
				Type: PublicAnnouncementType,
				ID:   id,
				Data: rawData,
			}

			if err := c.dataStore.Write(dataEntry); err != nil {
				eps.Log.Error(err)
				return context.InternalError()
			}

		}

		results = append(results, nil)

	}
	return context.Result(results)
}

var GetPublicAnnouncementsForm = forms.Form{
	Fields: []forms.Field{},
}

type GetPublicAnnouncementsParams struct{}

func (c *PublicServer) getAnnouncements(context *jsonrpc.Context, params *GetPublicAnnouncementsParams) *jsonrpc.Response {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	relevantAnnouncements := make([]*PublicAnnouncement, 0)
	for _, announcement := range c.announcements {
		if announcement.ExpiresAt != nil && announcement.ExpiresAt.Before(time.Now()) {
			continue
		}
		relevantAnnouncements = append(relevantAnnouncements, announcement)
	}
	return context.Result(relevantAnnouncements)
}

func MakePublicServer(settings *PublicServerSettings, definitions *eps.Definitions) (*PublicServer, error) {

	dataStore, err := helpers.InitializeDatastore(settings.Datastore, definitions)

	if err != nil {
		return nil, err
	}

	server := &PublicServer{
		settings:       settings,
		jsonrpcClient:  jsonrpc.MakeClient(settings.JSONRPCClient),
		tlsConnections: make(map[string]net.Conn),
		tlsHellos:      make(map[string][]byte),
		announcements:  make([]*PublicAnnouncement, 0),
		dataStore:      dataStore,
	}

	methods := map[string]*jsonrpc.Method{
		"requestConnection": {
			Form:    &PublicRequestConnectionForm,
			Handler: server.requestConnection,
		},
		"announceConnections": {
			Form:    &PublicAnnounceConnectionsForm,
			Handler: server.announceConnections,
		},
		"getAnnouncements": {
			Form:    &GetPublicAnnouncementsForm,
			Handler: server.getAnnouncements,
		},
	}

	if err := server.dataStore.Init(); err != nil {
		return nil, err
	}

	if err := server.update(); err != nil {
		return nil, err
	}

	handler, err := jsonrpc.MethodsHandler(methods)

	if err != nil {
		return nil, err
	}

	jsonrpcServer, err := jsonrpc.MakeJSONRPCServer(settings.JSONRPCServer, handler)

	if err != nil {
		return nil, err
	}

	server.jsonrpcServer = jsonrpcServer

	return server, nil
}

func (s *PublicServer) update() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if entries, err := s.dataStore.Read(); err != nil {
		return err
	} else {
		announcements := make([]*PublicAnnouncement, 0, len(entries))
		for _, entry := range entries {
			switch entry.Type {
			case PublicAnnouncementType:
				announcement := &PublicAnnouncement{}
				if err := json.Unmarshal(entry.Data, &announcement); err != nil {
					return fmt.Errorf("invalid record format!")
				}
				announcements = append(announcements, announcement)
			default:
				return fmt.Errorf("unknown entry type found...")
			}
		}
		validAnnouncements := make([]*PublicAnnouncement, 0)
		for _, announcement := range announcements {
			found := false
			for _, validAnnouncement := range validAnnouncements {
				if announcement.Domain == validAnnouncement.Domain && announcement.Operator == validAnnouncement.Operator {
					// we update an existing announcement
					validAnnouncement.ExpiresAt = announcement.ExpiresAt
					found = true
					break
				}
			}
			if !found {
				validAnnouncements = append(validAnnouncements, &PublicAnnouncement{
					Domain:    announcement.Domain,
					Operator:  announcement.Operator,
					ExpiresAt: announcement.ExpiresAt,
				})
			}
		}
		s.announcements = validAnnouncements
		return nil
	}
}

func (s *PublicServer) handleInternalConnection(internalConnection net.Conn) {

	eps.Log.Debugf("Internal connection received from '%s'", internalConnection.RemoteAddr().String())

	close := func() {
		internalConnection.Close()
	}

	// we give the client some time to announce itself
	internalConnection.SetReadDeadline(time.Now().Add(5 * time.Second))
	// we expect a secret token to be transmitted over the connection
	tokenBuf := make([]byte, 32)

	reqLen, err := internalConnection.Read(tokenBuf)

	if err != nil {
		eps.Log.Error(err)
		close()
		return
	}

	if reqLen != 32 {
		eps.Log.Error("Cannot read token, closing connection...")
		close()
		return
	}

	tokenStr := base64.StdEncoding.EncodeToString(tokenBuf)

	eps.Log.Debugf("Received token '%s'", tokenStr)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	var tlsConnection net.Conn
	var ok bool

	if tlsConnection, ok = s.epsConnections[tokenStr]; ok {
		// this is an EPS-EPS connection
		if tlsConnection == nil {
			// this is the first party to request a connection, we store it
			// and wait for the other party to connect...
			s.epsConnections[tokenStr] = internalConnection

			go func() {
				time.Sleep(time.Duration(s.settings.AcceptTimeout) * time.Second)
				s.mutex.Lock()
				defer s.mutex.Unlock()
				// connection still waiting, we close it
				if conn, ok := s.epsConnections[tokenStr]; ok && conn != nil {
					eps.Log.Warningf("TLS connection not accepted in time by other EPS, closing it...")
					if err := conn.Close(); err != nil {
						eps.Log.Error(err)
					}
					delete(s.epsConnections, tokenStr)
				}
			}()

			return
		}

		// we delete the connection
		delete(s.epsConnections, tokenStr)

		close = func() {
			internalConnection.Close()
			tlsConnection.Close()
		}

	} else {

		// this is a regular client-EPS connection
		tlsConnection, ok = s.tlsConnections[tokenStr]
		delete(s.tlsConnections, tokenStr)
		tlsHello, helloOk := s.tlsHellos[tokenStr]
		delete(s.tlsHellos, tokenStr)

		if !ok {
			eps.Log.Error("No connection found for token, closing...")
			internalConnection.Close()
			return
		}

		close = func() {
			internalConnection.Close()
			tlsConnection.Close()
		}

		if !helloOk {
			close()
			return
		}

		if n, err := internalConnection.Write(tlsHello); err != nil {
			eps.Log.Error(err)
			close()
			return
		} else if n != len(tlsHello) {
			eps.Log.Error("Can't forward TLS HelloClient")
			close()
			return
		}

	}

	pipe := func(left, right net.Conn) {
		buf := make([]byte, 4096)
		for {
			n, err := left.Read(buf)
			if err != nil {
				eps.Log.Error(err)
				close()
				return
			}
			if m, err := right.Write(buf[:n]); err != nil {
				eps.Log.Error(err)
				close()
				return
			} else if m != n {
				eps.Log.Errorf("cannot write all data")
				close()
				return
			}
		}
	}

	eps.Log.Debugf("Proxying connection...")

	go pipe(internalConnection, tlsConnection)
	go pipe(tlsConnection, internalConnection)

}

// we only return the first two bytes of the IP address
func anonymizeIP(ip string) string {
	values := strings.Split(ip, ".")
	// if it's an IPv6 we don't return any information currently
	if len(values) != 4 {
		return ""
	}
	return strings.Join(values[:2], ".")
}

func (s *PublicServer) handleTlsConnection(conn net.Conn) {

	eps.Log.Debugf("Received TLS connection from '%s'...", anonymizeIP(conn.RemoteAddr().String()))

	// we give the client 1 second to announce itself
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// 2 kB is more than enough for a TLS ClientHello packet
	buf := make([]byte, 2048)

	reqLen, err := conn.Read(buf)

	if err != nil {
		eps.Log.Error(err)
	}

	clientHello, err := tls.ParseClientHello(buf[:reqLen])

	close := func() {
		if err := conn.Close(); err != nil {
			eps.Log.Error(err)
		}
	}

	if err != nil {
		eps.Log.Error(err)
		close()
		return
	}

	if serverNameList := clientHello.ServerNameList(); serverNameList == nil {
		// no server name given, we close the connection
		close()
		return
	} else if hostName := serverNameList.HostName(); hostName == "" {
		close()
		return
	} else {

		var announcement *PublicAnnouncement

		eps.Log.Debugf("Looking for announcement for domain '%s'...", hostName)

		found := false
		s.mutex.Lock()
		for _, announcement = range s.announcements {
			if announcement.Domain == hostName {
				// if this announcement is already expired we ignore it
				if announcement.ExpiresAt != nil && announcement.ExpiresAt.Before(time.Now()) {
					continue
				}
				found = true
				break
			}
		}
		s.mutex.Unlock()

		// no matching announcement found...
		if !found {
			eps.Log.Debugf("No announcement found, closing...")
			close()
			return
		}

		randomBytes, err := helpers.RandomBytes(32)

		if err != nil {
			close()
			return
		}

		randomStr := base64.StdEncoding.EncodeToString(randomBytes)

		s.mutex.Lock()
		// we store the connection details for later use
		s.tlsConnections[randomStr] = conn
		s.tlsHellos[randomStr] = buf[:reqLen]
		s.mutex.Unlock()

		go func() {
			time.Sleep(time.Duration(s.settings.AcceptTimeout) * time.Second)
			s.mutex.Lock()
			defer s.mutex.Unlock()
			// connection still waiting, we close it
			if conn, ok := s.tlsConnections[randomStr]; ok {
				eps.Log.Warningf("TLS connection not accepted in time by private proxy, closing it...")
				if err := conn.Close(); err != nil {
					eps.Log.Error(err)
				}
				delete(s.tlsConnections, randomStr)
				delete(s.tlsHellos, randomStr)
			}
		}()

		// we tell the internal proxy about an incoming connection
		request := jsonrpc.MakeRequest(fmt.Sprintf("%s.incomingConnection", announcement.Operator), "", map[string]interface{}{
			"domain":   hostName,
			"token":    randomStr,
			"endpoint": s.settings.InternalEndpoint,
		})

		if result, err := s.jsonrpcClient.Call(request); err != nil {
			eps.Log.Errorf("RPC error when announcing incoming connection: %v", err)
			close()
			return
		} else {
			if result.Error != nil {
				eps.Log.Errorf("Error when announcing incoming connection: %v (%v)", result.Error.Message, result.Error.Data)
				close()
				return
			}
		}
	}
}

func (s *PublicServer) listenForTlsConnections() {
	for {
		if s.tlsListener == nil {
			// was closed
			break
		}
		conn, err := s.tlsListener.Accept()
		if err != nil {
			if err == net.ErrClosed {
				break
			}
			eps.Log.Error(err)
			continue
		}
		go s.handleTlsConnection(conn)
	}
}

func (s *PublicServer) listenForInternalConnections() {
	for {
		if s.internalListener == nil {
			// was closed
			break
		}
		conn, err := s.internalListener.Accept()
		if err != nil {
			if err == net.ErrClosed {
				break
			}
			eps.Log.Error(err)
			continue
		}
		go s.handleInternalConnection(conn)
	}

}

func (s *PublicServer) makeListener(address string) (net.Listener, error) {
	if listener, err := net.Listen("tcp", address); err != nil {
		return nil, err
	} else if s.settings.TCPRateLimits != nil {
		return epsNet.MakeRateLimitedListener(listener, s.settings.TCPRateLimits), nil
	} else {
		return listener, nil
	}

}

func (s *PublicServer) Start() error {
	var err error

	s.tlsListener, err = s.makeListener(s.settings.TLSBindAddress)
	go s.listenForTlsConnections()

	s.internalListener, err = s.makeListener(s.settings.InternalBindAddress)
	if err != nil {
		return err
	}
	go s.listenForInternalConnections()

	if err := s.jsonrpcServer.Start(); err != nil {
		return err
	}

	return nil
}

func (s *PublicServer) Stop() error {
	if s.tlsListener != nil {
		if err := s.tlsListener.Close(); err != nil {
			eps.Log.Error(err)
		}
		s.tlsListener = nil
	}
	return s.jsonrpcServer.Stop()
}
