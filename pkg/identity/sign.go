/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
)

// Sign function generates a digital signature of the supplied digest.
type Sign = func(digest []byte) ([]byte, error)

// NewPrivateKeySign returns a Sign function that uses the supplied private key.
func NewPrivateKeySign(privateKey crypto.PrivateKey) (Sign, error) {
	switch key := privateKey.(type) {
	case *ecdsa.PrivateKey:
		return ecdsaPrivateKeySign(key), nil
	default:
		return nil, fmt.Errorf("unsupported key type: %T", privateKey)
	}
}

func ecdsaPrivateKeySign(privateKey *ecdsa.PrivateKey) Sign {
	return func(digest []byte) ([]byte, error) {
		r, s, err := ecdsa.Sign(rand.Reader, privateKey, digest)
		if err != nil {
			return nil, err
		}

		s, err = toLowSByKey(&privateKey.PublicKey, s)
		if err != nil {
			return nil, err
		}

		return marshalECDSASignature(r, s)
	}
}
