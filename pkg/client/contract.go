/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/gateway"
)

// Contract represents a smart contract, and allows applications to:
//
// - Evaluate transactions that query state from the ledger using the EvaluateTransaction() method.
//
// - Submit transactions that store state to the ledger using the SubmitTransaction() method.
//
// For more complex transaction invocations, such as including transient data, transactions can be evaluated or
// submitted using the Evaluate() or SubmitSync() methods respectively.
//
// By default, proposal and transaction messages will be signed using the signing implementation specified when
// connecting the Gateway. In cases where an external client holds the signing credentials, a signing implementation
// can be omitted when connecting the Gateway and off-line signing can be carried out by:
//
// 1. Returning the serialized proposal or transaction message along with its digest to the client for them to
// generate a signature.
//
// 2. On receipt of the serialized message and signature from the client, creating a signed proposal or transaction
// using the Contract's NewSignedProposal() or NewSignedTransaction() methods respectively.
type Contract struct {
	client       gateway.GatewayClient
	signingID    *signingIdentity
	channelName  string
	chaincodeID  string
	contractName string
}

// EvaluateTransaction will evaluate a transaction function and return its results. A transaction proposal will be
// evaluated on endorsing peers but the transaction will not be sent to the ordering service and so will not be
// committed to the ledger. This can be used for querying the world state.
func (contract *Contract) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	return contract.Evaluate(name, WithStringArguments(args...))
}

// Evaluate a transaction function and return its result. This method provides greater control over the transaction
// proposal content and the endorsing peers on which it is evaluated. This allows transaction functions to be evaluated
// where the proposal must include transient data, or that will access ledger data with key-based endorsement policies.
func (contract *Contract) Evaluate(transactionName string, options ...ProposalOption) ([]byte, error) {
	proposal, err := contract.NewProposal(transactionName, options...)
	if err != nil {
		return nil, err
	}

	return proposal.Evaluate()
}

// SubmitTransaction will submit a transaction to the ledger and return its result only after it is committed to the
// ledger. The transaction function will be evaluated on endorsing peers and then submitted to the ordering service to
// be committed to the ledger. This method is equivalent to:
//   SubmitSync(name, client.WithStringArguments(args...))
func (contract *Contract) SubmitTransaction(name string, args ...string) ([]byte, error) {
	return contract.SubmitSync(name, WithStringArguments(args...))
}

// SubmitSync submits a transaction to the ledger and returns its result only after it has been committed to the ledger.
// This method provides greater control over the transaction proposal content and the endorsing peers on which it is
// evaluated. This allows transaction functions to be submitted where the proposal must include transient data, or that
// will access ledger data with key-based endorsement policies.
func (contract *Contract) SubmitSync(transactionName string, options ...ProposalOption) ([]byte, error) {
	result, err := contract.SubmitAsync(transactionName, options...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// SubmitAsync submits a transaction to the ledger and returns its result immediately after successfully sending to the
// orderer, along with a channel that can be used to receive notification when it has been committed to the ledger.
func (contract *Contract) SubmitAsync(transactionName string, options ...ProposalOption) ([]byte, error) {
	proposal, err := contract.NewProposal(transactionName, options...)
	if err != nil {
		return nil, err
	}

	transaction, err := proposal.Endorse()
	if err != nil {
		return nil, err
	}

	result := transaction.Result()

	_, err = transaction.Submit()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// NewProposal creates a proposal that can be sent to peers for endorsement. Supports off-line signing transaction flow.
func (contract *Contract) NewProposal(transactionName string, options ...ProposalOption) (*Proposal, error) {
	builder := &proposalBuilder{
		client:          contract.client,
		signingID:       contract.signingID,
		channelName:     contract.channelName,
		chaincodeID:     contract.chaincodeID,
		transactionName: contract.qualifiedTransactionName(transactionName),
	}

	for _, option := range options {
		if err := option(builder); err != nil {
			return nil, err
		}
	}

	return builder.build()
}

// NewSignedProposal creates a transaction proposal with signature, which can be sent to peers for endorsement.
func (contract *Contract) NewSignedProposal(bytes []byte, signature []byte) (*Proposal, error) {
	proposedTransaction := &gateway.ProposedTransaction{}
	if err := proto.Unmarshal(bytes, proposedTransaction); err != nil {
		return nil, fmt.Errorf("failed to deserialize proposal: %w", err)
	}

	proposal := &Proposal{
		client:              contract.client,
		signingID:           contract.signingID,
		channelID:           contract.channelName,
		proposedTransaction: proposedTransaction,
	}
	proposal.setSignature(signature)

	return proposal, nil
}

// NewSignedTransaction creates an endorsed transaction with signature, which can be submitted to the orderer for commit
// to the ledger.
func (contract *Contract) NewSignedTransaction(bytes []byte, signature []byte) (*Transaction, error) {
	preparedTransaction := &gateway.PreparedTransaction{}
	if err := proto.Unmarshal(bytes, preparedTransaction); err != nil {
		return nil, fmt.Errorf("failed to deserialize transaction: %w", err)
	}

	transaction := &Transaction{
		client:              contract.client,
		signingID:           contract.signingID,
		channelID:           contract.channelName,
		preparedTransaction: preparedTransaction,
	}
	transaction.setSignature(signature)

	return transaction, nil
}

func (contract *Contract) qualifiedTransactionName(name string) string {
	if len(contract.contractName) > 0 {
		return contract.contractName + ":" + name
	}
	return name
}
