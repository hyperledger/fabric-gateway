/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package test

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/require"
)

// AssertUnmarshall ensures that a protobuf is umarshalled without error
func AssertUnmarshall(t *testing.T, b []byte, m proto.Message) {
	err := proto.Unmarshal(b, m)
	require.NoError(t, err)
}

// AssertUnmarshallProposalPayload ensures that a ChaincodeProposalPayload protobuf is umarshalled without error
func AssertUnmarshallProposalPayload(t *testing.T, proposedTransaction *peer.SignedProposal) *peer.ChaincodeProposalPayload {
	proposal := &peer.Proposal{}
	AssertUnmarshall(t, proposedTransaction.ProposalBytes, proposal)

	payload := &peer.ChaincodeProposalPayload{}
	AssertUnmarshall(t, proposal.Payload, payload)

	return payload
}

// AssertUnmarshallInvocationSpec ensures that a ChaincodeInvocationSpec protobuf is umarshalled without error
func AssertUnmarshallInvocationSpec(t *testing.T, proposedTransaction *peer.SignedProposal) *peer.ChaincodeInvocationSpec {
	proposal := &peer.Proposal{}
	AssertUnmarshall(t, proposedTransaction.ProposalBytes, proposal)

	payload := &peer.ChaincodeProposalPayload{}
	AssertUnmarshall(t, proposal.Payload, payload)

	input := &peer.ChaincodeInvocationSpec{}
	AssertUnmarshall(t, payload.Input, input)

	return input
}

// AssertUnmarshallChannelheader ensures that a ChannelHeader protobuf is umarshalled without error
func AssertUnmarshallChannelheader(t *testing.T, proposedTransaction *peer.SignedProposal) *common.ChannelHeader {
	header := AssertUnmarshallHeader(t, proposedTransaction)

	channelHeader := &common.ChannelHeader{}
	AssertUnmarshall(t, header.ChannelHeader, channelHeader)

	return channelHeader
}

// AssertUnmarshallHeader ensures that a Header protobuf is umarshalled without error
func AssertUnmarshallHeader(t *testing.T, proposedTransaction *peer.SignedProposal) *common.Header {
	proposal := &peer.Proposal{}
	AssertUnmarshall(t, proposedTransaction.ProposalBytes, proposal)

	header := &common.Header{}
	AssertUnmarshall(t, proposal.Header, header)

	return header
}

// AssertUnmarshallSignatureHeader ensures that a SignatureHeader protobuf is umarshalled without error
func AssertUnmarshallSignatureHeader(t *testing.T, proposedTransaction *peer.SignedProposal) *common.SignatureHeader {
	header := AssertUnmarshallHeader(t, proposedTransaction)

	signatureHeader := &common.SignatureHeader{}
	AssertUnmarshall(t, header.SignatureHeader, signatureHeader)

	return signatureHeader
}
