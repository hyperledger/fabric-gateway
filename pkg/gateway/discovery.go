/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"context"
	"strconv"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/discovery"
	"github.com/hyperledger/fabric-protos-go/gossip"
	"github.com/pkg/errors"
)

type channelDiscovery struct {
	channel  string
	client   discovery.DiscoveryClient
	signer   *Signer
	authInfo *discovery.AuthInfo
	registry *registry
}

func newChannelDiscovery(channel string, client discovery.DiscoveryClient, signer *Signer, authInfo *discovery.AuthInfo, registry *registry) *channelDiscovery {
	return &channelDiscovery{channel, client, signer, authInfo, registry}
}

func (cd *channelDiscovery) invokeDiscovery(request *discovery.Request) (*discovery.Response, error) {
	payload, err := proto.Marshal(request)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to marshal discovery request: ")
	}

	signature, err := cd.signer.Sign(payload)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to sign discovery request: ")
	}

	response, err := cd.client.Discover(
		context.TODO(),
		&discovery.SignedRequest{
			Payload:   payload,
			Signature: signature,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "Discovery failed: ")
	}

	return response, nil

}

func (cd *channelDiscovery) discoverConfig() error {
	request := &discovery.Request{
		Authentication: cd.authInfo,
		Queries: []*discovery.Query{
			&discovery.Query{
				Channel: cd.channel,
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
			cd.registry.addOrderer(cd.channel, msp, ep.Host, ep.Port)
		}
	}

	return err
}

func (cd *channelDiscovery) discoverPeers() error {
	request := &discovery.Request{
		Authentication: cd.authInfo,
		Queries: []*discovery.Query{
			&discovery.Query{
				Channel: cd.channel,
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
			cd.registry.addPeer(cd.channel, msp, host, uint32(port))
		}
	}

	return err
}
