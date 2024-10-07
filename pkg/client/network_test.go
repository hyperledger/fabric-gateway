// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func AssertNewTestNetwork(t *testing.T, networkName string, options ...ConnectOption) *Network {
	gateway := AssertNewTestGateway(t, options...)
	return gateway.GetNetwork(networkName)
}

func TestNetwork(t *testing.T) {
	t.Run("GetContract returns correctly named Contract", func(t *testing.T) {
		chaincodeName := "chaincode"
		network := AssertNewTestNetwork(t, "network")

		contract := network.GetContract(chaincodeName)

		require.NotNil(t, contract)
		require.Equal(t, chaincodeName, contract.ChaincodeName(), "chaincode name")
		require.Equal(t, "", contract.ContractName(), "contract name")
	})

	t.Run("GetContractWithName returns correctly named Contract", func(t *testing.T) {
		chaincodeName := "chaincode"
		contractName := "contract"
		network := AssertNewTestNetwork(t, "network")

		contract := network.GetContractWithName(chaincodeName, contractName)

		require.NotNil(t, contract)
		require.Equal(t, chaincodeName, contract.ChaincodeName(), "chaincode name")
		require.Equal(t, contractName, contract.ContractName(), "contract name")
	})
}
