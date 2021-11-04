/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	gwproto "github.com/hyperledger/fabric-protos-go/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
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
	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(clientConnection),
		// Default timeouts for different gRPC calls
		client.WithEvaluateTimeout(5*time.Second),
		client.WithEndorseTimeout(15*time.Second),
		client.WithSubmitTimeout(5*time.Second),
		client.WithCommitStatusTimeout(1*time.Minute),
	)
	if err != nil {
		panic(err)
	}
	defer gateway.Close()

	fmt.Println("exampleSubmit:")
	exampleSubmit(gateway)
	fmt.Println()

	fmt.Println("exampleSubmitAsync:")
	exampleSubmitAsync(gateway)
	fmt.Println()

	fmt.Println("exampleSubmitPrivateData:")
	exampleSubmitPrivateData(gateway)
	fmt.Println()

	fmt.Println("exampleSubmitPrivateData2:")
	exampleSubmitPrivateData2(gateway)
	fmt.Println()

	fmt.Println("exampleStateBasedEndorsement:")
	exampleStateBasedEndorsement(gateway)
	fmt.Println()

	fmt.Println("exampleChaincodeEvents:")
	exampleChaincodeEvents(gateway)
	fmt.Println()

	fmt.Println("exampleChaincodeEventReplay:")
	exampleChaincodeEventReplay(gateway)
	fmt.Println()

	fmt.Println("exampleErrorHandling:")
	exampleErrorHandling(gateway)
	fmt.Println()
}

func exampleSubmit(gateway *client.Gateway) {
	network := gateway.GetNetwork("mychannel")
	contract := network.GetContract("basic")

	timestamp := time.Now().String()
	fmt.Printf("Submitting \"put\" transaction with arguments: time, %s\n", timestamp)

	// Submit transaction, blocking until the transaction has been committed on the ledger
	submitResult, err := contract.SubmitTransaction("put", "time", timestamp)
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Printf("Submit result: %s\n", string(submitResult))
	fmt.Println("Evaluating \"get\" query with arguments: time")

	evaluateResult, err := contract.EvaluateTransaction("get", "time")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}

	fmt.Printf("Query result = %s\n", string(evaluateResult))
}

func exampleSubmitAsync(gateway *client.Gateway) {
	network := gateway.GetNetwork("mychannel")
	contract := network.GetContract("basic")

	timestamp := time.Now().String()
	fmt.Printf("Submitting \"put\" transaction asynchronously with arguments: async, %s\n", timestamp)

	// Submit transaction asynchronously, blocking until the transaction has been sent to the orderer, and allowing
	// this thread to process the chaincode response (e.g. update a UI) without waiting for the commit notification
	submitResult, commit, err := contract.SubmitAsync("put", client.WithArguments("async", timestamp))
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction asynchronously: %w", err))
	}

	fmt.Printf("Submit result: %s\n", string(submitResult))
	fmt.Println("Waiting for transaction commit")

	if status, err := commit.Status(); err != nil {
		panic(fmt.Errorf("failed to get commit status: %w", err))
	} else if !status.Successful {
		panic(fmt.Errorf("transaction %s, failed to commit with status: %d", status.TransactionID, int32(status.Code)))
	}

	fmt.Printf("Transaction committed successfully\n")
	fmt.Println("Evaluating \"get\" query with arguments: async")

	evaluateResult, err := contract.EvaluateTransaction("get", "async")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}

	fmt.Printf("Query result = %s\n", string(evaluateResult))
}

func exampleSubmitPrivateData(gateway *client.Gateway) {
	network := gateway.GetNetwork("mychannel")
	contract := network.GetContract("private")

	timestamp := time.Now().String() // This is our 'sensitive' data for this example
	privateData := map[string][]byte{
		"collection": []byte("SharedCollection,Org3Collection"), // SharedCollection owned by Org1 & Org3, Org3Collection owned by Org3.
		"key":        []byte("my-private-key"),
		"value":      []byte(timestamp),
	}
	fmt.Printf("Submitting \"WritePrivateData\" transaction with private data: %s\n", privateData["value"])

	// Submit transaction, blocking until the transaction has been committed on the ledger.
	// The 'transient' data will not get written to the ledger, and is used to send sensitive data to the trusted endorsing peers.
	// The gateway will only send this to peers that are included in the ownership policy of all collections accessed by the chaincode function.
	// It is assumed that the gateway's organization is trusted and will invoke the chaincode to work out if extra endorsements are required from other orgs.
	// In this example, it will also seek endorsement from Org3, which is included in the ownership policy of both collections.
	if _, err := contract.Submit("WritePrivateData", client.WithTransient(privateData)); err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Printf("Transaction committed successfully\n")
	fmt.Println("Evaluating \"ReadPrivateData\" query with arguments: \"SharedCollection\", \"my-private-key\"")

	evaluateResult, err := contract.EvaluateTransaction("ReadPrivateData", "SharedCollection", "my-private-key")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}

	fmt.Printf("Query result = %s\n", string(evaluateResult))
}

func exampleSubmitPrivateData2(gateway *client.Gateway) {
	network := gateway.GetNetwork("mychannel")
	contract := network.GetContract("private")

	timestamp := time.Now().String() // This is our 'sensitive' data for this example
	privateData := map[string][]byte{
		"collection": []byte("Org1Collection,Org3Collection"), // Org1Collection owned by Org1, Org3Collection owned by Org3.
		"key":        []byte("my-private-key2"),
		"value":      []byte(timestamp),
	}
	fmt.Printf("Submitting \"WritePrivateData\" transaction with private data: %s\n", privateData["value"])

	// This example is similar to the previous private data example.
	// The difference here is that the gateway cannot assume that Org3 is trusted to receive transient data
	// that might be destined for storage in Org1Collection, since Org3 is not in its ownership policy.
	// The client application must explicitly specify which organizations must endorse using the WithEndorsingOrganizations() functional argument.
	if _, err := contract.Submit("WritePrivateData",
		client.WithTransient(privateData),
		client.WithEndorsingOrganizations("Org1MSP", "Org3MSP"),
	); err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	fmt.Printf("Transaction committed successfully\n")
	fmt.Println("Evaluating \"ReadPrivateData\" query with arguments: \"Org1Collection\", \"my-private-key2\"")

	evaluateResult, err := contract.EvaluateTransaction("ReadPrivateData", "Org1Collection", "my-private-key2")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}

	fmt.Printf("Query result = %s\n", string(evaluateResult))
}

func exampleStateBasedEndorsement(gateway *client.Gateway) {
	network := gateway.GetNetwork("mychannel")
	contract := network.GetContract("private")

	fmt.Println("Submitting \"SetStateWithEndorser\" transaction with arguments: \"sbe-key\", \"value1\", \"Org1MSP\"")
	// Submit transaction, blocking until the transaction has been committed on the ledger
	if _, err := contract.SubmitTransaction("SetStateWithEndorser", "sbe-key", "value1", "Org1MSP"); err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}
	fmt.Println("Transaction committed successfully")

	// Query the current state
	fmt.Println("Evaluating \"GetState\" query with arguments: \"sbe-key\"")
	evaluateResult, err := contract.EvaluateTransaction("GetState", "sbe-key")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	fmt.Printf("Query result = %s\n", string(evaluateResult))

	// Submit transaction to modify the state.
	// The state-based endorsement policy will override the chaincode policy for this state (key).
	fmt.Println("Submitting \"ChangeState\" transaction with arguments: \"sbe-key\", \"value2\"")
	if _, err = contract.SubmitTransaction("ChangeState", "sbe-key", "value2"); err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}
	fmt.Println("Transaction committed successfully")

	// Verify the current state
	fmt.Println("Evaluating \"GetState\" query with arguments: \"sbe-key\"")
	evaluateResult, err = contract.EvaluateTransaction("GetState", "sbe-key")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	fmt.Printf("Query result = %s\n", string(evaluateResult))

	// Now change the state-based endorsement policy for this state.
	fmt.Println("Submitting \"SetStateEndorsers\" transaction with arguments: \"sbe-key\", \"Org2MSP\", \"Org3MSP\"")
	if _, err = contract.SubmitTransaction("SetStateEndorsers", "sbe-key", "Org2MSP", "Org3MSP"); err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}
	fmt.Println("Transaction committed successfully")

	// Modify the state.  It will now require endorsement from Org2 and Org3 for this transaction to succeed.
	// The gateway will endorse this transaction proposal on one of its organization's peers and will determine if
	// extra endorsements are required to satisfy any state changes.
	// In this example, it will seek endorsements from Org2 and Org3 in order to satisfy the SBE policy.
	fmt.Println("Submitting \"ChangeState\" transaction with arguments: \"sbe-key\", \"value3\"")
	if _, err = contract.SubmitTransaction("ChangeState", "sbe-key", "value3"); err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}
	fmt.Println("Transaction committed successfully")

	// Verify the new state
	fmt.Println("Evaluating \"GetState\" query with arguments: \"sbe-key\"")
	evaluateResult, err = contract.EvaluateTransaction("GetState", "sbe-key")
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}
	fmt.Printf("Query result = %s\n", string(evaluateResult))
}

func exampleChaincodeEvents(gateway *client.Gateway) {
	network := gateway.GetNetwork("mychannel")
	contract := network.GetContract("basic")

	fmt.Printf("Read chaincode events")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events, err := network.ChaincodeEvents(ctx, "basic")
	if err != nil {
		panic(fmt.Errorf("failed to read chaincode events: %w", err))
	}

	// Submit a transaction that generates a chaincode event
	fmt.Println("Submitting \"event\" transaction with arguments: \"my-event-name\", \"my-event-payload\"")
	_, err = contract.SubmitTransaction("event", "my-event-name", "my-event-payload")
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	select {
	case ev := <-events:
		fmt.Printf("Received event name: %s, payload: %s, txId: %s\n", ev.EventName, ev.Payload, ev.TransactionID)
	case <-time.After(10 * time.Second):
		fmt.Println("Timed out waiting for chaincode event")
	}
}

func exampleChaincodeEventReplay(gateway *client.Gateway) {
	network := gateway.GetNetwork("mychannel")
	contract := network.GetContract("basic")

	// Submit a transaction that generates a chaincode event
	fmt.Println("Submitting \"event\" transaction with arguments: \"my-event-name\", \"my-event-replay-payload\"")
	_, commit, err := contract.SubmitAsync("event", client.WithArguments("my-event-name", "my-event-replay-payload"))
	if err != nil {
		panic(fmt.Errorf("failed to submit transaction: %w", err))
	}

	status, err := commit.Status()
	if err != nil {
		panic(fmt.Errorf("failed to get commit status: %w", err))
	}
	if !status.Successful {
		panic(fmt.Errorf("transaction failed to commit with status: %d", int32(status.Code)))
	}

	fmt.Printf("Read chaincode events starting at block number %d\n", status.BlockNumber)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events, err := network.ChaincodeEvents(ctx, "basic", client.WithStartBlock(status.BlockNumber))
	if err != nil {
		panic(fmt.Errorf("failed to read chaincode events: %w", err))
	}

	select {
	case ev := <-events:
		fmt.Printf("Received event name: %s, payload: %s, txId: %s\n", ev.EventName, ev.Payload, ev.TransactionID)
	case <-time.After(10 * time.Second):
		fmt.Println("Timed out waiting for chaincode event")
	}
}

func exampleErrorHandling(gateway *client.Gateway) {
	network := gateway.GetNetwork("mychannel")
	contract := network.GetContract("basic")

	fmt.Println("Submitting \"put\" transaction without arguments")

	// Submit transaction, passing in the wrong number of arguments.
	_, err := contract.SubmitTransaction("put")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		// Any error that originates from a peer or orderer node external to the gateway will have its details
		// embedded within the grpc status error.  The following code shows how to extract that.
		statusErr := status.Convert(err)
		for _, detail := range statusErr.Details() {
			errDetail := detail.(*gwproto.ErrorDetail)
			fmt.Printf("Error from endpoint: %s, mspId: %s, message: %s\n", errDetail.Address, errDetail.MspId, errDetail.Message)
		}
	}
}

// newGrpcConnection creates a gRPC connection to the Gateway server.
func newGrpcConnection() *grpc.ClientConn {
	certificate, err := loadCertificate(tlsCertPath)
	if err != nil {
		panic(fmt.Errorf("failed to obtain commit status: %w", err))
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, "peer0.org1.example.com")

	connection, err := grpc.Dial(peerEndpoint, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		panic(fmt.Errorf("failed to evaluate transaction: %w", err))
	}

	return connection
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

func loadCertificate(filename string) (*x509.Certificate, error) {
	certificatePEM, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return identity.CertificateFromPEM(certificatePEM)
}
