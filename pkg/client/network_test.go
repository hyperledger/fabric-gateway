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
	t.Run("GetContract returns correctly named Contract", func(t *testing.T) {
		contractName := "contract"
		mockClient := mock.NewGatewayClient()
		network := AssertNewTestNetwork(t, "network", WithClient(mockClient))

		contract := network.GetContract(contractName)

		if nil == contract {
			t.Fatal("Expected network, got nil")
		}
		if contract.name != contractName {
			t.Fatalf("Expected a network named %s, got %s", contractName, contract.name)
		}
	})
}
