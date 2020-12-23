/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"time"

	gateway "github.com/hyperledger/fabric-gateway/protos"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

// Proposal represents a transaction proposal that can be sent to peers for endorsement or evaluated as a query.
type Proposal struct {
	client        gateway.GatewayClient
	signingID     *signingIdentity
	transactionID string
	bytes         []byte
	signature     []byte
}

// Bytes of the serialized proposal message.
func (proposal *Proposal) Bytes() ([]byte, error) {
	return proposal.bytes, nil
}

// Digest of the proposal. This is used to generate a digital signature.
func (proposal *Proposal) Digest() ([]byte, error) {
	return proposal.signingID.Hash(proposal.bytes)
}

// TransactionID for the proposal.
func (proposal *Proposal) TransactionID() string {
	return proposal.transactionID
}

// Endorse the proposal to obtain an endorsed transaction for submission to the orderer.
func (proposal *Proposal) Endorse() (*Transaction, error) {
	txProposal, err := proposal.newProposedTransaction()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	preparedTransaction, err := proposal.client.Endorse(ctx, txProposal)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to endorse proposal")
	}

	result := &Transaction{
		client:              proposal.client,
		signingID:           proposal.signingID,
		preparedTransaction: preparedTransaction,
	}
	return result, nil
}

// Evaluate the proposal to obtain a transaction result. This is effectively a query.
func (proposal *Proposal) Evaluate() ([]byte, error) {
	txProposal, err := proposal.newProposedTransaction()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	result, err := proposal.client.Evaluate(ctx, txProposal)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to evaluate transaction")
	}

	return result.Value, nil
}

func (proposal *Proposal) newProposedTransaction() (*gateway.ProposedTransaction, error) {
	signedProposal, err := proposal.newSignedProposal()
	if err != nil {
		return nil, err
	}

	proposedTransaction := &gateway.ProposedTransaction{
		Proposal: signedProposal,
	}
	return proposedTransaction, nil
}

func (proposal *Proposal) newSignedProposal() (*peer.SignedProposal, error) {
	if err := proposal.sign(); err != nil {
		return nil, err
	}

	signedProposal := &peer.SignedProposal{
		ProposalBytes: proposal.bytes,
		Signature:     proposal.signature,
	}
	return signedProposal, nil
}

func (proposal *Proposal) isSigned() bool {
	return len(proposal.signature) > 0
}

func (proposal *Proposal) sign() error {
	if proposal.isSigned() {
		return nil
	}

	digest, err := proposal.Digest()
	if err != nil {
		return err
	}

	proposal.signature, err = proposal.signingID.Sign(digest)
	if err != nil {
		return err
	}

	return nil
}
