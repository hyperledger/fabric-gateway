/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
)

// Identity represents a client identity
type Identity struct {
	MspID       string
	Certificate *x509.Certificate
}

// NewIdentity creates a new Identity from a certificate PEM
func NewIdentity(mspID string, certificatePEM []byte) (*Identity, error) {
	certificate, err := CertificateFromPEM(certificatePEM)
	if err != nil {
		return nil, err
	}

	identity := &Identity{
		MspID:       mspID,
		Certificate: certificate,
	}
	return identity, nil
}

// Serialize an identity to protobuf SerializedIdentity message bytes
func Serialize(identity *Identity) ([]byte, error) {
	certPem, err := CertificateToPEM(identity.Certificate)
	if err != nil {
		return nil, err
	}

	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   identity.MspID,
		IdBytes: certPem,
	}
	return proto.Marshal(serializedIdentity)
}

// Deserialize SerializedIdentity protobuf message bytes to an Identity
func Deserialize(message []byte) (*Identity, error) {
	deserializedIdentity := &msp.SerializedIdentity{}
	if err := proto.Unmarshal(message, deserializedIdentity); err != nil {
		return nil, err
	}
	return NewIdentity(deserializedIdentity.Mspid, deserializedIdentity.IdBytes)
}

// CertificateToPEM converts a certificate to PEM encoded ASN.1 DER data
func CertificateToPEM(certificate *x509.Certificate) ([]byte, error) {
	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificate.Raw,
	}

	var buffer bytes.Buffer
	if err := pem.Encode(&buffer, block); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// CertificateFromPEM creates a certificate from PEM encoded data
func CertificateFromPEM(certificatePEM []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(certificatePEM)
	if block == nil {
		return nil, fmt.Errorf("Failed to parse certificate PEM")
	}

	return x509.ParseCertificate(block.Bytes)
}
