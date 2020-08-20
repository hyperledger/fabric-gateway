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
	"testing"
)

func TestUnsupportedPrivateKeyType(t *testing.T) {
	var privateKey crypto.PrivateKey
	_, err := NewPrivateKeySigner(privateKey)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestECDSASigner(t *testing.T) {
	privateKey := newECDSAPrivateKey(t)
	signer, err := NewPrivateKeySigner(privateKey)
	if err != nil {
		t.Errorf("Failed to create identity: %v", err)
	}

	assertSignature(t, signer)
}

func TestPrivateKeyPEMConversion(t *testing.T) {
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
}

func newECDSAPrivateKey(t *testing.T) *ecdsa.PrivateKey {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Errorf("Failed to generate key")
	}
	return privateKey
}

func assertSignature(t *testing.T, signer Signer) {
	signature, err := signer([]byte("digest"))
	if err != nil {
		t.Errorf("Signing error: %v", err)
	}

	if signature == nil {
		t.Errorf("Signature was nil")
	}
}
