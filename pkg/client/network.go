/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// Network represents a network of nodes that are members of a specific Fabric channel. The Network can be used to
// access deployed smart contracts, and to listen for events emitted when blocks are committed to the ledger. Network
// instances are obtained from a Gateway using the Gateway's GetNetwork() method.
type Network struct {
	client    *gatewayClient
	signingID *signingIdentity
	name      string
}

// Name of the Fabric channel this network represents.
func (network *Network) Name() string {
	return network.name
}

// GetContract returns a Contract representing the default smart contract for the named chaincode.
func (network *Network) GetContract(chaincodeName string) *Contract {
	return network.GetContractWithName(chaincodeName, "")
}

// GetContractWithName returns a Contract representing a smart contract within a named chaincode.
func (network *Network) GetContractWithName(chaincodeName string, contractName string) *Contract {
	return &Contract{
		client:        network.client,
		signingID:     network.signingID,
		channelName:   network.name,
		chaincodeName: chaincodeName,
		contractName:  contractName,
	}
}

// ChaincodeEvents returns a channel from which chaincode events emitted by transaction functions in the specified
// chaincode can be read.
func (network *Network) ChaincodeEvents(ctx context.Context, chaincodeName string, options ...EventOption) (<-chan *ChaincodeEvent, error) {
	events, err := network.NewChaincodeEventsRequest(chaincodeName, options...)
	if err != nil {
		return nil, err
	}

	return events.Events(ctx)
}

// NewChaincodeEventsRequest creates a request to read events emitted by the specified chaincode. Supports off-line
// signing flow.
func (network *Network) NewChaincodeEventsRequest(chaincodeName string, options ...EventOption) (*ChaincodeEventsRequest, error) {
	builder := &chaincodeEventsBuilder{
		eventsBuilder: &eventsBuilder{
			signingID:   network.signingID,
			channelName: network.name,
			client:      network.client,
		},
		chaincodeName: chaincodeName,
	}

	for _, option := range options {
		if err := option(builder.eventsBuilder); err != nil {
			return nil, err
		}
	}

	return builder.build()
}

// BlockEvents returns a channel from which block events can be read.
func (network *Network) BlockEvents(ctx context.Context, options ...EventOption) (<-chan *common.Block, error) {
	events, err := network.NewBlockEventsRequest(options...)
	if err != nil {
		return nil, err
	}

	return events.Events(ctx)
}

// NewBlockEventsRequest creates a request to read block events. Supports off-line signing flow.
func (network *Network) NewBlockEventsRequest(options ...EventOption) (*FullBlockEventsRequest, error) {
	builder := &fullBlockEventsBuilder{
		blockBuilder: &blockEventsBuilder{
			eventsBuilder: &eventsBuilder{
				signingID:   network.signingID,
				channelName: network.name,
				client:      network.client,
			},
		},
	}

	for _, option := range options {
		if err := option(builder.blockBuilder.eventsBuilder); err != nil {
			return nil, err
		}
	}

	return builder.build()
}

// FilteredBlockEvents returns a channel from which filtered block events can be read.
func (network *Network) FilteredBlockEvents(ctx context.Context, options ...EventOption) (<-chan *peer.FilteredBlock, error) {
	events, err := network.NewFilteredBlockEventsRequest(options...)
	if err != nil {
		return nil, err
	}

	return events.Events(ctx)
}

// NewFilteredBlockEventsRequest creates a request to read filtered block events. Supports off-line signing flow.
func (network *Network) NewFilteredBlockEventsRequest(options ...EventOption) (*FilteredBlockEventsRequest, error) {
	builder := &filteredBlockEventsBuilder{
		blockBuilder: &blockEventsBuilder{
			eventsBuilder: &eventsBuilder{
				signingID:   network.signingID,
				channelName: network.name,
				client:      network.client,
			},
		},
	}

	for _, option := range options {
		if err := option(builder.blockBuilder.eventsBuilder); err != nil {
			return nil, err
		}
	}

	return builder.build()
}

// FilteredBlockEvents returns a channel from which filtered block events can be read.
func (network *Network) BlockEventsWithPrivateData(ctx context.Context, options ...EventOption) (<-chan *peer.BlockAndPrivateData, error) {
	events, err := network.NewBlockEventsWithPrivateData(options...)
	if err != nil {
		return nil, err
	}

	return events.Events(ctx)
}

// NewBlockEventsWithPrivateData creates a request to read filtered block events. Supports off-line signing flow.
func (network *Network) NewBlockEventsWithPrivateData(options ...EventOption) (*BlockEventsWithPrivateDataRequest, error) {
	builder := &blockEventsWithPrivateDataBuilder{
		blockBuilder: &blockEventsBuilder{
			eventsBuilder: &eventsBuilder{
				signingID:   network.signingID,
				channelName: network.name,
				client:      network.client,
			},
		},
	}

	for _, option := range options {
		if err := option(builder.blockBuilder.eventsBuilder); err != nil {
			return nil, err
		}
	}

	return builder.build()
}
