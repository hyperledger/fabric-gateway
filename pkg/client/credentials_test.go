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
	id   identity.Identity
	sign identity.Sign
}

var _testCredentials *testCredentials

func GetTestCredentials() (identity.Identity, identity.Sign) {
	if nil == _testCredentials {
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

		_testCredentials = &testCredentials{
			id:   id,
			sign: sign,
		}
	}

	return _testCredentials.id, _testCredentials.sign
}
