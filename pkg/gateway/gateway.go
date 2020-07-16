/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"crypto/tls"
	"io/ioutil"
	"path/filepath"

	"github.com/hyperledger/fabric-gateway/pkg/util"
	"github.com/hyperledger/fabric-protos-go/discovery"
	fabutil "github.com/hyperledger/fabric/common/util"
	"github.com/pkg/errors"
)

type GatewayServer struct {
	discoveryAuth *discovery.AuthInfo
	registry      *registry
	gatewaySigner *Signer
}

// NewGatewayServer creates a server side implementation of the gateway server grpc
func NewGatewayServer() (*GatewayServer, error) {
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

	clientTLSCert, err := tls.X509KeyPair([]byte(id.Credentials.Certificate), []byte(id.Credentials.Key))
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

	registry := newRegistry()

	// seed the registry with the 'bootstrap peer' for invoking discovery
	// read the TLS root cert for the gateway's org
	pemPath := filepath.Join(
		"..",
		"..",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"tlsca",
		"tlsca.org1.example.com-cert.pem",
	)
	pem, err := ioutil.ReadFile(pemPath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read TLS cert")
	}
	registry.addMSP("Org1MSP", pem)

	registry.addPeer("mychannel", "Org1MSP", "peer0.org1.example.com", 7051)

	discoveryClient := registry.peers["peer0.org1.example.com:7051"].discoveryClient

	// hmm, need to know a channel name first
	chDiscovery := newChannelDiscovery("mychannel", discoveryClient, signer, authInfo, registry)
	chDiscovery.discoverConfig()
	chDiscovery.discoverPeers()

	return &GatewayServer{
		authInfo,
		registry,
		signer,
	}, nil
}
