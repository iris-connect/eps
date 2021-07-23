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
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/iris-connect/eps"
	"io/ioutil"
	"math/big"
	"net/url"
)

func VerifyCertificate(cert, rootCert *x509.Certificate, intermediateCerts []*x509.Certificate, name string) error {
	roots := x509.NewCertPool()
	roots.AddCert(rootCert)

	intermediates := x509.NewCertPool()

	for _, intermediateCert := range intermediateCerts {
		intermediates.AddCert(intermediateCert)
	}

	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
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

func VerifyFingerprint(cert *x509.Certificate, fingerprint string) bool {
	hash := sha256.Sum256(cert.Raw)
	hexHash := hex.EncodeToString(hash[:])
	return hexHash == fingerprint
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

func LoadSignedData(data []byte) (*eps.SignedData, error) {
	signedData := &eps.SignedData{}
	return signedData, json.Unmarshal(data, &signedData)
}

func Verify(signedData *eps.SignedData, rootCerts []*x509.Certificate, intermediateCerts []*x509.Certificate, name string) (bool, error) {

	cert, err := LoadCertificateFromString(signedData.Signature.Certificate, true)

	if err != nil {
		return false, err
	}

	// root certificate verification can be skipped (but shouldn't be)
	if rootCerts != nil {
		found := false
		for _, rootCert := range rootCerts {
			if err := VerifyCertificate(cert, rootCert, intermediateCerts, name); err == nil {
				found = true
				break
			}
		}
		// no valid root certificate found
		if !found {
			return false, nil
		}
	}

	if hash, err := StructuredHash(signedData.Data); err != nil {
		return false, err
	} else {

		ir, err := BigInt(signedData.Signature.R)

		if err != nil {
			return false, err
		}

		is, err := BigInt(signedData.Signature.S)

		if err != nil {
			return false, err
		}

		return ecdsa.Verify(cert.PublicKey.(*ecdsa.PublicKey), hash, ir, is), nil
	}
}

type SubjectInfo struct {
	Name     string
	DNSNames []string
	Groups   []string
}

func GetSubjectInfo(cert *x509.Certificate) (*SubjectInfo, error) {

	// subject alternative name extension
	id := asn1.ObjectIdentifier{2, 5, 29, 17}

	subjectInfo := &SubjectInfo{
		DNSNames: make([]string, 0),
		Groups:   make([]string, 0),
	}

	for _, extension := range cert.Extensions {
		if !extension.Id.Equal(id) {
			continue
		}
		// we unmarshal the ASN.1 object
		altNames := []asn1.RawValue{}
		if _, err := asn1.Unmarshal(extension.Value, &altNames); err != nil {
			return nil, err
		}
		for _, altName := range altNames {
			if altName.Tag == 6 {
				// tag 6 is URI (don't ask why...)
				if groupUrl, err := url.Parse(string(altName.Bytes)); err != nil {
					return nil, err
				} else {
					switch groupUrl.Scheme {
					case "iris-group":
						if groupUrl.Host != "" {
							subjectInfo.Groups = append(subjectInfo.Groups, groupUrl.Host)
						}
					case "iris-name":
						subjectInfo.Name = groupUrl.Host
					}
				}
			} else if altName.Tag == 2 {
				// tag 2 is DNS (don't ask why...)
				subjectInfo.DNSNames = append(subjectInfo.DNSNames, string(altName.Bytes))
			}
		}
	}
	return subjectInfo, nil

}
func Sign(data interface{}, key *ecdsa.PrivateKey, cert *x509.Certificate) (*eps.SignedData, error) {

	hash, err := StructuredHash(data)

	if err != nil {
		return nil, err
	}

	r, s, err := ecdsa.Sign(rand.Reader, key, hash[:])

	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}

	if err != nil {
		return nil, err
	} else {
		return &eps.SignedData{
			Data: data,
			Signature: &eps.Signature{
				R:           r.String(),
				S:           s.String(),
				Certificate: string(pem.EncodeToMemory(block)),
			},
		}, nil
	}
}
