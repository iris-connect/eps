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

package helpers

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
)

func VerifyCertificate(cert, rootCert *x509.Certificate, name string) error {
	roots := x509.NewCertPool()
	roots.AddCert(rootCert)

	opts := x509.VerifyOptions{
		DNSName:   name,
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
	}

	if _, err := cert.Verify(opts); err != nil {
		return err
	}

	return nil
}

func LoadCertificate(path string, verifyUsage bool) (*x509.Certificate, error) {
	certificatePEM, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode([]byte(certificatePEM))
	cert, err := x509.ParseCertificate(block.Bytes)

	if err != nil {
		return nil, err
	}

	if verifyUsage {
		if (cert.KeyUsage & x509.KeyUsageDigitalSignature) == 0 {
			return nil, fmt.Errorf("expected a certificate for signing")
		}

		if cert.PublicKeyAlgorithm != x509.ECDSA {
			return nil, fmt.Errorf("expected an ECDSA-based certificate")
		}

	}

	return cert, nil
}

func LoadPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	keyPEM, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode([]byte(keyPEM))

	if block.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("not an EC private key")
	}

	return x509.ParseECPrivateKey(block.Bytes)
}

/*
import (
	"encoding/json"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
)


func decode(pemEncoded string, pemEncodedPub string) (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
    block, _ := pem.Decode([]byte(pemEncoded))
    x509Encoded := block.Bytes
    privateKey, _ := x509.ParseECPrivateKey(x509Encoded)

    blockPub, _ := pem.Decode([]byte(pemEncodedPub))
    x509EncodedPub := blockPub.Bytes
    genericPublicKey, _ := x509.ParsePKIXPublicKey(x509EncodedPub)
    publicKey := genericPublicKey.(*ecdsa.PublicKey)

    return privateKey, publicKey
}

func ValidateCertificate()

func SignDirectoryEntry() {
	privateKey, err := ecd
}
*/
