/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"bytes"
	"context"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"

	"github.com/hyperledger/fabric-gateway/pkg/server/mocks"
	pb "github.com/hyperledger/fabric-protos-go/gateway"

	"github.com/hyperledger/fabric-protos-go/peer"
)

func TestEvaluate(t *testing.T) {
	proposal, err := createProposal()
	if err != nil {
		t.Fatalf("Failed to create the proposal: %s", err)
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

	ptx := &pb.ProposedTransaction{
		Proposal: sp,
	}

	result, err := server.Evaluate(context.TODO(), ptx)

	if err != nil {
		t.Fatalf("Failed to evaluate the proposal: %s", err)
	}

	if !bytes.Equal(result.Value, []byte("MyResult")) {
		t.Fatalf("Incorrect value: %s", result.Value)
	}
}

func TestEndorse(t *testing.T) {
	proposal, err := createProposal()
	if err != nil {
		t.Fatalf("Failed to create the proposal: %s", err)
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

	ptx := &pb.ProposedTransaction{
		Proposal: sp,
	}

	result, err := server.Endorse(context.TODO(), ptx)

	if err != nil {
		t.Fatalf("Failed to prepare the transaction: %s", err)
	}

	if !bytes.Equal(result.Response.Value, []byte("MyResult")) {
		t.Fatalf("Incorrect value: %s", result.Response.Value)
	}
}

func TestSubmit(t *testing.T) {
	proposal, err := createProposal()
	if err != nil {
		t.Fatalf("Failed to create the proposal: %s", err)
	}

	signer := func(digest []byte) ([]byte, error) {
		return []byte("mysignature"), nil
	}

	server := &Server{
		mocks.NewMockRegistry(),
	}

	sp, err := server.signProposal(proposal, signer)

	ptx := &pb.ProposedTransaction{
		Proposal: sp,
	}

	preparedTx, err := server.Endorse(context.TODO(), ptx)
	if err != nil {
		t.Fatalf("Failed to sign the proposal: %s", err)
	}

	if err != nil {
		t.Fatalf("Failed to prepare the transaction: %s", err)
	}

	// sign the envelope
	preparedTx.Envelope.Signature = []byte("mysignature")

	cs := mocks.NewMockSubmitServer()
	err = server.Submit(preparedTx, cs)

	if err != nil {
		t.Fatalf("Failed to commit the transaction: %s", err)
	}

	// if !bytes.Equal(result.Response.Value, []byte("MyResult")) {
	// 	t.Fatalf("Incorrect value: %s", result.Response.Value)
	// }
}

func createProposal() (*peer.Proposal, error) {
	invocationSpec := &peer.ChaincodeInvocationSpec{
		ChaincodeSpec: &peer.ChaincodeSpec{
			Type:        peer.ChaincodeSpec_NODE,
			ChaincodeId: &peer.ChaincodeID{Name: "my_chaincode"},
			Input:       &peer.ChaincodeInput{Args: nil},
		},
	}

	invocationSpecBytes, err := proto.Marshal(invocationSpec)
	if err != nil {
		return nil, err
	}

	payload := &peer.ChaincodeProposalPayload{
		Input: invocationSpecBytes,
	}

	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return nil, err
	}

	channelHeader := &common.ChannelHeader{
		ChannelId: "my_channel",
	}

	channelHeaderBytes, err := proto.Marshal(channelHeader)
	if err != nil {
		return nil, err
	}

	header := &common.Header{
		ChannelHeader: channelHeaderBytes,
	}

	headerBytes, err := proto.Marshal(header)
	if err != nil {
		return nil, err
	}

	proposal := &peer.Proposal{
		Header:  headerBytes,
		Payload: payloadBytes,
	}

	return proposal, nil
}
