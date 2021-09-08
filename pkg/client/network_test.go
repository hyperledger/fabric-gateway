/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func AssertNewTestNetwork(t *testing.T, networkName string, options ...ConnectOption) *Network {
	gateway := AssertNewTestGateway(t, options...)
	return gateway.GetNetwork(networkName)
}

func TestNetwork(t *testing.T) {
	t.Run("GetContract returns correctly named Contract", func(t *testing.T) {
		chaincodeID := "chaincode"
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		network := AssertNewTestNetwork(t, "network", WithClient(mockClient))

		contract := network.GetContract(chaincodeID)

		require.NotNil(t, contract)
		require.Equal(t, chaincodeID, contract.ChaincodeID(), "chaincodeID")
		require.Equal(t, "", contract.Name(), "name")
	})

	t.Run("GetContractWithName returns correctly named Contract", func(t *testing.T) {
		chaincodeID := "chaincode"
		contractName := "contract"
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		network := AssertNewTestNetwork(t, "network", WithClient(mockClient))

		contract := network.GetContractWithName(chaincodeID, contractName)

		require.NotNil(t, contract)
		require.Equal(t, chaincodeID, contract.ChaincodeID(), "chaincodeID")
		require.Equal(t, contractName, contract.Name(), "name")
	})
}
