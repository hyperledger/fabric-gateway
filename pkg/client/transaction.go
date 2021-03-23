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
	"github.com/hyperledger/fabric-protos-go/peer"
)

// Transaction represents an endorsed transaction that can be submitted to the orderer for commit to the ledger.
type Transaction struct {
	client              gateway.GatewayClient
	signingID           *signingIdentity
	channelID           string
	preparedTransaction *gateway.PreparedTransaction
}

// Result of the proposed transaction invocation.
func (transaction *Transaction) Result() []byte {
	return transaction.preparedTransaction.Result.Payload
}

// Bytes of the serialized transaction.
func (transaction *Transaction) Bytes() ([]byte, error) {
	transactionBytes, err := proto.Marshal(transaction.preparedTransaction)
	if err != nil {
		return nil, fmt.Errorf("failed to marshall Proposal protobuf: %w", err)
	}

	return transactionBytes, nil
}

// Digest of the transaction. This is used to generate a digital signature.
func (transaction *Transaction) Digest() ([]byte, error) {
	return transaction.signingID.Hash(transaction.preparedTransaction.Envelope.Payload)
}

// Submit the transaction to the orderer for commit to the ledger.
func (transaction *Transaction) Submit() ([]byte, error) {
	if err := transaction.sign(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	submitRequest := &gateway.SubmitRequest{
		TransactionId:       transaction.preparedTransaction.TransactionId,
		ChannelId:           transaction.channelID,
		PreparedTransaction: transaction.preparedTransaction.Envelope,
	}
	_, err := transaction.client.Submit(ctx, submitRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to submit transaction to the orderer: %w", err)
	}

	statusRequest := &gateway.CommitStatusRequest{
		ChannelId:     transaction.channelID,
		TransactionId: transaction.preparedTransaction.TransactionId,
	}
	status, err := transaction.client.CommitStatus(ctx, statusRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain transaction commit status: %w", err)
	}

	if status.Result != peer.TxValidationCode_VALID {
		return nil, fmt.Errorf("transaction commit failed with status: %v", peer.TxValidationCode_name[int32(status.Result)])
	}

	return transaction.preparedTransaction.Result.Payload, nil
}

func (transaction *Transaction) sign() error {
	if transaction.isSigned() {
		return nil
	}

	digest, err := transaction.Digest()
	if err != nil {
		return err
	}

	signature, err := transaction.signingID.Sign(digest)
	if err != nil {
		return err
	}

	transaction.setSignature(signature)

	return nil
}

func (transaction *Transaction) isSigned() bool {
	return len(transaction.preparedTransaction.Envelope.Signature) > 0
}

func (transaction *Transaction) setSignature(signature []byte) {
	transaction.preparedTransaction.Envelope.Signature = signature
}
