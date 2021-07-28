/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// This code was inspired by the github.com/hyperledger/fabric/bccsp/utils/ecdsa.go

package identity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/asn1"
	"fmt"
	"math/big"
)

// ecdsaSignature represents an ECDSA Signature
type ecdsaSignature struct {
	R, S *big.Int
}

// curveHalfOrders contains the precomputed curve group orders halved.
// It is used to ensure that signature' S value is lower or equal to the
// curve group order halved. We accept only low-S signatures.
// They are precomputed for efficiency reasons.
var curveHalfOrders = map[elliptic.Curve]*big.Int{
	elliptic.P256(): new(big.Int).Rsh(elliptic.P256().Params().N, 1),
	elliptic.P384(): new(big.Int).Rsh(elliptic.P384().Params().N, 1),
}

func getCurveHalfOrdersAt(curve elliptic.Curve) *big.Int {
	return big.NewInt(0).Set(curveHalfOrders[curve])
}

// marshalECDSASignature creates an ASN1 representation of a signature
func marshalECDSASignature(r, s *big.Int) ([]byte, error) {
	return asn1.Marshal(ecdsaSignature{r, s})
}

// toLowSByKey converts S to a low value if required determining the curve from the public key. Be aware that s param is mutated
func toLowSByKey(k *ecdsa.PublicKey, s *big.Int) (*big.Int, error) {
	return toLowSByCurve(k.Curve, s)
}

// toLowSByCurve converts S to a low value if required. Be aware that s param is mutated
func toLowSByCurve(curve elliptic.Curve, s *big.Int) (*big.Int, error) {
	halfOrder, ok := curveHalfOrders[curve]
	if !ok {
		return nil, fmt.Errorf("curve not recognized [%s]", curve)
	}

	// check that s is a low-S
	if s.Cmp(halfOrder) == 1 {
		// Set s to N - s that will be then in the lower part of signature space
		// less or equal to half order
		s.Sub(curve.Params().N, s)
	}

	return s, nil
}
