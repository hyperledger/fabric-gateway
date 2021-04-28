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

	"github.com/hyperledger/fabric-protos-go/peer"

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
	id, err := newIdentity()
	if err != nil {
		panic(err)
	}

	sign, err := newSign()
	if err != nil {
		panic(err)
	}

	clientConnection, err := getConnection()
	if err != nil {
		panic(err)
	}

	// Create a Gateway connection for a specific client identity
	gateway, err := client.Connect(id, client.WithSign(sign), client.WithClientConnection(clientConnection))
	if err != nil {
		panic(err)
	}
	defer gateway.Close()

	network := gateway.GetNetwork("mychannel")
	contract := network.GetContract("basic")

	timestamp := time.Now().String()

	// Submit a transaction, blocking until the transaction has been committed on the ledger.
	fmt.Printf("Submitting transaction to basic chaincode with value '%s'...\n", timestamp)
	result, err := contract.SubmitTransaction("put", "time", timestamp)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Submit result = %s\n\n", string(result))

	fmt.Println("Evaluating query...")
	result, err = contract.EvaluateTransaction("get", "time")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Query result = %s\n\n", string(result))

	// Submit transaction asynchronously, allowing this thread to process the chaincode response (e.g. update a UI)
	// without waiting for the commit notification
	fmt.Printf("Submitting transaction asynchronously to basic chaincode with value %s...\n", timestamp)
	result, commit, err := contract.SubmitAsync("put", client.WithStringArguments("async", timestamp))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Proposal result = %s\n", string(result))

	// wait for transactions to commit before querying the value
	status, err := commit.Status()
	if err != nil {
		panic(err)
	}
	if status != peer.TxValidationCode_VALID {
		panic(fmt.Errorf("transaction commit failed with status code: %d", int32(status)))
	}
	// Committed.  Check the value:
	result, err = contract.EvaluateTransaction("get", "async")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Transaction committed. Query result = %s\n", string(result))
}

// newIdentity creates a client identity for this Gateway connection using an X.509 certificate
func newIdentity() (*identity.X509Identity, error) {
	certificate, err := loadCertificate(certPath)
	if err != nil {
		return nil, err
	}

	return identity.NewX509Identity(mspID, certificate)
}

// newSign creates a function that generates a digital signature from a message digest using a private key
func newSign() (identity.Sign, error) {
	privateKeyPEM, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	return identity.NewPrivateKeySign(privateKey)
}

func getConnection() (*grpc.ClientConn, error) {
	certificate, err := loadCertificate(tlsCertPath)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, "peer0.org1.example.com")

	// The gRPC client connection should be shared by all Gateway connections to this endpoint
	return grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
}

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return identity.CertificateFromPEM(certificatePEM)
}
