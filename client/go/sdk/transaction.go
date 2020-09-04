/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/identity"
	gateway "github.com/hyperledger/fabric-gateway/protos"
	"github.com/pkg/errors"
)

type Transaction struct {
	contract            *Contract
	preparedTransaction *gateway.PreparedTransaction
}

func (transaction *Transaction) Result() []byte {
	return transaction.preparedTransaction.Response.Value
}

func (transaction *Transaction) Hash() ([]byte, error) {
	return identity.Hash(transaction.preparedTransaction.Envelope.Payload)
}

func (transaction *Transaction) Sign(signature []byte) *Transaction {
	transaction.preparedTransaction.Envelope.Signature = signature
	return transaction
}

func (transaction *Transaction) Submit() (chan error, error) {
	if err := transaction.signMessage(); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)

	stream, err := transaction.contract.network.gateway.client.Commit(ctx, transaction.preparedTransaction)
	if err != nil {
		cancel()
		return nil, errors.Wrap(err, "Failed to submit transaction to the orderer")
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
				commit <- errors.Wrap(err, "failed to receive event: ")
				return
			}
			fmt.Println(event)
		}
	}()

	return commit, nil
}

func (transaction *Transaction) signMessage() error {
	if transaction.preparedTransaction.Envelope.Signature != nil {
		return nil
	}

	digest, err := transaction.Hash()
	if err != nil {
		return err
	}

	transaction.preparedTransaction.Envelope.Signature, err = transaction.contract.network.gateway.sign(digest)
	if err != nil {
		return err
	}

	return nil
}
