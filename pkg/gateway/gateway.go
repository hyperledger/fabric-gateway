/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"path/filepath"

	"github.com/hyperledger/fabric-gateway/pkg/util"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

type GatewayServer struct {
	ccpPath         string
	endorserClients []peer.EndorserClient
	deliverClients  []peer.DeliverClient
	broadcastClient orderer.AtomicBroadcast_BroadcastClient
	gatewaySigner   *Signer
}

// NewGatewayServer creates a server side implementation of the gateway server grpc
func NewGatewayServer(
	ccpPath string,
	endorserClients []peer.EndorserClient,
	deliverClients []peer.DeliverClient,
	broadcastClient orderer.AtomicBroadcast_BroadcastClient,
) (*GatewayServer, error) {
	idfile := filepath.Join(
		"..",
		"..",
		"fabric-samples",
		"fabcar",
		"javascript",
		"wallet",
		"events.id",
	)

	id, err := util.ReadWalletIdentity(idfile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read gateway identity")
	}

	signer, err := CreateSigner(
		id.MspID,
		id.Credentials.Certificate,
		id.Credentials.Key,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gateway identity")
	}

	return &GatewayServer{
		ccpPath,
		endorserClients,
		deliverClients,
		broadcastClient,
		signer,
	}, nil
}
