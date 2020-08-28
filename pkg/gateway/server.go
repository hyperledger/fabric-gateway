/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"crypto/tls"
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go/discovery"
	fabutil "github.com/hyperledger/fabric/common/util"
	"github.com/pkg/errors"
)

type GatewayServer struct {
	discoveryAuth *discovery.AuthInfo
	registry      *registry
	gatewaySigner *signingIdentity
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
	certificate, err := identity.CertificateFromPEM([]byte(config.Certificate()))
	if err != nil {
		return nil, err
	}

	id, err := identity.NewX509Identity(config.MspID(), certificate)
	if err != nil {
		return nil, err
	}

	privateKey, err := identity.PrivateKeyFromPEM([]byte(config.Key()))
	if err != nil {
		return nil, err
	}

	signer, err := identity.NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, err
	}

	signingIdentity, err := newSigningIdentity(id, signer)
	if err != nil {
		return nil, err
	}

	clientTLSCert, err := tls.X509KeyPair([]byte(config.Certificate()), []byte(config.Key()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create tls cert")
	}
	clientID, err := signingIdentity.Serialize()
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize gateway id")
	}
	authInfo := &discovery.AuthInfo{
		ClientIdentity:    clientID,
		ClientTlsCertHash: fabutil.ComputeSHA256(clientTLSCert.Certificate[0]),
	}

	registry := newRegistry(signingIdentity)

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

	result := &GatewayServer{
		authInfo,
		registry,
		signingIdentity,
	}
	return result, nil
}
