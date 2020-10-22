/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package server

import (
	"bytes"
	"context"
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/server/mocks"
	pb "github.com/hyperledger/fabric-gateway/protos"

	"github.com/hyperledger/fabric-protos-go/peer"
)

func TestEvaluate(t *testing.T) {
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

	ptx := &pb.ProposedTransaction{
		Proposal: sp,
	}

	preparedTx, err := server.Endorse(context.TODO(), ptx)

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
