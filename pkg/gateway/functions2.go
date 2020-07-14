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
	response, err := gs.endorserClients[0].ProcessProposal(ctx, signedProposal) // choose suitable peer
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

	var responses []*peer.ProposalResponse
	// send to all the endorsers
	for i := 0; i < len(gs.endorserClients); i++ {
		response, err := gs.endorserClients[i].ProcessProposal(ctx, signedProposal)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to process proposal: ")
		}
		responses = append(responses, response)
	}

	env, err := createUnsignedTx(&proposal, responses...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to assemble transaction: ")
	}

	// get the txId
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
	done := make(chan bool)
	go listenForTxEvents(gs.deliverClients, "mychannel", txn.TxId, gs.gatewaySigner, done)

	err := gs.broadcastClient.Send(txn.Envelope)
	if err != nil {
		return errors.Wrap(err, "failed to send envelope to orderer")
	}

	oresp, err := gs.broadcastClient.Recv()
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
