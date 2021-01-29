/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-protos-go/common"
	pb "github.com/hyperledger/fabric-protos-go/gateway"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type MockRegistry struct {
	endorsers []peer.EndorserClient
	orderers  []orderer.AtomicBroadcast_BroadcastClient
}

// NewMockRegistry creates a mock registry implementation
func NewMockRegistry() *MockRegistry {
	endorser1 := &mockEndorserClient{
		response: CreateProposalResponse("MyResult", nil),
	}
	endorsers := make([]peer.EndorserClient, 0)
	endorsers = append(endorsers, endorser1)

	orderer1 := &mockBroadcastClient{"MyStatus"}
	orderers := make([]orderer.AtomicBroadcast_BroadcastClient, 0)
	orderers = append(orderers, orderer1)

	return &MockRegistry{
		endorsers,
		orderers,
	}
}

// GetEndorsers mock implementation
func (mr *MockRegistry) GetEndorsers(channel string, chaincode string) []peer.EndorserClient {
	return mr.endorsers
}

// GetDeliverers mock implementation
func (mr *MockRegistry) GetDeliverers(channel string) []peer.DeliverClient {
	return nil
}

// GetOrderers mock implementation
func (mr *MockRegistry) GetOrderers(channel string) []orderer.AtomicBroadcast_BroadcastClient {
	return mr.orderers
}

// ListenForTxEvents mock implementation
func (mr *MockRegistry) ListenForTxEvents(channel string, txid string, done chan<- bool) error {
	return nil
}

type mockEndorserClient struct {
	response *peer.ProposalResponse
}

func (mec *mockEndorserClient) ProcessProposal(ctx context.Context, in *peer.SignedProposal, opts ...grpc.CallOption) (*peer.ProposalResponse, error) {
	return mec.response, nil
}

func marshal(msg proto.Message, t *testing.T) []byte {
	buf, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %s", err)
	}
	return buf
}

// CreateProposalResponse returns a fake proposal response for a given response value
func CreateProposalResponse(value string, t *testing.T) *peer.ProposalResponse {
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
		Payload:  marshal(payload, t),
		Response: response,
	}
}

type MockSubmitServer struct {
}

// NewMockSubmitServer creates a mock server for testing
func NewMockSubmitServer() *MockSubmitServer {
	return &MockSubmitServer{}
}

// Send mock implementation
func (mcs *MockSubmitServer) Send(*pb.Event) error {
	return nil
}

// SetHeader mock implementation
func (mcs *MockSubmitServer) SetHeader(metadata.MD) error {
	return nil
}

// SendHeader mock implementation
func (mcs *MockSubmitServer) SendHeader(metadata.MD) error {
	return nil
}

// SetTrailer mock implementation
func (mcs *MockSubmitServer) SetTrailer(metadata.MD) {
}

// Context mock implementation
func (mcs *MockSubmitServer) Context() context.Context {
	return nil
}

// SendMsg mock implementation
func (mcs *MockSubmitServer) SendMsg(m interface{}) error {
	return nil
}

// RecvMsg mock implementation
func (mcs *MockSubmitServer) RecvMsg(m interface{}) error {
	return nil
}

type mockBroadcastClient struct {
	info string
}

func (mbc *mockBroadcastClient) Send(*common.Envelope) error {
	return nil
}

func (mbc *mockBroadcastClient) Recv() (*orderer.BroadcastResponse, error) {
	return &orderer.BroadcastResponse{
		Info: mbc.info,
	}, nil
}

func (mbc *mockBroadcastClient) Header() (metadata.MD, error) {
	return nil, nil
}

func (mbc *mockBroadcastClient) Trailer() metadata.MD {
	return nil
}

func (mbc *mockBroadcastClient) CloseSend() error {
	return nil
}

func (mbc *mockBroadcastClient) Context() context.Context {
	return nil
}

func (mbc *mockBroadcastClient) SendMsg(m interface{}) error {
	return nil
}

func (mbc *mockBroadcastClient) RecvMsg(m interface{}) error {
	return nil
}
