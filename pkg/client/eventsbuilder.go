/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"github.com/hyperledger/fabric-protos-go/orderer"
)

type eventsBuilder struct {
	client             *gatewayClient
	signingID          *signingIdentity
	channelName        string
	startPosition      *orderer.SeekPosition
	afterTransactionID string
}

func (builder *eventsBuilder) getStartPosition() *orderer.SeekPosition {
	if builder.startPosition != nil {
		return builder.startPosition
	}

	return &orderer.SeekPosition{
		Type: &orderer.SeekPosition_NextCommit{
			NextCommit: &orderer.SeekNextCommit{},
		},
	}
}

type eventOption = func(builder *eventsBuilder) error

// Checkpoint provides the current position for event processing.
type Checkpoint interface {
	// BlockNumber in which the next event is expected.
	BlockNumber() uint64
	// TransactionID of the last successfully processed event within the current block.
	TransactionID() string
}

// Checkpointer allows update of a checkpoint position after events are successfully processed.
type Checkpointer interface {
	// CheckpointBlock checkpoints the block number.
	CheckpointBlock(uint64) error
	// CheckpointTransaction checkpoints the transaction within a block.
	CheckpointTransaction(uint64, string) error
	// CheckpointChaincodeEvent checkpoints the chaincode event.
	CheckpointChaincodeEvent(*ChaincodeEvent) error

	Checkpoint
}

// WithStartBlock reads events starting at the specified block number.
func WithStartBlock(blockNumber uint64) eventOption {
	return func(builder *eventsBuilder) error {
		builder.startPosition = &orderer.SeekPosition{
			Type: &orderer.SeekPosition_Specified{
				Specified: &orderer.SeekSpecified{
					Number: blockNumber,
				},
			},
		}
		return nil
	}
}

// WithCheckpointer reads events starting at the checkpoint position.
func WithCheckpointer(checkpoint Checkpoint) eventOption {

	return func(builder *eventsBuilder) error {
		blockNumber := checkpoint.BlockNumber()
		transactionID := checkpoint.TransactionID()

		if blockNumber == 0 && transactionID == "" {
			return nil
		}
		builder.startPosition = &orderer.SeekPosition{
			Type: &orderer.SeekPosition_Specified{
				Specified: &orderer.SeekSpecified{
					Number: blockNumber,
				},
			},
		}
		builder.afterTransactionID = transactionID
		return nil
	}
}
