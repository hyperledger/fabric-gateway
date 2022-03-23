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

// Checkpointer provides the current position for event processing.
type Checkpointer interface {
	// BlockNumber in which the next event is expected.
	BlockNumber() uint64
	// TransactionID of the last successfully processed event within the current block.
	TransactionID() string
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

// WithCheckpointer reads events starting at the position recorded by the checkpointer.
func WithCheckpointer(checkpointer Checkpointer) eventOption {

	return func(builder *eventsBuilder) error {
		blockNumber := checkpointer.BlockNumber()
		transactionID := checkpointer.TransactionID()

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
