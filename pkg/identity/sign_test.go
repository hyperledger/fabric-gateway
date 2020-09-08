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

	t.Run("Private key PEM conversion", func(t *testing.T) {
		inputKey := newECDSAPrivateKey(t)

		pem, err := PrivateKeyToPEM(inputKey)
		if err != nil {
			t.Fatalf("Failed to create PEM: %v", err)
		}

		outputKey, err := PrivateKeyFromPEM(pem)
		if err != nil {
			t.Fatalf("Failed to create private key: %v", err)
		}

		if !(inputKey.PublicKey.X.Cmp(outputKey.(*ecdsa.PrivateKey).PublicKey.X) == 0 &&
			inputKey.PublicKey.Y.Cmp(outputKey.(*ecdsa.PrivateKey).PublicKey.Y) == 0 &&
			inputKey.D.Cmp(outputKey.(*ecdsa.PrivateKey).D) == 0) {
			//if !inputKey.Equal(outputKey) {
			t.Fatalf("Keys do not match. Expected:\n%v\nGot:\n%v", inputKey, outputKey)
		}
	})

	t.Run("Create private key fails with invalid PEM", func(t *testing.T) {
		pem := []byte("Non-PEM content")

		_, err := PrivateKeyFromPEM(pem)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})

	t.Run("Create private key from PEM fails with invalid private key data", func(t *testing.T) {
		pem := []byte("-----BEGIN PRIVATE KEY-----\nBAD/DATA-----END PRIVATE KEY-----")

		_, err := PrivateKeyFromPEM(pem)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})

	t.Run("Convert bad private key to PEM fails", func(t *testing.T) {
		var privateKey crypto.PrivateKey

		_, err := PrivateKeyToPEM(privateKey)
		if err == nil {
			t.Fatalf("Expected error, got nil")
		}
	})
}
