/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"testing"

	"github.com/golang/mock/gomock"
)

func AssertNewTestNetwork(t *testing.T, networkName string, options ...ConnectOption) *Network {
	gateway := AssertNewTestGateway(t, options...)
	return gateway.GetNetwork(networkName)
}

func TestNetwork(t *testing.T) {
	t.Run("GetContract returns correctly named Contract", func(t *testing.T) {
		chaincodeID := "chaincode"
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		network := AssertNewTestNetwork(t, "network", WithClient(mockClient))

		contract := network.GetContract(chaincodeID)

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

	t.Run("GetContractWithName returns correctly named Contract", func(t *testing.T) {
		chaincodeID := "chaincode"
		contractName := "contract"
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		network := AssertNewTestNetwork(t, "network", WithClient(mockClient))

		contract := network.GetContractWithName(chaincodeID, contractName)

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
