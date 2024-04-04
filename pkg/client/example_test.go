/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client_test

import (
	"crypto/x509"
	"fmt"
	"os"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func Example() {
	// Create gRPC client connection, which should be shared by all gateway connections to this endpoint.
	clientConnection, err := NewGrpcConnection()
	panicOnError(err)
	defer clientConnection.Close()

	// Create client identity and signing implementation based on X.509 certificate and private key.
	id := NewIdentity()
	sign := NewSign()

	// Create a Gateway connection for a specific client identity.
	gateway, err := client.Connect(id, client.WithSign(sign), client.WithClientConnection(clientConnection))
	panicOnError(err)
	defer gateway.Close()

	// Obtain smart contract deployed on the network.
	network := gateway.GetNetwork("channelName")
	contract := network.GetContract("chaincodeName")

	// Submit transactions that store state to the ledger.
	submitResult, err := contract.SubmitTransaction("transactionName", "arg1", "arg2")
	panicOnError(err)
	fmt.Printf("Submit result: %s", string(submitResult))

	// Evaluate transactions that query state from the ledger.
	evaluateResult, err := contract.EvaluateTransaction("transactionName", "arg1", "arg2")
	panicOnError(err)
	fmt.Printf("Evaluate result: %s", string(evaluateResult))
}

// NewGrpcConnection creates a new gRPC client connection
func NewGrpcConnection() (*grpc.ClientConn, error) {
	tlsCertificatePEM, err := os.ReadFile("tlsRootCertificate.pem")
	panicOnError(err)

	tlsCertificate, err := identity.CertificateFromPEM(tlsCertificatePEM)
	panicOnError(err)

	certPool := x509.NewCertPool()
	certPool.AddCert(tlsCertificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, "")

	return grpc.NewClient("dns:///gateway.example.org:1337", grpc.WithTransportCredentials(transportCredentials))
}

// NewIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func NewIdentity() *identity.X509Identity {
	certificatePEM, err := os.ReadFile("certificate.pem")
	panicOnError(err)

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	panicOnError(err)

	id, err := identity.NewX509Identity("mspID", certificate)
	panicOnError(err)

	return id
}

// NewSign creates a function that generates a digital signature from a message digest using a private key.
func NewSign() identity.Sign {
	privateKeyPEM, err := os.ReadFile("privateKey.pem")
	panicOnError(err)

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	panicOnError(err)

	sign, err := identity.NewPrivateKeySign(privateKey)
	panicOnError(err)

	return sign
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
