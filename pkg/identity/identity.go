/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"crypto/x509"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/msp"
)

// Identity represents a client identity.
type Identity interface {
	MspID() string       // ID of the Membership Service Provider to which this identity belongs.
	Credentials() []byte // Implementation-specific credentials.
}

// X509Identity represents a client identity backed by an X.509 certificate.
type X509Identity struct {
	mspID       string
	certificate []byte
}

// MspID returns the ID of the Membership Service Provider to which this identity belongs.
func (id *X509Identity) MspID() string {
	return id.mspID
}

// Credentials as an X.509 certificate in PEM encoded ASN.1 DER format.
func (id *X509Identity) Credentials() []byte {
	return id.certificate
}

// NewX509Identity creates a new Identity from an X.509 certificate.
func NewX509Identity(mspID string, certificate *x509.Certificate) (*X509Identity, error) {
	certificatePEM, err := CertificateToPEM(certificate)
	if err != nil {
		return nil, err
	}

	identity := &X509Identity{
		mspID:       mspID,
		certificate: certificatePEM,
	}
	return identity, nil
}

// Serialize an identity to protobuf SerializedIdentity message bytes
func Serialize(id Identity) ([]byte, error) {
	serializedIdentity := &msp.SerializedIdentity{
		Mspid:   id.MspID(),
		IdBytes: id.Credentials(),
	}
	return proto.Marshal(serializedIdentity)
}

// Deserialize SerializedIdentity protobuf message bytes to an Identity
func Deserialize(message []byte) (Identity, error) {
	serializedIdentity := &msp.SerializedIdentity{}
	if err := proto.Unmarshal(message, serializedIdentity); err != nil {
		return nil, err
	}

	result := &X509Identity{
		mspID:       serializedIdentity.Mspid,
		certificate: serializedIdentity.IdBytes,
	}
	return result, nil
}
