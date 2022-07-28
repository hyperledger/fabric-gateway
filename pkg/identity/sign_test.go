/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"crypto"
	"crypto/ecdsa"
	"encoding/asn1"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
	"github.com/stretchr/testify/require"
)

func TestSigner(t *testing.T) {
	privateKey, err := test.NewECDSAPrivateKey()
	require.NoError(t, err)

	digest := make([]byte, 256/8)
	rand.Read(digest)

	t.Run("Create signer with unsupported private key type fails", func(t *testing.T) {
		var privateKey crypto.PrivateKey
		_, err := NewPrivateKeySign(privateKey)
		require.Error(t, err)

		expectedType := fmt.Sprintf("%T", privateKey)
		require.Contains(t, err.Error(), expectedType)
	})

	t.Run("Create signer with ECDSA private key", func(t *testing.T) {
		sign, err := NewPrivateKeySign(privateKey)
		require.NoError(t, err)

		signature, err := sign(digest)
		require.NoError(t, err, "sign")

		isValid := ecdsa.VerifyASN1(&privateKey.PublicKey, digest, signature)
		require.True(t, isValid, "valid signature")
	})

	t.Run("ECDSA signatures are canonical", func(t *testing.T) {
		sign, err := NewPrivateKeySign(privateKey)
		require.NoError(t, err)

		halfOrder := new(big.Int).Rsh(privateKey.Params().N, 1)

		for i := 0; i < 10; i++ {
			signature, err := sign(digest)
			require.NoError(t, err, "sign")

			signatureRS := &ecdsaSignature{}
			_, err = asn1.Unmarshal(signature, signatureRS)
			require.NoError(t, err, "asn1.Unmarshal")

			require.LessOrEqual(t, signatureRS.S.Cmp(halfOrder), 0, "malleable: S = %v, halfOrder = %v", signatureRS.S, halfOrder)
		}
	})
}

func BenchmarkECDSA(b *testing.B) {
	privateKey, err := test.NewECDSAPrivateKey()
	require.NoError(b, err)

	sign, err := NewPrivateKeySign(privateKey)
	require.NoError(b, err)

	digest := make([]byte, 256/8)
	rand.Read(digest)

	for i := 0; i < b.N; i++ {
		sign(digest)
	}
}
