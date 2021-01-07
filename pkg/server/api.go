/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/golang/protobuf/proto"

	pb "github.com/hyperledger/fabric-gateway/protos"

	"github.com/hyperledger/fabric-protos-go/peer"
)

// Evaluate will invoke the transaction function as specified in the SignedProposal
func (gs *Server) Evaluate(ctx context.Context, proposedTransaction *pb.ProposedTransaction) (*pb.Result, error) {
	signedProposal := proposedTransaction.Proposal
	channel, chaincodeID, err := getChannelAndChaincodeFromSignedProposal(proposedTransaction.Proposal)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack channel header: %w", err)
	}
	endorsers := gs.registry.GetEndorsers(channel, chaincodeID)
	fmt.Printf("Evaluate: #endorsers=%d\n", len(endorsers))
	if len(endorsers) == 0 {
		return nil, fmt.Errorf("No endorsing peers found for channel: %s", proposedTransaction.ChannelId)
	}
	response, err := endorsers[0].ProcessProposal(ctx, signedProposal) // choose suitable peer
	if err != nil {
		return nil, fmt.Errorf("Failed to evaluate transaction: %w", err)
	}

	return getValueFromResponse(response)
}

// Endorse will collect endorsements by invoking the transaction function specified in the SignedProposal against
// sufficient Peers to satisfy the endorsement policy.
func (gs *Server) Endorse(ctx context.Context, proposedTransaction *pb.ProposedTransaction) (*pb.PreparedTransaction, error) {
	signedProposal := proposedTransaction.Proposal
	var proposal peer.Proposal
	if err := proto.Unmarshal(signedProposal.ProposalBytes, &proposal); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal signed proposal: %w", err)
	}
	channel, chaincodeID, err := getChannelAndChaincodeFromSignedProposal(proposedTransaction.Proposal)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack channel header: %w", err)
	}
	endorsers := gs.registry.GetEndorsers(channel, chaincodeID)

	var responses []*peer.ProposalResponse
	// send to all the endorsers
	for i := 0; i < len(endorsers); i++ {
		response, err := endorsers[i].ProcessProposal(ctx, signedProposal)
		if err != nil {
			return nil, fmt.Errorf("Failed to process proposal: %w", err)
		}
		responses = append(responses, response)
	}

	env, err := createUnsignedTx(&proposal, responses...)
	if err != nil {
		return nil, fmt.Errorf("Failed to assemble transaction: %w", err)
	}

	retVal, err := getValueFromResponse(responses[0])
	if err != nil {
		return nil, fmt.Errorf("Failed to extract value from reponse payload: %w", err)
	}

	preparedTxn := &pb.PreparedTransaction{
		TxId:      proposedTransaction.TxId,
		ChannelId: proposedTransaction.ChannelId,
		Response:  retVal,
		Envelope:  env,
	}
	return preparedTxn, nil
}

// Submit will send the signed transaction to the ordering service.  The output stream will close
// once the transaction is committed on a sufficient number of peers according to a defined policy.
func (gs *Server) Submit(txn *pb.PreparedTransaction, cs pb.Gateway_SubmitServer) error {
	orderers := gs.registry.GetOrderers(txn.ChannelId)

	if len(orderers) == 0 {
		return fmt.Errorf("no orderers discovered")
	}

	done := make(chan bool)
	go gs.registry.ListenForTxEvents("mychannel", txn.TxId, done)

	if err := orderers[0].Send(txn.Envelope); err != nil {
		return fmt.Errorf("Failed to send envelope to orderer: %w", err)
	}

	oresp, err := orderers[0].Recv()
	if err == io.EOF {
		return fmt.Errorf("Failed to to get response from orderer: %w", err)
	}
	if err != nil {
		return fmt.Errorf("Failed to to get response from orderer: %w", err)
	}
	if oresp == nil {
		return fmt.Errorf("received nil response from orderer")
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
