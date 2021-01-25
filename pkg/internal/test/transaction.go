/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package test

import (
	"testing"

	"github.com/gogo/protobuf/proto"
	gateway "github.com/hyperledger/fabric-gateway/protos/gateway"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func AssertUnmarshall(t *testing.T, b []byte, m proto.Message) {
	if err := proto.Unmarshal(b, m); err != nil {
		t.Fatal(err)
	}
}

func AssertUnmarshallInvocationSpec(t *testing.T, proposedTransaction *gateway.ProposedTransaction) *peer.ChaincodeInvocationSpec {
	proposal := &peer.Proposal{}
	AssertUnmarshall(t, proposedTransaction.Proposal.ProposalBytes, proposal)

	payload := &peer.ChaincodeProposalPayload{}
	AssertUnmarshall(t, proposal.Payload, payload)

	input := &peer.ChaincodeInvocationSpec{}
	AssertUnmarshall(t, payload.Input, input)

	return input
}

func AssertUnmarshallChannelheader(t *testing.T, proposedTransaction *gateway.ProposedTransaction) *common.ChannelHeader {
	header := AssertUnmarshallHeader(t, proposedTransaction)

	channelHeader := &common.ChannelHeader{}
	AssertUnmarshall(t, header.ChannelHeader, channelHeader)

	return channelHeader
}

func AssertUnmarshallHeader(t *testing.T, proposedTransaction *gateway.ProposedTransaction) *common.Header {
	proposal := &peer.Proposal{}
	AssertUnmarshall(t, proposedTransaction.Proposal.ProposalBytes, proposal)

	header := &common.Header{}
	AssertUnmarshall(t, proposal.Header, header)

	return header
}

func AssertUnmarshallSignatureHeader(t *testing.T, proposedTransaction *gateway.ProposedTransaction) *common.SignatureHeader {
	header := AssertUnmarshallHeader(t, proposedTransaction)

	signatureHeader := &common.SignatureHeader{}
	AssertUnmarshall(t, header.SignatureHeader, signatureHeader)

	return signatureHeader
}
