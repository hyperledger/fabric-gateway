/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"

	"github.com/hyperledger/fabric-protos-go/gateway"
	"google.golang.org/grpc"
)

type gatewayClient struct {
	grpcClient gateway.GatewayClient
	contexts   *contextFactory
}

func (client *gatewayClient) Endorse(in *gateway.EndorseRequest, opts ...grpc.CallOption) (*gateway.EndorseResponse, error) {
	ctx, cancel := client.contexts.Endorse()
	defer cancel()
	return client.EndorseWithContext(ctx, in, opts...)
}

func (client *gatewayClient) EndorseWithContext(ctx context.Context, in *gateway.EndorseRequest, opts ...grpc.CallOption) (*gateway.EndorseResponse, error) {
	return client.grpcClient.Endorse(ctx, in, opts...)
}

func (client *gatewayClient) Submit(in *gateway.SubmitRequest, opts ...grpc.CallOption) (*gateway.SubmitResponse, error) {
	ctx, cancel := client.contexts.Submit()
	defer cancel()
	return client.SubmitWithContext(ctx, in, opts...)
}

func (client *gatewayClient) SubmitWithContext(ctx context.Context, in *gateway.SubmitRequest, opts ...grpc.CallOption) (*gateway.SubmitResponse, error) {
	return client.grpcClient.Submit(ctx, in, opts...)
}

func (client *gatewayClient) CommitStatus(in *gateway.SignedCommitStatusRequest, opts ...grpc.CallOption) (*gateway.CommitStatusResponse, error) {
	ctx, cancel := client.contexts.CommitStatus()
	defer cancel()
	return client.CommitStatusWithContext(ctx, in, opts...)
}

func (client *gatewayClient) CommitStatusWithContext(ctx context.Context, in *gateway.SignedCommitStatusRequest, opts ...grpc.CallOption) (*gateway.CommitStatusResponse, error) {
	return client.grpcClient.CommitStatus(ctx, in, opts...)
}

func (client *gatewayClient) Evaluate(in *gateway.EvaluateRequest, opts ...grpc.CallOption) (*gateway.EvaluateResponse, error) {
	ctx, cancel := client.contexts.Evaluate()
	defer cancel()
	return client.EvaluateWithContext(ctx, in, opts...)
}

func (client *gatewayClient) EvaluateWithContext(ctx context.Context, in *gateway.EvaluateRequest, opts ...grpc.CallOption) (*gateway.EvaluateResponse, error) {
	return client.grpcClient.Evaluate(ctx, in, opts...)
}

func (client *gatewayClient) ChaincodeEvents(ctx context.Context, in *gateway.SignedChaincodeEventsRequest, opts ...grpc.CallOption) (gateway.Gateway_ChaincodeEventsClient, error) {
	return client.grpcClient.ChaincodeEvents(ctx, in, opts...)
}
