/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"bytes"
	"crypto/rand"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	pb "github.com/hyperledger/fabric-gateway/protos"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/pkg/errors"
)

func (gs *Server) signProposal(proposal *peer.Proposal, sign identity.Sign) (*peer.SignedProposal, error) {
	proposalBytes, err := proto.Marshal(proposal)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal chaincode proposal")
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
		return nil, errors.New("at least one proposal response is required")
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
			return nil, errors.Errorf("proposal response was not successful, error code %d, msg %s", r.Response.Status, r.Response.Message)
		}

		if n == 0 {
			a1 = r.Payload
			continue
		}

		if !bytes.Equal(a1, r.Payload) {
			return nil, errors.New("ProposalResponsePayloads do not match")
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
		return nil, errors.Wrap(err, "error marshaling ChaincodeActionPayload")
	}

	// create a transaction
	taa := &peer.TransactionAction{Header: hdr.SignatureHeader, Payload: capBytes}
	taas := make([]*peer.TransactionAction, 1)
	taas[0] = taa
	tx := &peer.Transaction{Actions: taas}

	// serialize the tx
	txBytes, err := proto.Marshal(tx)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling Transaction")
	}

	// create the payload
	payl := &common.Payload{Header: hdr, Data: txBytes}
	paylBytes, err := proto.Marshal(payl)
	if err != nil {
		return nil, errors.Wrap(err, "error marshaling Payload")
	}

	// here's the envelope
	return &common.Envelope{Payload: paylBytes}, nil
}

func getRandomNonce() ([]byte, error) {
	key := make([]byte, 24)

	_, err := rand.Read(key)
	if err != nil {
		return nil, errors.Wrap(err, "error getting random bytes")
	}
	return key, nil
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
	return unmarshalChannelHeader(header.ChannelHeader)
}

func getChannelHeaderFromEnvelope(envelope *common.Envelope) (*common.ChannelHeader, error) {
	var payload common.Payload
	err := proto.Unmarshal(envelope.Payload, &payload)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal signed proposal")
	}
	return unmarshalChannelHeader(payload.Header.ChannelHeader)
}

func unmarshalChannelHeader(bytes []byte) (*common.ChannelHeader, error) {
	var channelHeader common.ChannelHeader
	err := proto.Unmarshal(bytes, &channelHeader)
	return &channelHeader, errors.Wrap(err, "Failed to unmarshal channel header: ")
}

func unmarshalProposalResponsePayload(prpBytes []byte) (*peer.ProposalResponsePayload, error) {
	prp := &peer.ProposalResponsePayload{}
	err := proto.Unmarshal(prpBytes, prp)
	return prp, errors.Wrap(err, "error unmarshaling ProposalResponsePayload")
}

func unmarshalChaincodeAction(caBytes []byte) (*peer.ChaincodeAction, error) {
	chaincodeAction := &peer.ChaincodeAction{}
	err := proto.Unmarshal(caBytes, chaincodeAction)
	return chaincodeAction, errors.Wrap(err, "error unmarshaling ChaincodeAction")
}

func unmarshalHeader(bytes []byte) (*common.Header, error) {
	hdr := &common.Header{}
	err := proto.Unmarshal(bytes, hdr)
	return hdr, errors.Wrap(err, "error unmarshaling Header")
}

func unmarshalChaincodeProposalPayload(bytes []byte) (*peer.ChaincodeProposalPayload, error) {
	cpp := &peer.ChaincodeProposalPayload{}
	err := proto.Unmarshal(bytes, cpp)
	return cpp, errors.Wrap(err, "error unmarshaling ChaincodeProposalPayload")
}

func getBytesProposalPayloadForTx(
	payload *peer.ChaincodeProposalPayload,
) ([]byte, error) {
	// check for nil argument
	if payload == nil {
		return nil, errors.New("nil arguments")
	}

	// strip the transient bytes off the payload
	cppNoTransient := &peer.ChaincodeProposalPayload{Input: payload.Input, TransientMap: nil}
	cppBytes, err := proto.Marshal(cppNoTransient)
	return cppBytes, errors.Wrap(err, "error marshaling ChaincodeProposalPayload")
}
