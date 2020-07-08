/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"context"
	"io"

	"github.com/gogo/protobuf/proto"
	pb "github.com/hyperledger/fabric-gateway/protos"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
)

type gatewayServer struct {
	ccpPath         string
	endorserClients []peer.EndorserClient
	broadcastClient orderer.AtomicBroadcast_BroadcastClient
}

func (gs *gatewayServer) SubmitTransaction(ctx context.Context, txn *pb.Transaction) (*pb.Response, error) {
	signer, err := createSigner(txn.Id.Msp, txn.Id.Cert, txn.Id.Key)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create signer: ")
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

	signedTransaction, err := protoutil.CreateSignedTx(proposal, signer, responses...)
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

	return &pb.Response{Value: []byte(status)}, nil
}

func (gs *gatewayServer) EvaluateTransaction(ctx context.Context, txn *pb.Transaction) (*pb.Response, error) {
	response, err := gs.endorseProposal(ctx, txn)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to endorse proposal: ")
	}

	var retVal []byte

	if response.Payload != nil {
		payload, err := protoutil.UnmarshalProposalResponsePayload(response.Payload)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshal of proposal response payload failed")
		}

		extension, err := protoutil.UnmarshalChaincodeAction(payload.Extension)
		if err != nil {
			return nil, errors.Wrap(err, "unmarshal of chaincode action failed")
		}

		if extension != nil && extension.Response != nil {
			retVal = extension.Response.Payload
		}
	}

	return &pb.Response{Value: retVal}, nil
}

func (gs *gatewayServer) endorseProposal(ctx context.Context, txn *pb.Transaction) (*peer.ProposalResponse, error) {
	signer, err := createSigner(txn.Id.Msp, txn.Id.Cert, txn.Id.Key)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create signer: ")
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

func (gs *gatewayServer) createProposal(txn *pb.Transaction, signer *signer) (*peer.Proposal, error) {
	if txn.ChaincodeID == "" {
		return nil, errors.New("ChaincodeID is required")
	}

	if txn.TxnName == "" {
		return nil, errors.New("Fcn is required")
	}

	// Add function name to arguments
	argsArray := make([][]byte, len(txn.Args)+1)
	argsArray[0] = []byte(txn.TxnName)
	for i, arg := range txn.Args {
		argsArray[i+1] = []byte(arg)
	}

	// create invocation spec to target a chaincode with arguments
	ccis := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			Type:        peer.ChaincodeSpec_NODE,
			ChaincodeId: &peer.ChaincodeID{Name: txn.ChaincodeID},
			Input:       &peer.ChaincodeInput{Args: argsArray},
		},
	}

	creator, err := signer.Serialize()
	if err != nil {
		return nil, errors.Wrap(err, "Failed to serialize signer: ")
	}

	proposal, _, err := protoutil.CreateChaincodeProposal(
		common.HeaderType_ENDORSER_TRANSACTION,
		txn.Channel,
		ccis,
		creator,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create chaincode proposal")
	}

	return proposal, nil
}

func (gs *gatewayServer) signProposal(proposal *peer.Proposal, signer *signer) (*peer.SignedProposal, error) {
	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal chaincode proposal")
	}

	signature, err := signer.Sign(proposalBytes)
	if err != nil {
		return nil, err
	}

	sproposal := &peer.SignedProposal{
		ProposalBytes: proposalBytes,
		Signature:     signature,
	}
	return sproposal, nil
}

// NewGatewayServer creates a server side implementation of the gateway server grpc
func NewGatewayServer(ccpPath string, clients []peer.EndorserClient, broadcastClient orderer.AtomicBroadcast_BroadcastClient) (*gatewayServer, error) {
	return &gatewayServer{ccpPath, clients, broadcastClient}, nil
}
