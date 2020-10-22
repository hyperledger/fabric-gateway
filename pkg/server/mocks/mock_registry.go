/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mocks

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	pb "github.com/hyperledger/fabric-gateway/protos"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/orderer"
	"github.com/hyperledger/fabric-protos-go/peer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type mockRegistry struct {
	endorsers []peer.EndorserClient
	orderers  []orderer.AtomicBroadcast_BroadcastClient
}

func NewMockRegistry() *mockRegistry {
	endorser1 := &mockEndorserClient{
		response: CreateProposalResponse("MyResult", nil),
	}
	endorsers := make([]peer.EndorserClient, 0)
	endorsers = append(endorsers, endorser1)

	orderer1 := &mockBroadcastClient{"MyStatus"}
	orderers := make([]orderer.AtomicBroadcast_BroadcastClient, 0)
	orderers = append(orderers, orderer1)

	return &mockRegistry{
		endorsers,
		orderers,
	}
}

func (mr *mockRegistry) GetEndorsers(channel string) []peer.EndorserClient {
	return mr.endorsers
}

func (mr *mockRegistry) GetDeliverers(channel string) []peer.DeliverClient {
	return nil
}

func (mr *mockRegistry) GetOrderers(channel string) []orderer.AtomicBroadcast_BroadcastClient {
	return mr.orderers
}

func (mr *mockRegistry) ListenForTxEvents(channel string, txid string, done chan<- bool) error {
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

type mockSubmitServer struct {
}

func NewMockSubmitServer() *mockSubmitServer {
	return &mockSubmitServer{}
}

func (mcs *mockSubmitServer) Send(*pb.Event) error {
	return nil
}

func (mcs *mockSubmitServer) SetHeader(metadata.MD) error {
	return nil
}

func (mcs *mockSubmitServer) SendHeader(metadata.MD) error {
	return nil
}

func (mcs *mockSubmitServer) SetTrailer(metadata.MD) {
}

func (mcs *mockSubmitServer) Context() context.Context {
	return nil
}

func (mcs *mockSubmitServer) SendMsg(m interface{}) error {
	return nil
}

func (mcs *mockSubmitServer) RecvMsg(m interface{}) error {
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
