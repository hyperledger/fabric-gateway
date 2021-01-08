/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// Server represents the GRPC server for the Gateway
type Server struct {
	registry Registry
}

// Registry represents the current network topology
type Registry interface {
	GetEndorsers(channel string, chaincode string) []peer.EndorserClient
	GetDeliverers(channel string) []peer.DeliverClient
	GetOrderers(channel string) []orderer.AtomicBroadcast_BroadcastClient
	ListenForTxEvents(channel string, txid string, done chan<- bool) error
}

// NewGatewayServer creates a server side implementation of the gateway server grpc
func NewGatewayServer(registry Registry) (*Server, error) {
	result := &Server{
		registry,
	}
	return result, nil
}
