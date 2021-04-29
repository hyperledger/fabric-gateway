/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	mspID        = "Org1MSP"
	cryptoPath   = "../../scenario/fixtures/crypto-material/crypto-config/peerOrganizations/org1.example.com"
	certPath     = cryptoPath + "/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem"
	keyPath      = cryptoPath + "/users/User1@org1.example.com/msp/keystore/key.pem"
	tlsCertPath  = cryptoPath + "/peers/peer0.org1.example.com/tls/ca.crt"
	peerEndpoint = "localhost:7051"
)

func main() {
	// The gRPC client connection should be shared by all Gateway connections to this endpoint
	clientConnection := newGrpcConnection()
	defer clientConnection.Close()

	id := newIdentity()
	sign := newSign()

	// Create a Gateway connection for a specific client identity
	gateway, err := client.Connect(id, client.WithSign(sign), client.WithClientConnection(clientConnection))
	if err != nil {
		panic(err)
	}
	defer gateway.Close()

	network := gateway.GetNetwork("mychannel")
	contract := network.GetContract("basic")

	exampleSubmit(contract, "put", "time", time.Now().String())
	exampleEvaluate(contract, "get", "time")

	fmt.Println()

	exampleSubmitAsync(contract, "put", "async", time.Now().String())
	exampleEvaluate(contract, "get", "async")

	fmt.Println()
}

func exampleSubmit(contract *client.Contract, name string, args ...string) {
	fmt.Printf("Submitting \"%s\" transaction with arguments: %s\n", name, args)

	// Submit transaction, blocking until the transaction has been committed on the ledger
	result, err := contract.SubmitTransaction(name, args...)
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Printf("Submit result: %s\n", string(result))

}

func exampleSubmitAsync(contract *client.Contract, name string, args ...string) {
	fmt.Printf("Submitting \"%s\" transaction asynchronously with arguments: %s\n", name, args)

	// Submit transaction asynchronously, blocking until the transaction has been sent to the orderer, and allowing
	// this thread to process the chaincode response (e.g. update a UI) without waiting for the commit notification
	result, commit, err := contract.SubmitAsync(name, client.WithStringArguments(args...))
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction asynchronously: %w", err))
	}
	fmt.Printf("Submit result: %s\n", string(result))

	fmt.Println("Waiting for transaction commit")

	successful, err := commit.Successful()
	if err != nil {
		panic(fmt.Errorf("failed to obtain commit status: %w", err))
	}

	if !successful {
		status, err := commit.Status()
		if err != nil {
			panic(err)
		}

		panic(fmt.Errorf("transaction %s failed to commit with status code: %d", commit.TransactionID(), int32(status)))
	}
	fmt.Printf("Transaction committed successfully\n")
}

func exampleEvaluate(contract *client.Contract, name string, args ...string) {
	fmt.Printf("Evaluating \"%s\" query with arguments: %s\n", name, args)

	result, err := contract.EvaluateTransaction(name, args...)
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}

	fmt.Printf("Query result = %s\n", string(result))
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate.
func newIdentity() *identity.X509Identity {
	certificate, err := loadCertificate(certPath)
	if err != nil {
		panic(err)
	}

	id, err := identity.NewX509Identity(mspID, certificate)
	if err != nil {
		panic(err)
	}

	return id
}

// newSign creates a function that generates a digital signature from a message digest using a private key.
func newSign() identity.Sign {
	privateKeyPEM, err := ioutil.ReadFile(keyPath)
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

// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection() *grpc.ClientConn {
	certificate, err := loadCertificate(tlsCertPath)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, "peer0.org1.example.com")

	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(err)
	}

	return connection
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return identity.CertificateFromPEM(certificatePEM)
}
