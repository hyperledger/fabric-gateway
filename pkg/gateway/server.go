/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"crypto/tls"
	"fmt"

	"github.com/hyperledger/fabric-protos-go/discovery"
	fabutil "github.com/hyperledger/fabric/common/util"
	"github.com/pkg/errors"
)

type GatewayServer struct {
	discoveryAuth *discovery.AuthInfo
	registry      *registry
	gatewaySigner *Signer
}

type Config interface {
	BootstrapPeer() PeerEndpoint
	MspID() string
	Certificate() string
	Key() string
}

type PeerEndpoint struct {
	Host    string
	Port    uint32
	TLSCert []byte
}

// NewGatewayServer creates a server side implementation of the gateway server grpc
func NewGatewayServer(config Config) (*GatewayServer, error) {

	signer, err := CreateSigner(
		config.MspID(),
		config.Certificate(),
		config.Key(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gateway identity")
	}

	clientTLSCert, err := tls.X509KeyPair([]byte(config.Certificate()), []byte(config.Key()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tls cert")
	}
	clientID, err := signer.Serialize()
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize gateway id")
	}
	authInfo := &discovery.AuthInfo{
		ClientIdentity:    clientID,
		ClientTlsCertHash: fabutil.ComputeSHA256(clientTLSCert.Certificate[0]),
	}

	registry := newRegistry(signer)

	// seed the registry with the 'bootstrap peer' for invoking discovery
	registry.addMSP(config.MspID(), config.BootstrapPeer().TLSCert)

	registry.addPeer("mychannel", config.MspID(), config.BootstrapPeer().Host, config.BootstrapPeer().Port)

	url := fmt.Sprintf("%s:%d", config.BootstrapPeer().Host, config.BootstrapPeer().Port)
	discoveryClient := registry.peers[url].discoveryClient

	// hmm, need to know a channel name first
	chDiscovery := newChannelDiscovery("mychannel", discoveryClient, signer, authInfo, registry)
	err = chDiscovery.discoverConfig()
	if err != nil {
		fmt.Printf("ERROR discovering config: %s\n", err)
	}
	err = chDiscovery.discoverPeers()
	if err != nil {
		fmt.Printf("ERROR discovering peers: %s\n", err)
	}

	return &GatewayServer{
		authInfo,
		registry,
		signer,
	}, nil
}
