/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"strings"
	"testing"
)

func TestSigner(t *testing.T) {
	newECDSAPrivateKey := func(t *testing.T) *ecdsa.PrivateKey {
		privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatalf("Failed to generate key")
		}
		return privateKey
	}

	assertSignature := func(t *testing.T, sign Sign) {
		signature, err := sign([]byte("digest"))
		if err != nil {
			t.Fatalf("Signing error: %v", err)
		}

		if signature == nil {
			t.Fatalf("Signature was nil")
		}
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
		privateKey := newECDSAPrivateKey(t)
		signer, err := NewPrivateKeySign(privateKey)
		if err != nil {
			t.Fatalf("Failed to create identity: %v", err)
		}

		assertSignature(t, signer)
	})
}
