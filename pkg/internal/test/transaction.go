/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package test

import (
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/internal/util"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/runtime/protoiface"
)

// AssertProtoEqual ensures an expected protobuf message matches an actual message
func AssertProtoEqual(t *testing.T, expected protoiface.MessageV1, actual protoiface.MessageV1) {
	require.True(t, util.ProtoEqual(expected, actual), "Expected %v, got %v", expected, actual)
}

// AssertUnmarshal ensures that a protobuf is umarshaled without error
func AssertUnmarshal(t *testing.T, b []byte, m protoiface.MessageV1) {
	err := util.Unmarshal(b, m)
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
