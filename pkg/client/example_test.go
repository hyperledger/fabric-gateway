/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client_test

import (
	"fmt"
	"io/ioutil"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Example() {
	// Create gRPC client connection, which should be shared by all gateway connections to this endpoint.
	clientConnection, err := grpc.Dial("gateway.example.org:1337", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer clientConnection.Close()

	// Create client identity and signing implementation based on X.509 certificate and private key.
	id := NewIdentity()
	sign := NewSign()

	// Create a Gateway connection for a specific client identity.
	gateway, err := client.Connect(id, client.WithSign(sign), client.WithClientConnection(clientConnection))
	if err != nil {
		panic(err)
	}
	defer gateway.Close()

	// Obtain smart contract deployed on the network.
	network := gateway.GetNetwork("channelName")
	contract := network.GetContract("chaincodeName")

	// Submit transactions that store state to the ledger.
	submitResult, err := contract.SubmitTransaction("transactionName", "arg1", "arg2")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Submit result: %s", string(submitResult))

	// Evaluate transactions that query state from the ledger.
	evaluateResult, err := contract.EvaluateTransaction("transactionName", "arg1", "arg2")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Evaluate result: %s", string(evaluateResult))
}

// NewIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func NewIdentity() *identity.X509Identity {
	certificatePEM, err := ioutil.ReadFile("certificate.pem")
	if err != nil {
		panic(err)
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity("mspID", certificate)
	if err != nil {
		panic(err)
	}

	return id
}

// NewSign creates a function that generates a digital signature from a message digest using a private key.
func NewSign() identity.Sign {
	privateKeyPEM, err := ioutil.ReadFile("privateKey.pem")
	if err != nil {
		panic(err)
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		panic(err)
	}

	sign, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		panic(err)
	}

	return sign
}
