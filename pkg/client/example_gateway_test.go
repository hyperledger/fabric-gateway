/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client_test

import (
	"fmt"
	"io/ioutil"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/connection"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
)

func Example() {
	id := Identity()
	sign := Sign()
	endpoint := &connection.Endpoint{
		Host: "gateway.example.org",
		Port: 1337,
	}

	gateway, err := client.Connect(id, sign, client.WithEndpoint(endpoint))
	PanicOnError(err)

	network := gateway.GetNetwork("channelName")
	contract := network.GetContract("chaincodeName")

	result, err := contract.SubmitTransaction("transactionName", "arg1", "arg2")
	PanicOnError(err)

	fmt.Printf("Received transaction result: %s", result)
}

func Identity() *identity.X509Identity {
	certPEM, err := ioutil.ReadFile("certificate.pem")
	PanicOnError(err)

	certificate, err := identity.CertificateFromPEM(certPEM)
	PanicOnError(err)

	id, err := identity.NewX509Identity("mspID", certificate)
	PanicOnError(err)

	return id
}

func Sign() identity.Sign {
	privateKeyPEM, err := ioutil.ReadFile("privateKey.pem")
	PanicOnError(err)

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	PanicOnError(err)

	sign, err := identity.NewPrivateKeySign(privateKey)
	PanicOnError(err)

	return sign
}

func PanicOnError(err interface{}) {
	if err != nil {
		panic(err)
	}
}
