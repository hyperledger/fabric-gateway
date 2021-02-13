/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
)

type testCredentials struct {
	identity identity.Identity
	sign     identity.Sign
}

var TestCredentials testCredentials

func init() {
	privateKey, err := test.NewECDSAPrivateKey()
	if err != nil {
		panic(err)
	}

	certificate, err := test.NewCertificate(privateKey)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity("mspID", certificate)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	TestCredentials = testCredentials{
		identity: id,
		sign:     sign,
	}
}
