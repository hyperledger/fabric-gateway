/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	gateway "github.com/hyperledger/fabric-gateway/protos"
)

// Proposal represents a transaction proposal that can be sent to peers for endorsement or evaluated as a query.
type Proposal struct {
	client              gateway.GatewayClient
	signingID           *signingIdentity
	proposedTransaction *gateway.ProposedTransaction
}

// Bytes of the serialized proposal message.
func (proposal *Proposal) Bytes() ([]byte, error) {
	return proto.Marshal(proposal.proposedTransaction)
}

// Digest of the proposal. This is used to generate a digital signature.
func (proposal *Proposal) Digest() ([]byte, error) {
	return proposal.signingID.Hash(proposal.proposedTransaction.Proposal.ProposalBytes)
}

// TransactionID for the proposal.
func (proposal *Proposal) TransactionID() string {
	return proposal.proposedTransaction.TxId
}

// Endorse the proposal to obtain an endorsed transaction for submission to the orderer.
func (proposal *Proposal) Endorse() (*Transaction, error) {
	if err := proposal.sign(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	preparedTransaction, err := proposal.client.Endorse(ctx, proposal.proposedTransaction)
	if err != nil {
		return nil, fmt.Errorf("failed to endorse proposal: %w", err)
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
	if err := proposal.sign(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	result, err := proposal.client.Evaluate(ctx, proposal.proposedTransaction)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate transaction: %w", err)
	}

	return result.Value, nil
}

func (proposal *Proposal) setSignature(signature []byte) {
	proposal.proposedTransaction.Proposal.Signature = signature
}

func (proposal *Proposal) isSigned() bool {
	return len(proposal.proposedTransaction.Proposal.Signature) > 0
}

func (proposal *Proposal) sign() error {
	if proposal.isSigned() {
		return nil
	}

	digest, err := proposal.Digest()
	if err != nil {
		return err
	}

	signature, err := proposal.signingID.Sign(digest)
	if err != nil {
		return err
	}

	proposal.setSignature(signature)

	return nil
}
