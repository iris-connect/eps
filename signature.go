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
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
)

type Signature struct {
	R           string `json:"r"`
	S           string `json:"s"`
	Certificate string `json:"c"`
}

type SignedData struct {
	Signature *Signature  `json:"signature"`
	Data      interface{} `json:"data"`
}

func VerifyCertificate(cert, rootCert *x509.Certificate, name string) error {
	roots := x509.NewCertPool()
	roots.AddCert(rootCert)

	opts := x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
	}

	if name != "" {
		opts.DNSName = name
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

	return LoadCertificateFromString(string(certificatePEM), verifyUsage)
}

func LoadCertificateFromString(data string, verifyUsage bool) (*x509.Certificate, error) {

	block, _ := pem.Decode([]byte(data))

	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("not a certificate")
	}

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

func BigInt(s string) (*big.Int, error) {

	i := &big.Int{}

	if _, ok := i.SetString(s, 10); !ok {
		return nil, fmt.Errorf("not a big integer in base 10")
	}

	return i, nil
}

func LoadSignedData(data []byte) (*SignedData, error) {
	signedData := &SignedData{}
	return signedData, json.Unmarshal(data, &signedData)
}

func Verify(signedData *SignedData, rootCert *x509.Certificate, name string) (bool, error) {

	cert, err := LoadCertificateFromString(signedData.Signature.Certificate, true)

	if err != nil {
		return false, err
	}

	// root certificate verification can be skipped (but shouldn't be)
	if rootCert != nil {
		if err := VerifyCertificate(cert, rootCert, name); err != nil {
			return false, nil
		}
	}

	if rawData, err := json.Marshal(signedData.Data); err != nil {
		return false, err
	} else {
		s := sha256.Sum256(rawData)

		ir, err := BigInt(signedData.Signature.R)

		if err != nil {
			return false, err
		}

		is, err := BigInt(signedData.Signature.S)

		if err != nil {
			return false, err
		}

		return ecdsa.Verify(cert.PublicKey.(*ecdsa.PublicKey), s[:], ir, is), nil
	}
}

func Sign(data interface{}, key *ecdsa.PrivateKey, cert *x509.Certificate) (*SignedData, error) {

	rawData, err := json.Marshal(data)

	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(rawData)

	r, s, err := ecdsa.Sign(rand.Reader, key, hash[:])

	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}

	if err != nil {
		return nil, err
	} else {
		return &SignedData{
			Data: data,
			Signature: &Signature{
				R:           r.String(),
				S:           s.String(),
				Certificate: string(pem.EncodeToMemory(block)),
			},
		}, nil
	}
}
