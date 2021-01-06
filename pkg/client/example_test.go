/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client_test

import (
	"io/ioutil"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
)

func Example() ([]byte, error) {
	id, err := NewIdentity()
	if err != nil {
		return nil, err
	}

	sign, err := NewSign()
	if err != nil {
		return nil, err
	}

	// The gRPC client connection should be shared all Gateway connections to this endpoint
	clientConnection, err := grpc.Dial("gateway.example.org:1337", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	// Create a Gateway connection for a specific client identity
	gateway, err := client.Connect(id, client.WithSign(sign), client.WithClientConnection(clientConnection))
	if err != nil {
		return nil, err
	}

	network := gateway.GetNetwork("channelName")
	contract := network.GetContract("chaincodeName")

	result, err := contract.SubmitTransaction("transactionName", "arg1", "arg2")

	gateway.Close()

	return result, err
}

// NewIdentity creates a client identity for this Gateway connection using an X.509 certificate
func NewIdentity() (*identity.X509Identity, error) {
	certificatePEM, err := ioutil.ReadFile("certificate.pem")
	if err != nil {
		return nil, err
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		return nil, err
	}

	return identity.NewX509Identity("mspID", certificate)
}

// NewSign creates a function that generates a digital signature from a message digest using a private key
func NewSign() (identity.Sign, error) {
	privateKeyPEM, err := ioutil.ReadFile("privateKey.pem")
	if err != nil {
		return nil, err
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	return identity.NewPrivateKeySign(privateKey)
}
