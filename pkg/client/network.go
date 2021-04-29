/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
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
