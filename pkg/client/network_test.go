/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/internal/test/mock"
)

func AssertNewTestNetwork(t *testing.T, networkName string, options ...ConnectOption) *Network {
	gateway := AssertNewTestGateway(t, options...)
	return gateway.GetNetwork(networkName)
}

func TestNetwork(t *testing.T) {
	t.Run("GetDefaultContract returns correctly named Contract", func(t *testing.T) {
		chaincodeID := "chaincode"
		mockClient := mock.NewGatewayClient()
		network := AssertNewTestNetwork(t, "network", WithClient(mockClient))

		contract := network.GetDefaultContract(chaincodeID)

		if nil == contract {
			t.Fatal("Expected network, got nil")
		}
		if contract.chaincodeID != chaincodeID {
			t.Fatalf("Expected a network with chaincode ID %s, got %s", chaincodeID, contract.chaincodeID)
		}
		if len(contract.contractName) > 0 {
			t.Fatalf("Expected a network with empty contract name, got %s", contract.contractName)
		}
	})

	t.Run("GetDefaultContract returns correctly named Contract", func(t *testing.T) {
		chaincodeID := "chaincode"
		contractName := "contract"
		mockClient := mock.NewGatewayClient()
		network := AssertNewTestNetwork(t, "network", WithClient(mockClient))

		contract := network.GetContract(chaincodeID, contractName)

		if nil == contract {
			t.Fatal("Expected network, got nil")
		}
		if contract.chaincodeID != chaincodeID {
			t.Fatalf("Expected a network with chaincode ID %s, got %s", chaincodeID, contract.chaincodeID)
		}
		if contract.contractName != contractName {
			t.Fatalf("Expected a network with contract name %s, got %s", contractName, contract.contractName)
		}
	})
}
