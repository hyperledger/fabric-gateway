/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/internal/util"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type baseBlockEventsRequest struct {
	client    *gatewayClient
	signingID *signingIdentity
	request   *common.Envelope
}

// Bytes of the serialized block events request.
func (events *baseBlockEventsRequest) Bytes() ([]byte, error) {
	requestBytes, err := util.Marshal(events.request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshall Envelope protobuf: %w", err)
	}

	return requestBytes, nil
}

// Digest of the block events request. This is used to generate a digital signature.
func (events *baseBlockEventsRequest) Digest() []byte {
	return events.signingID.Hash(events.request.GetPayload())
}

func (events *baseBlockEventsRequest) sign() error {
	if events.isSigned() {
		return nil
	}

	digest := events.Digest()
	signature, err := events.signingID.Sign(digest)
	if err != nil {
		return err
	}

	events.setSignature(signature)

	return nil
}

func (events *baseBlockEventsRequest) isSigned() bool {
	return len(events.request.Signature) > 0
}

func (events *baseBlockEventsRequest) setSignature(signature []byte) {
	events.request.Signature = signature
}

// FilteredBlockEventsRequest delivers filtered block events.
type FilteredBlockEventsRequest struct {
	baseBlockEventsRequest
}

// Events returns a channel from which filtered block events can be read.
func (events *FilteredBlockEventsRequest) Events(ctx context.Context) (<-chan *peer.FilteredBlock, error) {
	if err := events.sign(); err != nil {
		return nil, err
	}

	eventsClient, err := events.client.FilteredBlockEvents(ctx, events.request)
	if err != nil {
		return nil, err
	}

	results := make(chan *peer.FilteredBlock)
	go func() {
		defer close(results)

		for {
			response, err := eventsClient.Recv()
			result := response.GetFilteredBlock()
			if err != nil || result == nil {
				return
			}

			results <- result
		}
	}()

	return results, nil
}

// BlockEventsRequest delivers block events.
type BlockEventsRequest struct {
	baseBlockEventsRequest
}

// Events returns a channel from which block events can be read.
func (events *BlockEventsRequest) Events(ctx context.Context) (<-chan *common.Block, error) {
	if err := events.sign(); err != nil {
		return nil, err
	}

	eventsClient, err := events.client.BlockEvents(ctx, events.request)
	if err != nil {
		return nil, err
	}

	results := make(chan *common.Block)
	go func() {
		defer close(results)

		for {
			response, err := eventsClient.Recv()
			result := response.GetBlock()
			if err != nil || result == nil {
				return
			}

			results <- result
		}
	}()

	return results, nil
}

// BlockAndPrivateDataEventsRequest delivers block and private data events.
type BlockAndPrivateDataEventsRequest struct {
	baseBlockEventsRequest
}

// Events returns a channel from which block and private data events can be read.
func (events *BlockAndPrivateDataEventsRequest) Events(ctx context.Context) (<-chan *peer.BlockAndPrivateData, error) {
	if err := events.sign(); err != nil {
		return nil, err
	}

	eventsClient, err := events.client.BlockAndPrivateEventsData(ctx, events.request)
	if err != nil {
		return nil, err
	}

	results := make(chan *peer.BlockAndPrivateData)
	go func() {
		defer close(results)

		for {
			response, err := eventsClient.Recv()
			result := response.GetBlockAndPrivateData()
			if err != nil || result == nil {
				return
			}

			results <- result
		}
	}()

	return results, nil
}
