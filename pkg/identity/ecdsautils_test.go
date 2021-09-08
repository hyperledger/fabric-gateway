/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package identity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestECDSAUtils(t *testing.T) {
	type TestDefinition struct {
		Description string
		toLowS      func(*big.Int) (*big.Int, error)
	}

	lowLevelKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	curve := elliptic.P256()

	tests := []*TestDefinition{
		{
			Description: "toLowSByKey",
			toLowS: func(s *big.Int) (*big.Int, error) {
				return toLowSByKey(&lowLevelKey.PublicKey, s)
			},
		},
		{
			Description: "toLowByCurve",
			toLowS: func(s *big.Int) (*big.Int, error) {
				return toLowSByCurve(curve, s)
			},
		},
	}

	for _, def := range tests {
		t.Run(def.Description, func(t *testing.T) {
			t.Run("zero S is not changed", func(t *testing.T) {
				s := big.NewInt(0)

				s, err := def.toLowS(s)
				require.NoError(t, err)

				require.Equal(t, 0, big.NewInt(0).Cmp(s))
			})

			t.Run("smaller than half-order S is not changed", func(t *testing.T) {
				s := big.NewInt(1)
				s.Sub(getCurveHalfOrdersAt(curve), s)
				initialS := big.NewInt(0).Set(s)

				s, err := def.toLowS(s)
				require.NoError(t, err)

				require.Equal(t, 0, initialS.Cmp(s))
			})

			t.Run("half-order S is not changed", func(t *testing.T) {
				s := getCurveHalfOrdersAt(curve)

				s, err := def.toLowS(s)
				require.NoError(t, err)

				require.Equal(t, 0, getCurveHalfOrdersAt(curve).Cmp(s))
			})

			t.Run("larger than half-order S is reduced to <= half-order", func(t *testing.T) {
				s := big.NewInt(1)
				s.Add(getCurveHalfOrdersAt(curve), s)

				s, err := def.toLowS(s)
				require.NoError(t, err)

				require.NotEqual(t, -1, getCurveHalfOrdersAt(curve).Cmp(s))
			})
		})
	}
}
