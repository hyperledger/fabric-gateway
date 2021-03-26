/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-protos-go/gateway"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// Commit provides access to a committed transaction.
type Commit struct {
	client        gateway.GatewayClient
	channelID     string
	transactionID string
}

// Status of the committed transaction. If the transaction has not yet committed, this call blocks until the commit
// occurs.
func (commit *Commit) Status() (peer.TxValidationCode, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	request := &gateway.CommitStatusRequest{
		ChannelId:     commit.channelID,
		TransactionId: commit.transactionID,
	}

	status, err := commit.client.CommitStatus(ctx, request)
	if err != nil {
		return 0, fmt.Errorf("failed to obtain transaction commit status: %w", err)
	}

	return status.Result, nil
}

// TransactionID of the transaction.
func (commit *Commit) TransactionID() string {
	return commit.transactionID
}
