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
			t.Errorf("Failed to generate key")
		}
		return privateKey
	}

	assertSignature := func(t *testing.T, sign Sign) {
		signature, err := sign([]byte("digest"))
		if err != nil {
			t.Errorf("Signing error: %v", err)
		}

		if signature == nil {
			t.Errorf("Signature was nil")
		}
	}

	t.Run("Create signer with unsupported private key type fails", func(t *testing.T) {
		var privateKey crypto.PrivateKey
		_, err := NewPrivateKeySign(privateKey)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}

		expectedType := fmt.Sprintf("%T", privateKey)
		if !strings.Contains(err.Error(), expectedType) {
			t.Errorf("Expected error to contain %s: %s", expectedType, err)
		}
	})

	t.Run("Create signer with ECDSA private key PEM", func(t *testing.T) {
		privateKey := newECDSAPrivateKey(t)
		pem, err := PrivateKeyToPEM(privateKey)
		if err != nil {
			t.Errorf("Failed to create PEM: %v", err)
		}

		signer, err := NewPrivateKeyPEMSign(pem)
		if err != nil {
			t.Errorf("Failed to create identity: %v", err)
		}

		assertSignature(t, signer)
	})

	t.Run("Create signer with ECDSA private key", func(t *testing.T) {
		privateKey := newECDSAPrivateKey(t)
		signer, err := NewPrivateKeySign(privateKey)
		if err != nil {
			t.Errorf("Failed to create identity: %v", err)
		}

		assertSignature(t, signer)
	})

	t.Run("Private key PEM conversion", func(t *testing.T) {
		inputKey := newECDSAPrivateKey(t)

		pem, err := PrivateKeyToPEM(inputKey)
		if err != nil {
			t.Errorf("Failed to create PEM: %v", err)
		}

		outputKey, err := PrivateKeyFromPEM(pem)
		if err != nil {
			t.Errorf("Failed to create private key: %v", err)
		}

		if !inputKey.Equal(outputKey) {
			t.Errorf("Keys do not match. Expected:\n%v\nGot:\n%v", inputKey, outputKey)
		}
	})

	t.Run("Create private key from invalid PEM fails", func(t *testing.T) {
		pem := []byte("Non-PEM content")

		_, err := PrivateKeyFromPEM(pem)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("Create signer from PEM with invalid private key data fails", func(t *testing.T) {
		pem := []byte("-----BEGIN PRIVATE KEY-----\nBAD/DATA-----END PRIVATE KEY-----")

		_, err := NewPrivateKeyPEMSign(pem)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})

	t.Run("Convert bad private key to PEM fails", func(t *testing.T) {
		var privateKey crypto.PrivateKey

		_, err := PrivateKeyToPEM(privateKey)
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	})
}
