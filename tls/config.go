package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/iris-gateway/eps"
	"io/ioutil"
)

func TLSConfig(settings *eps.TLSSettings) (*tls.Config, error) {

	bs, err := ioutil.ReadFile(settings.CACertificateFile)

	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()

	if ok := certPool.AppendCertsFromPEM(bs); !ok {
		return nil, fmt.Errorf("cannot import CA certificate")
	}

	cert, err := tls.LoadX509KeyPair(settings.CertificateFile, settings.KeyFile)

	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		CipherSuites: []uint16{
			tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
		PreferServerCipherSuites: true,
		Certificates:             []tls.Certificate{cert},
		ClientCAs:                certPool,
		RootCAs:                  certPool,
	}

	return tlsConfig, nil
}

func TLSClientConfig(settings *eps.TLSSettings, serverName string) (*tls.Config, error) {

	if config, err := TLSConfig(settings); err != nil {
		return nil, err
	} else {
		config.ServerName = serverName
		return config, nil
	}
}

func TLSServerConfig(settings *eps.TLSSettings) (*tls.Config, error) {

	if config, err := TLSConfig(settings); err != nil {
		return nil, err
	} else {
		config.ClientAuth = tls.RequireAndVerifyClientCert
		return config, nil
	}
}
