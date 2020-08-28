/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/pkg/errors"

	pb "github.com/hyperledger/fabric-gateway/protos"

	"github.com/hyperledger/fabric-protos-go/peer"
)

func (gs *GatewayServer) Evaluate(ctx context.Context, signedProposal *peer.SignedProposal) (*pb.Result, error) {
	channelHeader, err := getChannelHeaderFromSignedProposal(signedProposal)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to unpack channel header: ")
	}
	endorsers := gs.registry.getEndorsers(channelHeader.ChannelId)
	if len(endorsers) == 0 {
		return nil, errors.New("No endorsing peers found for channel: " + channelHeader.ChannelId)
	}
	response, err := endorsers[0].ProcessProposal(ctx, signedProposal) // choose suitable peer
	if err != nil {
		return nil, errors.Wrap(err, "Failed to evaluate transaction: ")
	}

	return getValueFromResponse(response)
}

func (gs *GatewayServer) Prepare(ctx context.Context, signedProposal *peer.SignedProposal) (*pb.PreparedTransaction, error) {
	var proposal peer.Proposal
	err := proto.Unmarshal(signedProposal.ProposalBytes, &proposal)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal signed proposal")
	}
	channelHeader, err := getChannelHeaderFromSignedProposal(signedProposal)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to unpack channel header: ")
	}
	endorsers := gs.registry.getEndorsers(channelHeader.ChannelId)

	var responses []*peer.ProposalResponse
	// send to all the endorsers
	for i := 0; i < len(endorsers); i++ {
		response, err := endorsers[i].ProcessProposal(ctx, signedProposal)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to process proposal: ")
		}
		responses = append(responses, response)
	}

	env, err := createUnsignedTx(&proposal, responses...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to assemble transaction: ")
	}

	retVal, err := getValueFromResponse(responses[0])
	if err != nil {
		return nil, errors.Wrap(err, "failed to extract value from reponse payload")
	}

	preparedTxn := &pb.PreparedTransaction{
		TxId:     channelHeader.TxId,
		Response: retVal,
		Envelope: env,
	}
	return preparedTxn, nil
}

func (gs *GatewayServer) Commit(txn *pb.PreparedTransaction, cs pb.Gateway_CommitServer) error {
	channelHeader, err := getChannelHeaderFromEnvelope(txn.Envelope)
	if err != nil {
		return errors.Wrap(err, "Failed to unpack channel header: ")
	}
	deliverers := gs.registry.getDeliverers(channelHeader.ChannelId)
	orderers := gs.registry.getOrderers(channelHeader.ChannelId)

	if len(orderers) == 0 {
		return errors.New("no orderers discovered")
	}

	done := make(chan bool)
	go listenForTxEvents(deliverers, "mychannel", txn.TxId, gs.gatewaySigner, done)

	err = orderers[0].Send(txn.Envelope)
	if err != nil {
		return errors.Wrap(err, "failed to send envelope to orderer")
	}

	oresp, err := orderers[0].Recv()
	if err == io.EOF {
		return errors.Wrap(err, "failed to to get response from orderer")
	}
	if err != nil {
		return errors.Wrap(err, "failed to to get response from orderer")
	}

	status := oresp.Info

	cs.Send(&pb.Event{
		Value: []byte(status),
	})

	select {
	case <-done:
		fmt.Println("received enough commit events")
	case <-time.After(5 * time.Second):
		fmt.Println("timed out waiting for commit events")
	}

	return nil
}

func getChannelHeaderFromSignedProposal(signedProposal *peer.SignedProposal) (*common.ChannelHeader, error) {
	var proposal peer.Proposal
	err := proto.Unmarshal(signedProposal.ProposalBytes, &proposal)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal signed proposal")
	}
	var header common.Header
	err = proto.Unmarshal(proposal.Header, &header)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal header: ")
	}
	var channelHeader common.ChannelHeader
	err = proto.Unmarshal(header.ChannelHeader, &channelHeader)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal channel header: ")
	}
	return &channelHeader, nil
}

func getChannelHeaderFromEnvelope(envelope *common.Envelope) (*common.ChannelHeader, error) {
	var payload common.Payload
	err := proto.Unmarshal(envelope.Payload, &payload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal signed proposal")
	}
	var channelHeader common.ChannelHeader
	err = proto.Unmarshal(payload.Header.ChannelHeader, &channelHeader)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal channel header: ")
	}
	return &channelHeader, nil
}
