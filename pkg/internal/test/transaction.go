/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package test

import (
	"testing"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// AssertProtoEqual ensures an expected protobuf message matches an actual message
func AssertProtoEqual(t *testing.T, expected protoreflect.ProtoMessage, actual protoreflect.ProtoMessage) {
	require.True(t, proto.Equal(expected, actual), "Expected %v, got %v", expected, actual)
}

// AssertUnmarshal ensures that a protobuf is umarshaled without error
func AssertUnmarshal(t *testing.T, b []byte, m protoreflect.ProtoMessage) {
	err := proto.Unmarshal(b, m)
	require.NoError(t, err)
}

// AssertUnmarshalProposalPayload ensures that a ChaincodeProposalPayload protobuf is umarshalled without error
func AssertUnmarshalProposalPayload(t *testing.T, proposedTransaction *peer.SignedProposal) *peer.ChaincodeProposalPayload {
	proposal := &peer.Proposal{}
	AssertUnmarshal(t, proposedTransaction.ProposalBytes, proposal)

	payload := &peer.ChaincodeProposalPayload{}
	AssertUnmarshal(t, proposal.Payload, payload)

	return payload
}

// AssertUnmarshalInvocationSpec ensures that a ChaincodeInvocationSpec protobuf is umarshalled without error
func AssertUnmarshalInvocationSpec(t *testing.T, proposedTransaction *peer.SignedProposal) *peer.ChaincodeInvocationSpec {
	proposal := &peer.Proposal{}
	AssertUnmarshal(t, proposedTransaction.ProposalBytes, proposal)

	payload := &peer.ChaincodeProposalPayload{}
	AssertUnmarshal(t, proposal.Payload, payload)

	input := &peer.ChaincodeInvocationSpec{}
	AssertUnmarshal(t, payload.Input, input)

	return input
}

// AssertUnmarshalChannelheader ensures that a ChannelHeader protobuf is umarshalled without error
func AssertUnmarshalChannelheader(t *testing.T, proposedTransaction *peer.SignedProposal) *common.ChannelHeader {
	header := AssertUnmarshalHeader(t, proposedTransaction)

	channelHeader := &common.ChannelHeader{}
	AssertUnmarshal(t, header.ChannelHeader, channelHeader)

	return channelHeader
}

// AssertUnmarshalHeader ensures that a Header protobuf is umarshalled without error
func AssertUnmarshalHeader(t *testing.T, proposedTransaction *peer.SignedProposal) *common.Header {
	proposal := &peer.Proposal{}
	AssertUnmarshal(t, proposedTransaction.ProposalBytes, proposal)

	header := &common.Header{}
	AssertUnmarshal(t, proposal.Header, header)

	return header
}

// AssertUnmarshalSignatureHeader ensures that a SignatureHeader protobuf is umarshalled without error
func AssertUnmarshalSignatureHeader(t *testing.T, proposedTransaction *peer.SignedProposal) *common.SignatureHeader {
	header := AssertUnmarshalHeader(t, proposedTransaction)

	signatureHeader := &common.SignatureHeader{}
	AssertUnmarshal(t, header.SignatureHeader, signatureHeader)

	return signatureHeader
}
