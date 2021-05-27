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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

func TLSConfig(settings *TLSSettings) (*tls.Config, error) {

	certPool := x509.NewCertPool()

	for _, certificateFile := range settings.CACertificateFiles {

		bs, err := ioutil.ReadFile(certificateFile)

		if err != nil {
			return nil, err
		}

		if ok := certPool.AppendCertsFromPEM(bs); !ok {
			return nil, fmt.Errorf("cannot import CA certificate")
		}

	}

	certs := []tls.Certificate{}

	// we only add a certificate if it is given (for clients we can e.g. omit it)
	if settings.CertificateFile != "" && settings.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(settings.CertificateFile, settings.KeyFile)

		if err != nil {
			return nil, err
		}

		certs = append(certs, cert)

	}

	tlsConfig := &tls.Config{
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		},
		PreferServerCipherSuites: true,
		Certificates:             certs,
		ClientCAs:                certPool,
		RootCAs:                  certPool,
		ServerName:               settings.ServerName,
	}

	return tlsConfig, nil
}

func TLSClientConfig(settings *TLSSettings) (*tls.Config, error) {

	if config, err := TLSConfig(settings); err != nil {
		return nil, err
	} else {
		return config, nil
	}
}

func TLSServerConfig(settings *TLSSettings) (*tls.Config, error) {

	if config, err := TLSConfig(settings); err != nil {
		return nil, err
	} else {
		if settings.VerifyClient {
			config.ClientAuth = tls.RequireAndVerifyClientCert
		}
		return config, nil
	}
}
