/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client_test

import (
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
)

func ExampleContract_evaluate() ([]byte, error) {
	var contract *client.Contract

	result, err := contract.Evaluate(
		"transactionName",
		client.WithArguments("one", "two"),
		// Specify additional proposal options, such as transient data
	)

	return result, err
}

func ExampleContract_submit() ([]byte, error) {
	var contract *client.Contract

	result, err := contract.Submit(
		"transactionName",
		client.WithArguments("one", "two"),
		// Specify additional proposal options, such as transient data
	)

	return result, err
}

func ExampleContract_offlineSignProposal() (*client.Proposal, error) {
	var contract *client.Contract
	var sign identity.Sign // Signing function

	unsignedProposal, err := contract.NewProposal("transactionName", client.WithArguments("one", "two"))
	if err != nil {
		return nil, err
	}

	proposalBytes, err := unsignedProposal.Bytes()
	if err != nil {
		return nil, err
	}

	digest := unsignedProposal.Digest()

	// Generate signature from digest
	signature, err := sign(digest)
	if err != nil {
		return nil, err
	}

	signedProposal, err := contract.NewSignedProposal(proposalBytes, signature)

	return signedProposal, err
}

func ExampleContract_offlineSignTransaction() (*client.Transaction, error) {
	var proposal *client.Proposal
	var sign identity.Sign // Signing function
	var contract *client.Contract

	unsignedTransaction, err := proposal.Endorse()
	if err != nil {
		return nil, err
	}

	transactionBytes, err := unsignedTransaction.Bytes()
	if err != nil {
		return nil, err
	}

	digest := unsignedTransaction.Digest()

	// Generate signature from digest
	signature, err := sign(digest)
	if err != nil {
		return nil, err
	}

	signedTransaction, err := contract.NewSignedTransaction(transactionBytes, signature)

	return signedTransaction, err
}

func ExampleContract_offlineSignCommit() (*client.Commit, error) {
	var transaction *client.Transaction
	var sign identity.Sign // Signing function
	var network *client.Network

	unsignedCommit, err := transaction.Submit()
	if err != nil {
		return nil, err
	}

	commitBytes, err := unsignedCommit.Bytes()
	if err != nil {
		return nil, err
	}

	digest := unsignedCommit.Digest()

	// Generate signature from digest
	signature, err := sign(digest)
	if err != nil {
		return nil, err
	}

	signedCommit, err := network.NewSignedCommit(commitBytes, signature)

	return signedCommit, err
}
