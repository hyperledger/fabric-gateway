/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
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

	t.Run("NewX509Identity", func(t *testing.T) {
		identity, err := NewX509Identity(mspID, certificate)
		if err != nil {
			t.Fatalf("Failed to create identity: %v", err)
		}

		if identity.MspID() != mspID {
			t.Fatalf("Expected %s, got %s", mspID, identity.MspID())
		}
	})
}
