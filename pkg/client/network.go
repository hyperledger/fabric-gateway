/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
)

// Network represents a blockchain network, or Fabric channel. The Network can be used to access deployed smart
// contracts, and to listen for events emitted when blocks are committed to the ledger.
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
func (network *Network) ChaincodeEvents(ctx context.Context, chaincodeName string, options ...ChaincodeEventsOption) (<-chan *ChaincodeEvent, error) {
	events, err := network.NewChaincodeEventsRequest(chaincodeName, options...)
	if err != nil {
		return nil, err
	}

	return events.Events(ctx)
}

// NewChaincodeEventsRequest creates a request to read events emitted by the specified chaincode. Supports off-line
// signing flow.
func (network *Network) NewChaincodeEventsRequest(chaincodeName string, options ...ChaincodeEventsOption) (*ChaincodeEventsRequest, error) {
	builder := &chaincodeEventsBuilder{
		client:        network.client,
		signingID:     network.signingID,
		channelName:   network.name,
		chaincodeName: chaincodeName,
	}

	for _, option := range options {
		if err := option(builder); err != nil {
			return nil, err
		}
	}

	return builder.build()
}
