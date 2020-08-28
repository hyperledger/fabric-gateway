/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// Identity represents a client identity.
type Identity struct {
	MspID   string // ID of the Membership Service Provider to which this identity belongs.
	IDBytes []byte // Credential data. For an X.509 identity this would be PEM encoded ASN.1 DER certificate data.
}

// NewX509Identity creates a new Identity from a certificate PEM.
func NewX509Identity(mspID string, certificate *x509.Certificate) (*Identity, error) {
	certificatePEM, err := CertificateToPEM(certificate)
	if err != nil {
		return nil, err
	}

	identity := &Identity{
		MspID:   mspID,
		IDBytes: certificatePEM,
	}
	return identity, nil
}

// CertificateToPEM converts a certificate to PEM encoded ASN.1 DER data.
func CertificateToPEM(certificate *x509.Certificate) ([]byte, error) {
	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificate.Raw,
	}
	return pemEncode(block)
}

// CertificateFromPEM creates a certificate from PEM encoded data.
func CertificateFromPEM(certificatePEM []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(certificatePEM)
	if block == nil {
		return nil, fmt.Errorf("Failed to parse certificate PEM")
	}

	return x509.ParseCertificate(block.Bytes)
}
