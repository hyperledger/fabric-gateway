/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"context"
	"io"

	pb "github.com/hyperledger/fabric-gateway/protos"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

func (gs *GatewayServer) SubmitTransaction(ctx context.Context, txn *pb.Transaction) (*pb.Result, error) {
	signer, err := CreateSigner(txn.Id.Msp, txn.Id.Cert, txn.Id.Key)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create Signer: ")
	}

	proposal, err := gs.createProposal(txn, signer)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create proposal: ")
	}

	signedProposal, err := gs.signProposal(proposal, signer)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to sign proposal: ")
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

	signedTransaction, err := createSignedTx(proposal, gs.gatewaySigner, responses...)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to assemble transaction: ")
	}

	err = gs.broadcastClient.Send(signedTransaction)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send envelope to orderer")
	}

	oresp, err := gs.broadcastClient.Recv()
	if err == io.EOF {
		return nil, errors.Wrap(err, "failed to to get response from orderer")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to to get response from orderer")
	}

	status := oresp.Info

	return &pb.Result{Value: []byte(status)}, nil
}

func (gs *GatewayServer) EvaluateTransaction(ctx context.Context, txn *pb.Transaction) (*pb.Result, error) {
	response, err := gs.endorseProposal(ctx, txn)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to endorse proposal: ")
	}

	return getValueFromResponse(response)
}

func (gs *GatewayServer) endorseProposal(ctx context.Context, txn *pb.Transaction) (*peer.ProposalResponse, error) {
	signer, err := CreateSigner(txn.Id.Msp, txn.Id.Cert, txn.Id.Key)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create Signer: ")
	}

	proposal, err := gs.createProposal(txn, signer)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create proposal: ")
	}

	signedProposal, err := gs.signProposal(proposal, signer)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to sign proposal: ")
	}

	response, err := gs.endorserClients[0].ProcessProposal(ctx, signedProposal)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to process proposal: ")
	}

	return response, nil
}
