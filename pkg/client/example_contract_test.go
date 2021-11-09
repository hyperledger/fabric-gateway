/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client_test

import (
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
)

func ExampleContract_Evaluate() {
	var contract *client.Contract // Obtained from Network.

	result, err := contract.Evaluate(
		"transactionName",
		client.WithArguments("one", "two"),
		// Specify additional proposal options, such as transient data
	)

	fmt.Printf("Result: %s, Err: %v", result, err)
}

func ExampleContract_Submit() {
	var contract *client.Contract // Obtained from Network.

	result, err := contract.Submit(
		"transactionName",
		client.WithArguments("one", "two"),
		// Specify additional proposal options, such as transient data.
	)

	fmt.Printf("Result: %s, Err: %v", result, err)
}

func ExampleContract_Submit_privateData() {
	var contract *client.Contract // Obtained from Network.

	privateData := map[string][]byte{
		"price": []byte("3000"),
	}

	result, err := contract.Submit(
		"transactionName",
		client.WithArguments("one", "two"),
		client.WithTransient(privateData),
		client.WithEndorsingOrganizations("Org1MSP", "Org3MSP"),
	)

	fmt.Printf("Result: %s, Err: %v", result, err)
}

func ExampleContract_SubmitAsync() {
	var contract *client.Contract // Obtained from Network.

	// Create a transaction proposal.
	result, commit, err := contract.SubmitAsync("transactionName", client.WithArguments("one", "two"))
	if err != nil {
		panic(err)
	}

	// Use transaction result to update UI or return REST response after successful submit to the orderer.
	fmt.Printf("Result: %s", result)

	// Wait for transaction commit.
	status, err := commit.Status()
	if err != nil {
		panic(err)
	}
	if !status.Successful {
		panic(fmt.Errorf("transaction %s failed to commit with status code %d", status.TransactionID, int32(status.Code)))
	}
}

func ExampleContract_offlineSign() {
	var gateway *client.Gateway
	var contract *client.Contract // Obtained from Network.
	var sign identity.Sign        // Signing function.

	// Create a transaction proposal.
	unsignedProposal, err := contract.NewProposal("transactionName", client.WithArguments("one", "two"))
	if err != nil {
		panic(err)
	}

	// Off-line sign the proposal.
	proposalBytes, err := unsignedProposal.Bytes()
	if err != nil {
		panic(err)
	}
	proposalDigest := unsignedProposal.Digest()
	proposalSignature, err := sign(proposalDigest)
	if err != nil {
		panic(err)
	}
	signedProposal, err := gateway.NewSignedProposal(proposalBytes, proposalSignature)
	if err != nil {
		panic(err)
	}

	// Endorse proposal to create an endorsed transaction.
	unsignedTransaction, err := signedProposal.Endorse()
	if err != nil {
		panic(err)
	}

	// Off-line sign the transaction.
	transactionBytes, err := unsignedTransaction.Bytes()
	if err != nil {
		panic(err)
	}
	digest := unsignedTransaction.Digest()
	transactionSignature, err := sign(digest)
	if err != nil {
		panic(err)
	}
	signedTransaction, err := gateway.NewSignedTransaction(transactionBytes, transactionSignature)
	if err != nil {
		panic(err)
	}

	// Submit transaction to the orderer.
	unsignedCommit, err := signedTransaction.Submit()
	if err != nil {
		panic(err)
	}

	// Off-line sign the transaction commit status request
	commitBytes, err := unsignedCommit.Bytes()
	if err != nil {
		panic(err)
	}
	commitDigest := unsignedCommit.Digest()
	commitSignature, err := sign(commitDigest)
	if err != nil {
		panic(err)
	}
	signedCommit, err := gateway.NewSignedCommit(commitBytes, commitSignature)

	// Wait for transaction commit.
	status, err := signedCommit.Status()
	if err != nil {
		panic(err)
	}
	if !status.Successful {
		panic(fmt.Errorf("transaction %s failed to commit with status code %d", status.TransactionID, int32(status.Code)))
	}

	fmt.Printf("Result: %s, Err: %v", signedTransaction.Result(), err)
}
