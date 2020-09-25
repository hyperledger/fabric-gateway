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

// Evaluate a transaction and return its result.
func (contract *Contract) Evaluate(name string, options ...ProposalOption) ([]byte, error) {
	proposal, err := contract.NewProposal(name, options...)
	if err != nil {
		return nil, err
	}

	return proposal.Evaluate()
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
		contract:  contract,
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
		contract:            contract,
		preparedTransaction: preparedTransaction,
	}

	transaction.setSignature(signature)

	return transaction, nil
}
