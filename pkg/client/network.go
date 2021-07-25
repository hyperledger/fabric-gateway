/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/gateway"
)

// Network represents a blockchain network, or Fabric channel. The Network can be used to access deployed smart
// contracts, and to listen for events emitted when blocks are committed to the ledger.
type Network struct {
	client    gateway.GatewayClient
	signingID *signingIdentity
	name      string
}

// Name of the Fabric channel this network represents.
func (network *Network) Name() string {
	return network.name
}

// GetContract returns a Contract representing the default smart contract for the named chaincode.
func (network *Network) GetContract(chaincodeID string) *Contract {
	return network.GetContractWithName(chaincodeID, "")
}

// GetContractWithName returns a Contract representing a smart contract within a named chaincode.
func (network *Network) GetContractWithName(chaincodeID string, contractName string) *Contract {
	return &Contract{
		client:       network.client,
		signingID:    network.signingID,
		channelName:  network.name,
		chaincodeID:  chaincodeID,
		contractName: contractName,
	}
}

// NewSignedCommit creates an commit with signature, which can be used to access a committed transaction.
func (network *Network) NewSignedCommit(bytes []byte, signature []byte) (*Commit, error) {
	signedRequest := &gateway.SignedCommitStatusRequest{}
	if err := proto.Unmarshal(bytes, signedRequest); err != nil {
		return nil, fmt.Errorf("failed to deserialize signed commit status request: %w", err)
	}

	request := &gateway.CommitStatusRequest{}
	if err := proto.Unmarshal(signedRequest.Request, request); err != nil {
		return nil, fmt.Errorf("failed to deserialize commit status request: %w", err)
	}

	commit := newCommit(network.client, network.signingID, request.TransactionId, signedRequest)
	commit.setSignature(signature)

	return commit, nil
}

// ChaincodeEvents returns a channel from which chaincode events emitted by transaction functions in the specified
// chaincode can be read.
func (network *Network) ChaincodeEvents(ctx context.Context, chaincodeID string) (<-chan *ChaincodeEvent, error) {
	events, err := network.NewChaincodeEventsRequest(chaincodeID)
	if err != nil {
		return nil, err
	}

	return events.Events(ctx)
}

// NewChaincodeEventsRequest creates a request to read events emitted by the specified chaincode.
func (network *Network) NewChaincodeEventsRequest(chaincodeID string) (*ChaincodeEventsRequest, error) {
	request, err := network.newSignedChaincodeEventsRequestProto(chaincodeID)
	if err != nil {
		return nil, err
	}

	result := &ChaincodeEventsRequest{
		client:        network.client,
		signingID:     network.signingID,
		signedRequest: request,
	}
	return result, nil
}

func (network *Network) newSignedChaincodeEventsRequestProto(chaincodeID string) (*gateway.SignedChaincodeEventsRequest, error) {
	request, err := network.newChaincodeEventsRequestProto(chaincodeID)
	if err != nil {
		return nil, err
	}

	requestBytes, err := proto.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize chaincode events request: %w", err)
	}

	signedRequest := &gateway.SignedChaincodeEventsRequest{
		Request: requestBytes,
	}
	return signedRequest, nil
}

func (network *Network) newChaincodeEventsRequestProto(chaincodeID string) (*gateway.ChaincodeEventsRequest, error) {
	creator, err := network.signingID.Creator()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize identity: %w", err)
	}

	request := &gateway.ChaincodeEventsRequest{
		ChannelId:   network.name,
		Identity:    creator,
		ChaincodeId: chaincodeID,
	}
	return request, nil
}

// NewSignedChaincodeEventsRequest creates a signed request to read events emitted by a specific chaincode.
func (network *Network) NewSignedChaincodeEventsRequest(bytes []byte, signature []byte) (*ChaincodeEventsRequest, error) {
	request := &gateway.SignedChaincodeEventsRequest{}
	if err := proto.Unmarshal(bytes, request); err != nil {
		return nil, fmt.Errorf("failed to deserialize signed chaincode events request: %w", err)
	}

	result := &ChaincodeEventsRequest{
		client:        network.client,
		signingID:     network.signingID,
		signedRequest: request,
	}
	return result, nil
}
