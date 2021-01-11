/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/golang/protobuf/proto"
	gateway "github.com/hyperledger/fabric-gateway/protos"
)

// Transaction represents an endorsed transaction that can be submitted to the orderer for commit to the ledger.
type Transaction struct {
	client              gateway.GatewayClient
	signingID           *signingIdentity
	preparedTransaction *gateway.PreparedTransaction
}

// Result of the proposed transaction invocation.
func (transaction *Transaction) Result() []byte {
	return transaction.preparedTransaction.Response.Value
}

// Bytes of the serialized transaction.
func (transaction *Transaction) Bytes() ([]byte, error) {
	transactionBytes, err := proto.Marshal(transaction.preparedTransaction)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshall Proposal protobuf: %w", err)
	}

	return transactionBytes, nil
}

// Digest of the transaction. This is used to generate a digital signature.
func (transaction *Transaction) Digest() ([]byte, error) {
	return transaction.signingID.Hash(transaction.preparedTransaction.Envelope.Payload)
}

// Submit the transaction to the orderer for commit to the ledger.
func (transaction *Transaction) Submit() (chan error, error) {
	if err := transaction.sign(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

	stream, err := transaction.client.Submit(ctx, transaction.preparedTransaction)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("Failed to submit transaction to the orderer: %w", err)
	}

	commit := make(chan error)
	go func() {
		defer cancel()
		for {
			event, err := stream.Recv()
			if err == io.EOF {
				commit <- nil
				return
			}
			if err != nil {
				commit <- fmt.Errorf("Failed to receive event: %w", err)
				return
			}
			fmt.Println(event)
		}
	}()

	return commit, nil
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
