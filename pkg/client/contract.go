/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"github.com/golang/protobuf/proto"
	gateway "github.com/hyperledger/fabric-gateway/protos"
	"github.com/pkg/errors"
)

// Contract represents a chaincode smart contract in a network. The Contract can be used to submit and evaluate
// transaction functions on the smart contract, and to listen for chaincode events emitted by the smart contract.
type Contract struct {
	network *Network
	name    string
}

// EvaluateTransaction will evaluate a transaction function and return its results. A transaction proposal will be
// evaluated on endorsing peers but the transaction will not be sent to the ordering service and so will not be
// committed to the ledger. This can be used for querying the world state.
func (contract *Contract) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	byteArgs := stringsAsBytes(args)
	return contract.Evaluate(name, WithArguments(byteArgs...))
}

// Evaluate a transaction function and return its result. This method provides greater control over the transaction
// proposal content and the endorsing peers on which it is evaluated. This allows transaction functions to be evaluated
// where the proposal must include transient data, or that will access ledger data with key-based endorsement policies.
func (contract *Contract) Evaluate(name string, options ...ProposalOption) ([]byte, error) {
	proposal, err := contract.NewProposal(name, options...)
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
	result, commit, err := contract.SubmitAsync(transactionName, options...)
	if err != nil {
		return nil, err
	}

	if err = <-commit; err != nil {
		return nil, err
	}

	return result, nil
}

// SubmitAsync submits a transaction to the ledger and returns its result immediately after successfully sending to the
// orderer, along with a channel that can be used to receive notification when it has been committed to the ledger.
func (contract *Contract) SubmitAsync(transactionName string, options ...ProposalOption) ([]byte, chan error, error) {
	proposal, err := contract.NewProposal(transactionName, options...)
	if err != nil {
		return nil, nil, err
	}

	transaction, err := proposal.Endorse()
	if err != nil {
		return nil, nil, err
	}

	result := transaction.Result()

	commit, err := transaction.Submit()
	if err != nil {
		return nil, nil, err
	}

	return result, commit, nil
}

// NewProposal creates a proposal that can be sent to peers for endorsement. Supports off-line signing transaction flow.
func (contract *Contract) NewProposal(transactionName string, options ...ProposalOption) (*Proposal, error) {
	builder := &proposalBuilder{
		contract: contract,
		name:     transactionName,
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
	proposal := &Proposal{
		client:    contract.network.gateway.client,
		bytes:     bytes,
		signature: signature,
	}
	return proposal, nil
}

// NewSignedTransaction creates an endorsed transaction with signature, which can be submitted to the orderer for commit
// to the ledger.
func (contract *Contract) NewSignedTransaction(bytes []byte, signature []byte) (*Transaction, error) {
	var preparedTransaction *gateway.PreparedTransaction
	if err := proto.Unmarshal(bytes, &gateway.PreparedTransaction{}); err != nil {
		return nil, errors.Wrap(err, "Failed to deserialize transaction")
	}

	transaction := &Transaction{
		client:              contract.network.gateway.client,
		sign:                contract.network.gateway.sign,
		hash:                contract.network.gateway.hash,
		preparedTransaction: preparedTransaction,
	}

	transaction.setSignature(signature)

	return transaction, nil
}
