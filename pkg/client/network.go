/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

// Network represents a blockchain network, or Channel
type Network struct {
	gateway *Gateway
	name    string
}

// GetContract returns a smart contract
func (nw *Network) GetContract(name string) *Contract {
	return &Contract{
		network: nw,
		name:    name,
	}
}
