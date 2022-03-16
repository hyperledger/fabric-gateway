/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"fmt"

	"github.com/hyperledger/fabric-protos-go/orderer"
)

type eventsBuilder struct {
	client        *gatewayClient
	signingID     *signingIdentity
	channelName   string
	startPosition *orderer.SeekPosition
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
type Checkpointer struct {
	StartBlock func() (uint64, bool)
	AfterTransactionID string
}

// WithCheckpointer reads events starting at the specified start block and after the transaction ID.
func WithCheckpointer(checkpointer *Checkpointer) eventOption {
	return func(builder *eventsBuilder) error {
	if(checkpointer.StartBlock != nil){
	if startBlock,ok := checkpointer.StartBlock(); ok {
		builder.startPosition = &orderer.SeekPosition{
			Type: &orderer.SeekPosition_Specified{
				Specified: &orderer.SeekSpecified{
					Number: startBlock,
				},
			},
		}
	}
		if len(checkpointer.AfterTransactionID) > 0 {
			builder.afterTransactionID = checkpointer.AfterTransactionID
		}
	}
	return nil
}
}