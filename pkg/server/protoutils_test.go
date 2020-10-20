/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"bytes"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-gateway/pkg/server/mocks"

	"github.com/hyperledger/fabric-protos-go/peer"
)

func TestSignProposal(t *testing.T) {
	t.Run("Sign with a dummy signer", func(t *testing.T) {
		proposal := &peer.Proposal{
			Header:  []byte{},
			Payload: []byte{},
		}

		signer := func(digest []byte) ([]byte, error) {
			return []byte("mysignature"), nil
		}

		server := &Server{
			mocks.NewMockRegistry(),
		}

		sp, err := server.signProposal(proposal, signer)

		if err != nil {
			t.Fatalf("Failed to sign the proposal: %s", err)
		}

		if !bytes.Equal(sp.Signature, []byte("mysignature")) {
			t.Fatalf("Incorrect signature: %v", sp.Signature)
		}
	})
}

func marshal(msg proto.Message, t *testing.T) []byte {
	buf, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}
	return buf
}

func createProposalResponse(value string, t *testing.T) *peer.ProposalResponse {
	response := &peer.Response{
		Status:  200,
		Payload: []byte(value),
	}
	action := &peer.ChaincodeAction{
		Response: response,
	}
	payload := &peer.ProposalResponsePayload{
		ProposalHash: []byte{},
		Extension:    marshal(action, t),
	}
	return &peer.ProposalResponse{
		Payload: marshal(payload, t),
		Response: response,
	}
}

func TestGetValueFromResponse(t *testing.T) {
	response := createProposalResponse("MyResult", t)

	result, err := getValueFromResponse(response)
	if err != nil {
		t.Fatalf("Failed to extract value from reponse: %v", err)
	}
	if !bytes.Equal(result.Value, []byte("MyResult")) {
		t.Fatalf("Incorrect value: %s", result.Value)
	}
}

func TestCreateUnsignedTx(t *testing.T) {
	proposal := &peer.Proposal{
		Header:  []byte{},
		Payload: []byte{},
	}
	response := createProposalResponse("MyResult", t)
	_, err := createUnsignedTx(proposal, response)
	if err != nil {
		t.Fatalf("Failed to create unsigned tx: %s", err)
	}
}

func TestCreateUnsignedTxWithNoResponses(t *testing.T) {
	proposal := &peer.Proposal{
		Header:  []byte{},
		Payload: []byte{},
	}
	_, err := createUnsignedTx(proposal)
	if err == nil {
		t.Fatalf("Should have failed to create unsigned tx: %s", err)
	}
}

func TestCreateUnsignedTxWithMatchingResponses(t *testing.T) {
	proposal := &peer.Proposal{
		Header:  []byte{},
		Payload: []byte{},
	}
	response1 := createProposalResponse("MyResult", t)
	response2 := createProposalResponse("MyResult", t)
	_, err := createUnsignedTx(proposal, response1, response2)
	if err != nil {
		t.Fatalf("Failed to create unsigned tx: %s", err)
	}
}

func TestCreateUnsignedTxWithUnmatchingResponses(t *testing.T) {
	proposal := &peer.Proposal{
		Header:  []byte{},
		Payload: []byte{},
	}
	response1 := createProposalResponse("MyResult", t)
	response2 := createProposalResponse("DifferentResult", t)
	_, err := createUnsignedTx(proposal, response1, response2)
	if err == nil {
		t.Fatalf("Should have failed to create unsigned tx: %s", err)
	}
}
