// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"google.golang.org/grpc/status"
)

func ExampleContract_Evaluate() {
	var contract *client.Contract // Obtained from Network.

	result, err := contract.Evaluate(
		"transactionName",
		client.WithArguments("one", "two"),
		// Specify additional proposal options, such as transient data
	)

	fmt.Printf("Result: %s, Err: %v\n", result, err)
}

func ExampleContract_Evaluate_errorHandling() {
	var contract *client.Contract // Obtained from Network.

	result, err := contract.Evaluate("transactionName")
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			panic(fmt.Errorf("timeout: %w", err))
		} else {
			panic(fmt.Errorf("gRPC status %v: %w", status.Code(err), err))
		}
	}

	fmt.Printf("Result: %s, Err: %v\n", result, err)
}

func ExampleContract_Submit() {
	var contract *client.Contract // Obtained from Network.

	result, err := contract.Submit(
		"transactionName",
		client.WithArguments("one", "two"),
		// Specify additional proposal options, such as transient data.
	)

	fmt.Printf("Result: %s, Err: %v\n", result, err)
}

func ExampleContract_Submit_errorHandling() {
	var contract *client.Contract // Obtained from Network.

	result, err := contract.Submit("transactionName")
	if err != nil {
		var endorseErr *client.EndorseError
		var submitErr *client.SubmitError
		var commitStatusErr *client.CommitStatusError
		var commitErr *client.CommitError

		if errors.As(err, &endorseErr) {
			fmt.Printf("Failed to endorse proposal for transaction %s with gRPC status %v: %s\n",
				endorseErr.TransactionID, status.Code(endorseErr), endorseErr)
		} else if errors.As(err, &submitErr) {
			fmt.Printf("Failed to submit endorsed transaction %s to orderer with gRPC status %v: %s\n",
				submitErr.TransactionID, status.Code(submitErr), submitErr)
		} else if errors.As(err, &commitStatusErr) {
			if errors.Is(err, context.DeadlineExceeded) {
				fmt.Printf("Timeout waiting for transaction %s commit status: %s",
					commitStatusErr.TransactionID, commitStatusErr)
			} else {
				fmt.Printf("Failed to obtain commit status for transaction %s with gRPC status %v: %s\n",
					commitStatusErr.TransactionID, status.Code(commitStatusErr), commitStatusErr)
			}
		} else if errors.As(err, &commitErr) {
			fmt.Printf("Transaction %s failed to commit with status %d: %s\n",
				commitErr.TransactionID, int32(commitErr.Code), err)
		} else {
			fmt.Printf("Unexpected error type %T: %s", err, err)
		}
	}

	fmt.Printf("Result: %s, Err: %v\n", result, err)
}

func ExampleContract_Submit_errorDetails() {
	var contract *client.Contract // Obtained from Network.

	result, err := contract.Submit("transactionName")
	fmt.Printf("Result: %s, Err: %v\n", result, err)

	if err != nil {
		// Any error that originates from a peer or orderer node external to the gateway will have its details
		// embedded within the gRPC status error. The following code shows how to extract that.
		for _, detail := range status.Convert(err).Details() {
			switch detail := detail.(type) {
			case *gateway.ErrorDetail:
				fmt.Printf("- address: %s; mspId: %s; message: %s\n", detail.GetAddress(), detail.GetMspId(), detail.GetMessage())
			}
		}
	}
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

	// Submit transaction to the orderer.
	result, commit, err := contract.SubmitAsync("transactionName", client.WithArguments("one", "two"))
	panicOnError(err)

	// Use transaction result to update UI or return REST response after successful submit to the orderer.
	fmt.Printf("Result: %s", result)

	// Wait for transaction commit.
	status, err := commit.Status()
	panicOnError(err)
	if !status.Successful {
		panic(fmt.Errorf("transaction %s failed to commit with status code %d", status.TransactionID, int32(status.Code)))
	}
}

func ExampleContract_NewProposal() {
	var contract *client.Contract // Obtained from Network.

	proposal, err := contract.NewProposal("transactionName", client.WithArguments("one", "two"))
	panicOnError(err)

	transaction, err := proposal.Endorse()
	panicOnError(err)

	commit, err := transaction.Submit()
	panicOnError(err)

	status, err := commit.Status()
	panicOnError(err)

	fmt.Printf("Commit status code: %d, Result: %s\n", int32(status.Code), transaction.Result())
}

func ExampleContract_offlineSign() {
	var gateway *client.Gateway
	var contract *client.Contract // Obtained from Network.
	var sign identity.Sign        // Signing function.

	// Create a transaction proposal.
	unsignedProposal, err := contract.NewProposal("transactionName", client.WithArguments("one", "two"))
	panicOnError(err)

	// Off-line sign the proposal.
	proposalBytes, err := unsignedProposal.Bytes()
	panicOnError(err)
	proposalDigest := unsignedProposal.Digest()
	proposalSignature, err := sign(proposalDigest)
	panicOnError(err)
	signedProposal, err := gateway.NewSignedProposal(proposalBytes, proposalSignature)
	panicOnError(err)

	// Endorse proposal to create an endorsed transaction.
	unsignedTransaction, err := signedProposal.Endorse()
	panicOnError(err)

	// Off-line sign the transaction.
	transactionBytes, err := unsignedTransaction.Bytes()
	panicOnError(err)
	transactionDigest := unsignedTransaction.Digest()
	transactionSignature, err := sign(transactionDigest)
	panicOnError(err)
	signedTransaction, err := gateway.NewSignedTransaction(transactionBytes, transactionSignature)
	panicOnError(err)

	// Submit transaction to the orderer.
	unsignedCommit, err := signedTransaction.Submit()
	panicOnError(err)

	// Off-line sign the transaction commit status request
	commitBytes, err := unsignedCommit.Bytes()
	panicOnError(err)
	commitDigest := unsignedCommit.Digest()
	commitSignature, err := sign(commitDigest)
	panicOnError(err)
	signedCommit, err := gateway.NewSignedCommit(commitBytes, commitSignature)
	panicOnError(err)

	// Wait for transaction commit.
	status, err := signedCommit.Status()
	panicOnError(err)
	if !status.Successful {
		panic(fmt.Errorf("transaction %s failed to commit with status code %d", status.TransactionID, int32(status.Code)))
	}

	fmt.Printf("Result: %s\n", signedTransaction.Result())
}
