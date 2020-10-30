/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mock

import (
	"context"
	"errors"

	proto "github.com/hyperledger/fabric-gateway/protos"
	"google.golang.org/grpc"
)

// GatewayClient mock implementation whose method implementations can be overridden by assigning to properties
type GatewayClient struct {
	MockEndorse  func(ctx context.Context, in *proto.ProposedTransaction, opts ...grpc.CallOption) (*proto.PreparedTransaction, error)
	MockSubmit   func(ctx context.Context, in *proto.PreparedTransaction, opts ...grpc.CallOption) (proto.Gateway_SubmitClient, error)
	MockEvaluate func(ctx context.Context, in *proto.ProposedTransaction, opts ...grpc.CallOption) (*proto.Result, error)
}

// NewGatewayClient creates a mock
func NewGatewayClient() *GatewayClient {
	return &GatewayClient{
		MockEndorse: func(ctx context.Context, in *proto.ProposedTransaction, opts ...grpc.CallOption) (*proto.PreparedTransaction, error) {
			return nil, errors.New("Not implemented")
		},
		MockSubmit: func(ctx context.Context, in *proto.PreparedTransaction, opts ...grpc.CallOption) (proto.Gateway_SubmitClient, error) {
			return nil, errors.New("Not implemented")
		},
		MockEvaluate: func(ctx context.Context, in *proto.ProposedTransaction, opts ...grpc.CallOption) (*proto.Result, error) {
			return nil, errors.New("Not implemented")
		},
	}
}

// Endorse mock implementation
func (mock *GatewayClient) Endorse(ctx context.Context, in *proto.ProposedTransaction, opts ...grpc.CallOption) (*proto.PreparedTransaction, error) {
	return mock.MockEndorse(ctx, in, opts...)
}

// Submit mock implementation
func (mock *GatewayClient) Submit(ctx context.Context, in *proto.PreparedTransaction, opts ...grpc.CallOption) (proto.Gateway_SubmitClient, error) {
	return mock.MockSubmit(ctx, in, opts...)
}

// Evaluate mock implementation
func (mock *GatewayClient) Evaluate(ctx context.Context, in *proto.ProposedTransaction, opts ...grpc.CallOption) (*proto.Result, error) {
	return mock.MockEvaluate(ctx, in, opts...)
}
