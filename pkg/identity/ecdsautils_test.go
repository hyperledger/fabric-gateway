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
)

func TestToLowSByKey(t *testing.T) {
	lowLevelKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	s, err := toLowSByKey(&lowLevelKey.PublicKey, big.NewInt(0))
	if err != nil {
		t.Fatal(err)
	}

	if s.Cmp(big.NewInt(0)) != 0 {
		t.Fatalf("Unexpected change of S value")
	}

	s = new(big.Int)
	s.Set(getCurveHalfOrdersAt(elliptic.P256()))
	compareS := new(big.Int)
	compareS.Set(getCurveHalfOrdersAt(elliptic.P256()))

	s, err = toLowSByKey(&lowLevelKey.PublicKey, s)
	if err != nil {
		t.Fatal(err)
	}

	if s.Cmp(compareS) != 0 {
		t.Fatalf("Unexpected change of S value")
	}

	s = s.Add(s, big.NewInt(100))
	compareS = compareS.Add(s, big.NewInt(100))

	s, err = toLowSByKey(&lowLevelKey.PublicKey, s)
	if err != nil {
		t.Fatal(err)
	}

	if s.Cmp(compareS) != -1 {
		t.Fatalf("S value has not been reduced")
	}

	compareS = big.NewInt(0)
	compareS = compareS.Set(s)

	s, err = toLowSByKey(&lowLevelKey.PublicKey, s)
	if err != nil {
		t.Fatal(err)
	}

	if s.Cmp(compareS) != 0 {
		t.Fatalf("Unexpected change of S value")
	}
}

func TestToLowSByCurve(t *testing.T) {
	curve := elliptic.P256()

	s, err := toLowSByCurve(curve, big.NewInt(0))
	if err != nil {
		t.Fatal(err)
	}

	if s.Cmp(big.NewInt(0)) != 0 {
		t.Fatalf("Unexpected change of S value")
	}

	s = new(big.Int)
	s.Set(getCurveHalfOrdersAt(elliptic.P256()))
	compareS := new(big.Int)
	compareS.Set(getCurveHalfOrdersAt(elliptic.P256()))

	s, err = toLowSByCurve(curve, s)
	if err != nil {
		t.Fatal(err)
	}

	if s.Cmp(compareS) != 0 {
		t.Fatalf("Unexpected change of S value")
	}

	s = s.Add(s, big.NewInt(100))
	compareS = compareS.Add(s, big.NewInt(100))

	s, err = toLowSByCurve(curve, s)
	if err != nil {
		t.Fatal(err)
	}

	if s.Cmp(compareS) != -1 {
		t.Fatalf("S value has not been reduced")
	}

	compareS = big.NewInt(0)
	compareS = compareS.Set(s)

	s, err = toLowSByCurve(curve, s)
	if err != nil {
		t.Fatal(err)
	}

	if s.Cmp(compareS) != 0 {
		t.Fatalf("Unexpected change of S value")
	}
}
