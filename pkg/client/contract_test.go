/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"testing"
)

func AssertNewTestContract(t *testing.T, chaincodeName string, options ...ConnectOption) *Contract {
	network := AssertNewTestNetwork(t, "network", options...)
	return network.GetContract(chaincodeName)
}

func AssertNewTestContractWithName(t *testing.T, chaincodeName string, contractName string, options ...ConnectOption) *Contract {
	network := AssertNewTestNetwork(t, "network", options...)
	return network.GetContractWithName(chaincodeName, contractName)
}

func bytesAsStrings(bytes [][]byte) []string {
	results := make([]string, 0, len(bytes))

	for _, v := range bytes {
		results = append(results, string(v))
	}

	return results
}
