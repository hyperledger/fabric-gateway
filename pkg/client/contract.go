/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"github.com/golang/protobuf/proto"
	gateway "github.com/hyperledger/fabric-gateway/protos"
	"github.com/pkg/errors"
)

// Contract represents a chaincode smart contract.
type Contract struct {
	network *Network
	name    string
}

// EvaluateTransaction will evaluate a transaction function and return its results. The transaction function 'name'
// will be evaluated on the endorsing peers but the responses will not be sent to the ordering service and hence will
// not be committed to the ledger. This can be used for querying the world state.
func (contract *Contract) EvaluateTransaction(name string, args ...string) ([]byte, error) {
	byteArgs := stringsAsBytes(args)
	return contract.Evaluate(name, WithArguments(byteArgs...))
}

// Evaluate a transaction and return its result.
func (contract *Contract) Evaluate(name string, options ...ProposalOption) ([]byte, error) {
	proposal, err := contract.NewProposal(name, options...)
	if err != nil {
		return nil, err
	}

	return proposal.Evaluate()
}

// SubmitTransaction will submit a transaction to the ledger. The transaction function 'name' will be evaluated on the
// endorsing peers and then submitted to the ordering service for committing to the ledger.
func (contract *Contract) SubmitTransaction(name string, args ...string) ([]byte, error) {
	byteArgs := stringsAsBytes(args)
	return contract.SubmitSync(name, WithArguments(byteArgs...))
}

// SubmitSync a transaction and returns its result immediately after successfully sending the the orderer.
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

// SubmitAsync a transaction and returns its result immediately after successfully sending the the orderer.
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

	commit, err := transaction.Commit()
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

// NewSignedTransaction creates an endorsed transaction with signature, which can be submitted to the orderer for commit to the ledger.
func (contract *Contract) NewSignedTransaction(bytes []byte, signature []byte) (*Transaction, error) {
	var preparedTransaction *gateway.PreparedTransaction
	if err := proto.Unmarshal(bytes, &gateway.PreparedTransaction{}); err != nil {
		return nil, errors.Wrap(err, "Failed to deserialize transaction")
	}

	transaction := &Transaction{
		client:              contract.network.gateway.client,
		preparedTransaction: preparedTransaction,
	}

	transaction.setSignature(signature)

	return transaction, nil
}

func stringsAsBytes(strings []string) [][]byte {
	results := make([][]byte, len(strings))

	for i, v := range strings {
		results[i] = []byte(v)
	}

	return results
}
