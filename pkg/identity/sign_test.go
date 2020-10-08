/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"crypto"
	"fmt"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
)

func TestSigner(t *testing.T) {
	privateKey, err := test.NewECDSAPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Create signer with unsupported private key type fails", func(t *testing.T) {
		var privateKey crypto.PrivateKey
		_, err := NewPrivateKeySign(privateKey)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}

		expectedType := fmt.Sprintf("%T", privateKey)
		if !strings.Contains(err.Error(), expectedType) {
			t.Fatalf("Expected error to contain %s: %s", expectedType, err)
		}
	})

	t.Run("Create signer with ECDSA private key", func(t *testing.T) {
		sign, err := NewPrivateKeySign(privateKey)
		if err != nil {
			t.Fatalf("Failed to create identity: %v", err)
		}

		signature, err := sign([]byte("digest"))
		if err != nil {
			t.Fatalf("Signing error: %v", err)
		}

		if signature == nil {
			t.Fatalf("Signature was nil")
		}
	})
}
