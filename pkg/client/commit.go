/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/gateway"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// Commit provides access to a committed transaction.
type Commit struct {
	client        gateway.GatewayClient
	signingID     *signingIdentity
	transactionID string
	signedRequest *gateway.SignedCommitStatusRequest
}

func newCommit(
	client gateway.GatewayClient,
	signingID *signingIdentity,
	transactionID string,
	signedRequest *gateway.SignedCommitStatusRequest,
) *Commit {
	return &Commit{
		client:        client,
		signingID:     signingID,
		transactionID: transactionID,
		signedRequest: signedRequest,
	}
}

// Bytes of the serialized commit.
func (commit *Commit) Bytes() ([]byte, error) {
	requestBytes, err := proto.Marshal(commit.signedRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshall SignedCommitStatusRequest protobuf: %w", err)
	}

	return requestBytes, nil
}

// Digest of the commit status request. This is used to generate a digital signature.
func (commit *Commit) Digest() []byte {
	return commit.signingID.Hash(commit.signedRequest.GetRequest())
}

// TransactionID of the transaction.
func (commit *Commit) TransactionID() string {
	return commit.transactionID
}

// Status of the committed transaction. If the transaction has not yet committed, this call blocks until the commit
// occurs.
func (commit *Commit) Status(ctx context.Context) (*Status, error) {
	if err := commit.sign(); err != nil {
		return nil, err
	}

	response, err := commit.client.CommitStatus(ctx, commit.signedRequest)
	if err != nil {
		return nil, err
	}

	status := &Status{
		Code:          response.GetResult(),
		Successful:    response.GetResult() == peer.TxValidationCode_VALID,
		TransactionID: commit.transactionID,
		BlockNumber:   response.GetBlockNumber(),
	}
	return status, nil
}

func (commit *Commit) sign() error {
	if commit.isSigned() {
		return nil
	}

	digest := commit.Digest()
	signature, err := commit.signingID.Sign(digest)
	if err != nil {
		return err
	}

	commit.setSignature(signature)

	return nil
}

func (commit *Commit) isSigned() bool {
	return len(commit.signedRequest.GetSignature()) > 0
}

func (commit *Commit) setSignature(signature []byte) {
	commit.signedRequest.Signature = signature
}

// Status of a committed transaction.
type Status struct {
	Code          peer.TxValidationCode
	Successful    bool
	TransactionID string
	BlockNumber   uint64
}
