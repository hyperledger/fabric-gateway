/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"testing"
)

func AssertNewTestContract(t *testing.T, chaincodeID string, options ...ConnectOption) *Contract {
	network := AssertNewTestNetwork(t, "network", options...)
	return network.GetContract(chaincodeID)
}

func AssertNewTestContractWithName(t *testing.T, chaincodeID string, contractName string, options ...ConnectOption) *Contract {
	network := AssertNewTestNetwork(t, "network", options...)
	return network.GetContractWithName(chaincodeID, contractName)
}

func bytesAsStrings(bytes [][]byte) []string {
	results := make([]string, 0, len(bytes))

	for _, v := range bytes {
		results = append(results, string(v))
	}

	return results
}
