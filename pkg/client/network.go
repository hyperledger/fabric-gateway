/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

// Network represents a blockchain network, or Fabric channel. The Network can be used to access deployed smart
// contracts, and to listen for events emitted when blocks are committed to the ledger.
type Network struct {
	gateway *Gateway
	name    string
}

// GetContract returns a Contract representing the named smart contract.
func (network *Network) GetContract(name string) *Contract {
	return &Contract{
		network: network,
		name:    name,
	}
}
