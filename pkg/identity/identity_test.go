/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"bytes"
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
)

func TestIdentity(t *testing.T) {
	const mspID = "mspID"

	privateKey, err := test.NewECDSAPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	certificate, err := test.NewCertificate(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	certificatePEM, err := CertificateToPEM(certificate)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("NewX509Identity", func(t *testing.T) {
		identity, err := NewX509Identity(mspID, certificate)
		if err != nil {
			t.Fatalf("Failed to create identity: %v", err)
		}

		if identity.MspID() != mspID {
			t.Fatalf("Expected %s, got %s", mspID, identity.MspID())
		}
	})

	t.Run("Serialize", func(t *testing.T) {
		inputIdentity := &X509Identity{
			mspID:       mspID,
			certificate: []byte(certificatePEM),
		}

		identityMessage, err := Serialize(inputIdentity)
		if err != nil {
			t.Fatalf("Failed to serialize identity: %v", err)
		}

		outputIdentity, err := Deserialize(identityMessage)
		if err != nil {
			t.Fatalf("Failed to deserialize identity: %v", err)
		}

		if outputIdentity.MspID() != inputIdentity.MspID() {
			t.Fatalf("Expected MspID %s, got %s", inputIdentity.MspID(), outputIdentity.MspID())
		}

		if !bytes.Equal(inputIdentity.Credentials(), outputIdentity.Credentials()) {
			t.Fatalf("Expected Credentials:\n%v\nGot:\n%v", inputIdentity.Credentials(), outputIdentity.Credentials())
		}
	})

	t.Run("Deserialize fails on bad message", func(t *testing.T) {
		if identity, err := Deserialize([]byte("BAD_SERIALIZED_IDENTITY")); err == nil {
			t.Fatalf("Expected an error, got identity: %v", identity)
		}
	})
}
