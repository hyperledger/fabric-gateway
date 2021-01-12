/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package network

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-gateway/pkg/hash"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go/discovery"
	"github.com/hyperledger/fabric-protos-go/gossip"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type channelDiscovery struct {
	client   discovery.DiscoveryClient
	sign     identity.Sign
	hash     hash.Hash
	authInfo *discovery.AuthInfo
	registry *registry
}

func newChannelDiscovery(client discovery.DiscoveryClient, sign identity.Sign, hash hash.Hash, authInfo *discovery.AuthInfo, registry *registry) *channelDiscovery {
	return &channelDiscovery{
		client:   client,
		sign:     sign,
		hash:     hash,
		authInfo: authInfo,
		registry: registry,
	}
}

func (cd *channelDiscovery) invokeDiscovery(request *discovery.Request) (*discovery.Response, error) {
	payload, err := proto.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal discovery request: %w", err)
	}

	digest, err := cd.hash(payload)
	if err != nil {
		return nil, err
	}

	signature, err := cd.sign(digest)
	if err != nil {
		return nil, fmt.Errorf("failed to sign discovery request: %w", err)
	}

	response, err := cd.client.Discover(
		context.TODO(),
		&discovery.SignedRequest{
			Payload:   payload,
			Signature: signature,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	return response, nil

}

func (cd *channelDiscovery) discoverConfig(channel string) error {
	request := &discovery.Request{
		Authentication: cd.authInfo,
		Queries: []*discovery.Query{
			{
				Channel: channel,
				Query: &discovery.Query_ConfigQuery{
					ConfigQuery: &discovery.ConfigQuery{},
				},
			},
		},
	}

	response, err := cd.invokeDiscovery(request)
	if err != nil {
		return err
	}

	derr := response.Results[0].GetError()
	if derr != nil {
		return errors.New(derr.Content)
	}

	result := response.Results[0].GetConfigResult()

	// update the tlscerts
	for msp, info := range result.GetMsps() {
		cd.registry.addMSP(msp, info.GetTlsRootCerts()[0])
	}

	// update the orderers
	for msp, eps := range result.GetOrderers() {
		for _, ep := range eps.Endpoint {
			cd.registry.addOrderer(channel, msp, ep.Host, ep.Port)
		}
	}

	return err
}

func (cd *channelDiscovery) discoverPeers(channel string) error {
	request := &discovery.Request{
		Authentication: cd.authInfo,
		Queries: []*discovery.Query{
			{
				Channel: channel,
				Query: &discovery.Query_PeerQuery{
					PeerQuery: &discovery.PeerMembershipQuery{},
				},
			},
		},
	}

	response, err := cd.invokeDiscovery(request)
	if err != nil {
		return err
	}

	derr := response.Results[0].GetError()
	if derr != nil {
		return errors.New(derr.Content)
	}

	members := response.Results[0].GetMembers().PeersByOrg

	// update the peers
	for msp, peers := range members {
		for _, peer := range peers.Peers {
			var msg = &gossip.GossipMessage{}
			proto.Unmarshal(peer.MembershipInfo.Payload, msg)
			ep := msg.GetAliveMsg().Membership.Endpoint
			parts := strings.Split(ep, ":")
			host := parts[0]
			port, _ := strconv.Atoi(parts[1])
			cd.registry.addPeer(channel, msp, host, uint32(port))
		}
	}

	return err
}

func (cd *channelDiscovery) discoverEndorsers(channel string, chaincode string) ([]peer.EndorserClient, error) {
	endorsers := make([]peer.EndorserClient, 0)
	request := &discovery.Request{
		Authentication: cd.authInfo,
		Queries: []*discovery.Query{
			{
				Channel: channel,
				Query: &discovery.Query_CcQuery{
					CcQuery: &discovery.ChaincodeQuery{
						Interests: []*discovery.ChaincodeInterest{{
							Chaincodes: []*discovery.ChaincodeCall{{
								Name: chaincode,
							}},
						}},
					},
				},
			},
		},
	}

	response, err := cd.invokeDiscovery(request)
	if err != nil {
		return nil, err
	}

	derr := response.Results[0].GetError()
	if derr != nil {
		return nil, errors.New(derr.Content)
	}

	// only one chaincode atm
	descriptor := response.Results[0].GetCcQueryRes().GetContent()[0]

	e := descriptor.EndorsersByGroups

	fmt.Printf("Number of Layouts: %d\n", len(descriptor.Layouts))
	// choose first layout
	for i, l := range descriptor.Layouts {
		fmt.Printf("layout: %d\n", i)
		layout := l.QuantitiesByGroup
		r := make([]*discovery.Peer, 0)
		for group, quantity := range layout {
			endorsers := e[group].Peers
			fmt.Printf("group: %s, quantity: %d, available: %d\n", group, quantity, len(endorsers))
			peerGroup := endorsers[0:quantity]
			r = append(r, peerGroup...)
		}

		for _, peer := range r {
			var msg = &gossip.GossipMessage{}
			proto.Unmarshal(peer.MembershipInfo.Payload, msg)
			ep := msg.GetAliveMsg().Membership.Endpoint
			parts := strings.Split(ep, ":")
			host := parts[0]
			port, _ := strconv.Atoi(parts[1])
			url := fmt.Sprintf("%s:%d", host, port)
			if endorser, ok := cd.registry.peers[url]; ok {
				endorsers = append(endorsers, endorser.endorserClient)
			}
			fmt.Printf("host: %s, port: %d\n", host, port)
		}
	}
	return endorsers, err
}
