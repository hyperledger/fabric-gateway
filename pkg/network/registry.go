/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package network

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/discovery"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	fabutil "github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// Config represents the startup configuration of the gateway
type Config interface {
	BootstrapPeers() []string
	MspID() string
	Certificate() []byte
	Key() []byte
	TLSRootCert() []byte
}

type registry struct {
	signer     *signingIdentity
	peers      map[string]peerClient
	orderers   map[string]ordererClient
	msps       map[string]mspInfo
	channels   map[string]channelInfo
	discoverer *channelDiscovery
}

type peerClient struct {
	endpoint        endpoint
	endorserClient  peer.EndorserClient
	deliverClient   peer.DeliverClient
	discoveryClient discovery.DiscoveryClient
}

type ordererClient struct {
	endpoint        endpoint
	broadcastClient orderer.AtomicBroadcast_BroadcastClient
}

type mspInfo struct {
	tlsRootCert []byte
	orderers    map[string]bool
	peers       map[string]bool
}

type channelInfo struct {
	mspid    string
	orderers map[string]bool
	peers    map[string]bool
	refresh  bool
}

type endpoint struct {
	host             string
	port             uint32
	hostnameOverride string
}

func NewRegistry(config Config) (*registry, error) {
	certificate, err := identity.CertificateFromPEM(config.Certificate())
	if err != nil {
		return nil, err
	}

	id, err := identity.NewX509Identity(config.MspID(), certificate)
	if err != nil {
		return nil, err
	}

	privateKey, err := identity.PrivateKeyFromPEM(config.Key())
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

	reg := &registry{
		signer:   signingIdentity,
		peers:    make(map[string]peerClient),
		orderers: make(map[string]ordererClient),
		msps:     make(map[string]mspInfo),
		channels: make(map[string]channelInfo),
	}

	parts := strings.Split(config.BootstrapPeers()[0], ":")
	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	reg.addMSP(config.MspID(), config.TLSRootCert())
	reg.addPeer("", config.MspID(), host, uint32(port))

	clientTLSCert, err := tls.X509KeyPair(config.Certificate(), config.Key())
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
	url := fmt.Sprintf("%s:%d", host, port)
	discoveryClient := reg.peers[url].discoveryClient

	reg.discoverer = newChannelDiscovery(discoveryClient, signer, authInfo, reg)
	return reg, nil
}

func (reg *registry) addPeer(channel string, mspid string, host string, port uint32) error {
	fmt.Printf("addPeer: %s, %s, %s, %d\n", channel, mspid, host, port)
	// assumes that the msp registry has already been populated with this mspid
	url := fmt.Sprintf("%s:%d", host, port)
	if _, ok := reg.peers[url]; !ok {
		// this peer is new - connect to it and add to the peers registry
		ep := endpoint{host, port, host}
		tlscert := reg.msps[mspid].tlsRootCert
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(tlscert) {
			return errors.New("Failed to append certificate to client credentials")
		}
		creds := credentials.NewClientTLSFromCert(certPool, host)
		conn, err := grpc.Dial(translateURL(url), grpc.WithTransportCredentials(creds))
		if err != nil {
			return errors.Wrap(err, "failed to connect to peer: ")
		}
		ec := peer.NewEndorserClient(conn)
		// query channels for this peer
		channels, err := reg.getChannels(ec)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(url, channels)

		reg.peers[url] = peerClient{
			endpoint:        ep,
			endorserClient:  ec,
			deliverClient:   peer.NewDeliverClient(conn),
			discoveryClient: discovery.NewDiscoveryClient(conn),
		}
	}
	// add a reference to the msp registry
	reg.msps[mspid].peers[url] = true
	// add a reference to the channel registry
	if _, ok := reg.channels[channel]; !ok {
		reg.channels[channel] = channelInfo{
			mspid:    mspid,
			peers:    make(map[string]bool),
			orderers: make(map[string]bool),
			refresh:  true,
		}
	}
	reg.channels[channel].peers[url] = true

	return nil
}

func (reg *registry) addOrderer(channel string, mspid string, host string, port uint32) error {
	fmt.Printf("addOrderer: %s, %s, %s, %d\n", channel, mspid, host, port)
	// assumes that the msp registry has already been populated with this mspid
	url := fmt.Sprintf("%s:%d", host, port)
	if _, ok := reg.orderers[url]; !ok {
		// this peer is new - connect to it and add to the orderers registry
		ep := endpoint{host, port, host}
		tlscert := reg.msps[mspid].tlsRootCert
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(tlscert) {
			return errors.New("Failed to append certificate to client credentials")
		}
		creds := credentials.NewClientTLSFromCert(certPool, host)
		conn, err := grpc.Dial(translateURL(url), grpc.WithTransportCredentials(creds))
		if err != nil {
			return err
		}
		broadcastClient, err := orderer.NewAtomicBroadcastClient(conn).Broadcast(context.TODO())
		if err != nil {
			rpcStatus, ok := status.FromError(err)
			if ok {
				fmt.Println(rpcStatus.Message())
			}
			return errors.Wrap(err, "failed to connect to orderer: ")
		}
		reg.orderers[url] = ordererClient{
			endpoint:        ep,
			broadcastClient: broadcastClient,
		}
	}
	// add a reference to the msp registry
	reg.msps[mspid].orderers[url] = true
	// add a reference to the channel registry
	if _, ok := reg.channels[channel]; !ok {
		reg.channels[channel] = channelInfo{
			mspid:    mspid,
			peers:    make(map[string]bool),
			orderers: make(map[string]bool),
			refresh:  true,
		}
	}
	reg.channels[channel].orderers[url] = true

	return nil
}

func (reg *registry) addMSP(mspid string, cert []byte) {
	// nothing safe about this
	reg.msps[mspid] = mspInfo{
		tlsRootCert: cert,
		peers:       make(map[string]bool),
		orderers:    make(map[string]bool),
	}
}

func (reg *registry) GetEndorsers(channel string) []peer.EndorserClient {
	reg.discoverChannel(channel)
	// at the moment this returns all endorsing peers in a channel
	// eventually this should return a chaincode specific set
	endorsers := make([]peer.EndorserClient, 0)
	for p := range reg.channels[channel].peers {
		endorsers = append(endorsers, reg.peers[p].endorserClient)
	}
	return endorsers
}

func (reg *registry) GetDeliverers(channel string) []peer.DeliverClient {
	reg.discoverChannel(channel)
	// at the moment this returns all endorsing peers in a channel
	// eventually this should return a chaincode specific set
	deliverers := make([]peer.DeliverClient, 0)
	for p := range reg.channels[channel].peers {
		deliverers = append(deliverers, reg.peers[p].deliverClient)
	}
	return deliverers
}

func (reg *registry) GetOrderers(channel string) []orderer.AtomicBroadcast_BroadcastClient {
	reg.discoverChannel(channel)
	// at the moment this returns all endorsing peers in a channel
	// eventually this should return a chaincode specific set
	orderers := make([]orderer.AtomicBroadcast_BroadcastClient, 0)
	for o := range reg.channels[channel].orderers {
		orderers = append(orderers, reg.orderers[o].broadcastClient)
	}
	return orderers
}

func (reg *registry) discoverChannel(channel string) error {
	fmt.Printf("discoverChannel: %s\n", channel)
	fmt.Println(reg.channels)
	refresh := false
	if _, ok := reg.channels[channel]; !ok {
		refresh = true
	} else {
		refresh = reg.channels[channel].refresh
	}
	if refresh {
		fmt.Println("invoking discovery")
		err := reg.discoverer.discoverConfig(channel)
		if err != nil {
			return err
		}
		err = reg.discoverer.discoverPeers(channel)
		if err != nil {
			return err
		}
		info := reg.channels[channel]
		info.refresh = false
		return nil
	}
	return nil
}

func translateURL(url string) string {
	if os.Getenv("DISCOVERY_AS_LOCALHOST") == "TRUE" {
		parts := strings.Split(url, ":")
		return "localhost:" + parts[1]
	}
	return url
}

func (reg *registry) getChannels(target peer.EndorserClient) ([]string, error) {
	// create invocation spec to target a chaincode with arguments
	ccis := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			ChaincodeId: &peer.ChaincodeID{Name: "cscc"},
			Input:       &peer.ChaincodeInput{Args: [][]byte{[]byte("GetChannels")}},
		},
	}

	creator, err := reg.signer.Serialize()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to serialize Signer: ")
	}

	proposal, _, err := protoutil.CreateChaincodeProposal(
		common.HeaderType_ENDORSER_TRANSACTION,
		"",
		ccis,
		creator,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create chaincode proposal")
	}

	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal chaincode proposal")
	}

	signature, err := reg.signer.Sign(proposalBytes)
	if err != nil {
		return nil, err
	}

	signedProposal := &peer.SignedProposal{
		ProposalBytes: proposalBytes,
		Signature:     signature,
	}

	response, err := target.ProcessProposal(context.TODO(), signedProposal)
	if err != nil {
		return nil, err
	}

	cqr := &peer.ChannelQueryResponse{}
	err = proto.Unmarshal(response.GetResponse().Payload, cqr)
	if err != nil {
		return nil, err
	}

	channelNames := make([]string, 0)

	for _, channel := range cqr.Channels {
		channelNames = append(channelNames, channel.ChannelId)
		fmt.Println(channel.ChannelId)
	}

	return channelNames, nil
}
