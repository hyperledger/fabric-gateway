/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"bytes"
	"fmt"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	pb "github.com/hyperledger/fabric-gateway/protos"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func (gs *Server) signProposal(proposal *peer.Proposal, sign identity.Sign) (*peer.SignedProposal, error) {
	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chaincode proposal: %w", err)
	}

	signature, err := sign(proposalBytes)
	if err != nil {
		return nil, err
	}

	sproposal := &peer.SignedProposal{
		ProposalBytes: proposalBytes,
		Signature:     signature,
	}
	return sproposal, nil
}

func getValueFromResponse(response *peer.ProposalResponse) (*pb.Result, error) {
	var retVal []byte

	if response.Payload != nil {
		payload, err := unmarshalProposalResponsePayload(response.Payload)
		if err != nil {
			return nil, err
		}

		extension, err := unmarshalChaincodeAction(payload.Extension)
		if err != nil {
			return nil, err
		}

		if extension != nil && extension.Response != nil {
			if extension.Response.Status > 200 {
				return nil, fmt.Errorf("error %d, %s", extension.Response.Status, extension.Response.Message)
			}
			retVal = extension.Response.Payload
		}
	}

	return &pb.Result{Value: retVal}, nil
}

func createUnsignedTx(
	proposal *peer.Proposal,
	resps ...*peer.ProposalResponse,
) (*common.Envelope, error) {
	if len(resps) == 0 {
		return nil, fmt.Errorf("at least one proposal response is required")
	}

	// the original header
	hdr, err := unmarshalHeader(proposal.Header)
	if err != nil {
		return nil, err
	}

	// the original payload
	pPayl, err := unmarshalChaincodeProposalPayload(proposal.Payload)
	if err != nil {
		return nil, err
	}

	// ensure that all actions are bitwise equal and that they are successful
	var a1 []byte
	for n, r := range resps {
		if r.Response.Status < 200 || r.Response.Status >= 400 {
			return nil, fmt.Errorf("proposal response was not successful, error code %d, msg %s", r.Response.Status, r.Response.Message)
		}

		if n == 0 {
			a1 = r.Payload
			continue
		}

		if !bytes.Equal(a1, r.Payload) {
			return nil, fmt.Errorf("ProposalResponsePayloads do not match")
		}
	}

	// fill endorsements
	endorsements := make([]*peer.Endorsement, len(resps))
	for n, r := range resps {
		endorsements[n] = r.Endorsement
	}

	// create ChaincodeEndorsedAction
	cea := &peer.ChaincodeEndorsedAction{ProposalResponsePayload: resps[0].Payload, Endorsements: endorsements}

	// obtain the bytes of the proposal payload that will go to the transaction
	propPayloadBytes, err := getBytesProposalPayloadForTx(pPayl)
	if err != nil {
		return nil, err
	}

	// serialize the chaincode action payload
	cap := &peer.ChaincodeActionPayload{ChaincodeProposalPayload: propPayloadBytes, Action: cea}
	capBytes, err := proto.Marshal(cap)
	if err != nil {
		return nil, fmt.Errorf("error marshaling ChaincodeActionPayload: %w", err)
	}

	// create a transaction
	taa := &peer.TransactionAction{Header: hdr.SignatureHeader, Payload: capBytes}
	taas := make([]*peer.TransactionAction, 1)
	taas[0] = taa
	tx := &peer.Transaction{Actions: taas}

	// serialize the tx
	txBytes, err := proto.Marshal(tx)
	if err != nil {
		return nil, fmt.Errorf("error marshaling Transaction: %w", err)
	}

	// create the payload
	payl := &common.Payload{Header: hdr, Data: txBytes}
	paylBytes, err := proto.Marshal(payl)
	if err != nil {
		return nil, fmt.Errorf("error marshaling Payload: %w", err)
	}

	// here's the envelope
	return &common.Envelope{Payload: paylBytes}, nil
}

func getChannelAndChaincodeFromSignedProposal(signedProposal *peer.SignedProposal) (string, string, error) {
	var proposal peer.Proposal
	err := proto.Unmarshal(signedProposal.ProposalBytes, &proposal)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal signed proposal: %w", err)
	}
	var header common.Header
	err = proto.Unmarshal(proposal.Header, &header)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal header: %w", err)
	}
	channelHeader, err := unmarshalChannelHeader(header.ChannelHeader)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal channel header: %w", err)
	}
	var payload peer.ChaincodeProposalPayload
	err = proto.Unmarshal(proposal.Payload, &payload)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal chaincode payload: %w", err)
	}
	var spec peer.ChaincodeInvocationSpec
	err = proto.Unmarshal(payload.Input, &spec)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal chaincode invocation spec: %w", err)
	}
	return channelHeader.ChannelId, spec.ChaincodeSpec.ChaincodeId.Name, nil
}

func unmarshalChannelHeader(bytes []byte) (*common.ChannelHeader, error) {
	var channelHeader common.ChannelHeader
	if err := proto.Unmarshal(bytes, &channelHeader); err != nil {
		return nil, fmt.Errorf("failed to unmarshal channel header: %w", err)
	}
	return &channelHeader, nil
}

func unmarshalProposalResponsePayload(prpBytes []byte) (*peer.ProposalResponsePayload, error) {
	prp := &peer.ProposalResponsePayload{}
	if err := proto.Unmarshal(prpBytes, prp); err != nil {
		return nil, fmt.Errorf("error unmarshaling ProposalResponsePayload: %w", err)
	}
	return prp, nil
}

func unmarshalChaincodeAction(caBytes []byte) (*peer.ChaincodeAction, error) {
	chaincodeAction := &peer.ChaincodeAction{}
	if err := proto.Unmarshal(caBytes, chaincodeAction); err != nil {
		return nil, fmt.Errorf("error unmarshaling ChaincodeAction: %w", err)
	}
	return chaincodeAction, nil
}

func unmarshalHeader(bytes []byte) (*common.Header, error) {
	hdr := &common.Header{}
	if err := proto.Unmarshal(bytes, hdr); err != nil {
		return nil, fmt.Errorf("error unmarshaling Header: %w", err)
	}
	return hdr, nil
}

func unmarshalChaincodeProposalPayload(bytes []byte) (*peer.ChaincodeProposalPayload, error) {
	cpp := &peer.ChaincodeProposalPayload{}
	if err := proto.Unmarshal(bytes, cpp); err != nil {
		return nil, fmt.Errorf("error unmarshaling ChaincodeProposalPayload: %w", err)
	}
	return cpp, nil
}

func getBytesProposalPayloadForTx(
	payload *peer.ChaincodeProposalPayload,
) ([]byte, error) {
	// check for nil argument
	if payload == nil {
		return nil, fmt.Errorf("nil arguments")
	}

	// strip the transient bytes off the payload
	cppNoTransient := &peer.ChaincodeProposalPayload{Input: payload.Input, TransientMap: nil}
	cppBytes, err := proto.Marshal(cppNoTransient)
	if err != nil {
		return nil, fmt.Errorf("error marshaling ChaincodeProposalPayload: %w", err)
	}
	return cppBytes, nil
}
