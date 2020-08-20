/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/hyperledger/fabric/bccsp/utils"
)

// Signer function generates a digital signature of the supplied digest
type Signer = func(digest []byte) ([]byte, error)

// NewPrivateKeySigner returns a Signer that uses the supplied private key
func NewPrivateKeySigner(privateKey crypto.PrivateKey) (Signer, error) {
	switch key := privateKey.(type) {
	case *ecdsa.PrivateKey:
		return newECDSAPrivateKeySigner(key), nil
	default:
		return nil, fmt.Errorf("Unsupported key type: %T", privateKey)
	}
}

func newECDSAPrivateKeySigner(privateKey *ecdsa.PrivateKey) Signer {
	return func(digest []byte) ([]byte, error) {
		r, s, err := ecdsa.Sign(rand.Reader, privateKey, digest)
		if err != nil {
			return nil, err
		}

		s, err = utils.ToLowS(&privateKey.PublicKey, s)
		if err != nil {
			return nil, err
		}

		return utils.MarshalECDSASignature(r, s)
	}
}

// PrivateKeyFromPEM creates a private key from PEM encoded data
func PrivateKeyFromPEM(privateKeyPEM []byte) (crypto.PrivateKey, error) {
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("Failed to parse private key PEM")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// PrivateKeyToPEM converts a private key to PEM encoded PKCS #8 data
func PrivateKeyToPEM(privateKey crypto.PrivateKey) ([]byte, error) {
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	var buffer bytes.Buffer
	if err := pem.Encode(&buffer, block); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
