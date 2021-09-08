/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"crypto"
	"fmt"
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
	"github.com/stretchr/testify/require"
)

func TestSigner(t *testing.T) {
	privateKey, err := test.NewECDSAPrivateKey()
	require.NoError(t, err)

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

		signature, err := sign([]byte("digest"))
		require.NoError(t, err)

		require.NotEmpty(t, signature)
	})
}
