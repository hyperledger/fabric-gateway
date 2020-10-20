/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type mockRegistry struct{}

func NewMockRegistry() *mockRegistry {
	return &mockRegistry{}
}

func (mr *mockRegistry) GetEndorsers(channel string) []peer.EndorserClient {
	return nil
}

func (mr *mockRegistry) GetDeliverers(channel string) []peer.DeliverClient {
	return nil
}

func (mr *mockRegistry) GetOrderers(channel string) []orderer.AtomicBroadcast_BroadcastClient {
	return nil
}

func (mr *mockRegistry) ListenForTxEvents(channel string, txid string, done chan<- bool) error {
	return nil
}
