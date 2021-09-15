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
	"github.com/hyperledger/fabric-protos-go/gateway"
)

// Proposal represents a transaction proposal that can be sent to peers for endorsement or evaluated as a query.
type Proposal struct {
	client              gateway.GatewayClient
	signingID           *signingIdentity
	channelID           string
	proposedTransaction *gateway.ProposedTransaction
}

// Bytes of the serialized proposal message.
func (proposal *Proposal) Bytes() ([]byte, error) {
	transactionBytes, err := proto.Marshal(proposal.proposedTransaction)
	if err != nil {
		return nil, fmt.Errorf("failed to marshall Proposal protobuf: %w", err)
	}

	return transactionBytes, nil
}

// Digest of the proposal. This is used to generate a digital signature.
func (proposal *Proposal) Digest() []byte {
	return proposal.signingID.Hash(proposal.proposedTransaction.Proposal.ProposalBytes)
}

// TransactionID for the proposal.
func (proposal *Proposal) TransactionID() string {
	return proposal.proposedTransaction.GetTransactionId()
}

// Endorse the proposal to obtain an endorsed transaction for submission to the orderer.
func (proposal *Proposal) Endorse() (*Transaction, error) {
	if err := proposal.sign(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	endorseRequest := &gateway.EndorseRequest{
		TransactionId:          proposal.proposedTransaction.GetTransactionId(),
		ChannelId:              proposal.channelID,
		ProposedTransaction:    proposal.proposedTransaction.GetProposal(),
		EndorsingOrganizations: proposal.proposedTransaction.GetEndorsingOrganizations(),
	}
	response, err := proposal.client.Endorse(ctx, endorseRequest)
	if err != nil {
		return nil, err
	}

	preparedTransaction := &gateway.PreparedTransaction{
		TransactionId: proposal.proposedTransaction.GetTransactionId(),
		Envelope:      response.GetPreparedTransaction(),
		Result:        response.GetResult(),
	}
	result := &Transaction{
		client:              proposal.client,
		signingID:           proposal.signingID,
		channelID:           proposal.channelID,
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

	evaluateRequest := &gateway.EvaluateRequest{
		TransactionId:       proposal.proposedTransaction.GetTransactionId(),
		ChannelId:           proposal.channelID,
		ProposedTransaction: proposal.proposedTransaction.GetProposal(),
		TargetOrganizations: proposal.proposedTransaction.GetEndorsingOrganizations(),
	}
	response, err := proposal.client.Evaluate(ctx, evaluateRequest)
	if err != nil {
		return nil, err
	}

	return response.GetResult().GetPayload(), nil
}

func (proposal *Proposal) setSignature(signature []byte) {
	proposal.proposedTransaction.Proposal.Signature = signature
}

func (proposal *Proposal) isSigned() bool {
	return len(proposal.proposedTransaction.GetProposal().GetSignature()) > 0
}

func (proposal *Proposal) sign() error {
	if proposal.isSigned() {
		return nil
	}

	digest := proposal.Digest()
	signature, err := proposal.signingID.Sign(digest)
	if err != nil {
		return err
	}

	proposal.setSignature(signature)

	return nil
}
