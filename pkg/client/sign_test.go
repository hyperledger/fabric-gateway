/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/internal/test/mock"
	gateway "github.com/hyperledger/fabric-gateway/protos"
	"github.com/hyperledger/fabric-protos-go/common"
	"google.golang.org/grpc"
)

func TestSign(t *testing.T) {
	t.Run("Evaluate signs proposal using client signing implementation", func(t *testing.T) {
		expected := []byte("SIGNATURE")
		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		var actual []byte
		mockClient := mock.NewGatewayClient()
		mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
			actual = in.Proposal.Signature

			value := &gateway.Result{}
			return value, nil
		}
		contract := AssertNewDefaultTestContract(t, "contract", WithClient(mockClient), WithSign(sign))

		if _, err := contract.EvaluateTransaction("transaction"); err != nil {
			t.Fatal(err)
		}

		if bytes.Compare(actual, expected) != 0 {
			t.Fatalf("Expected signature: %v\nGot: %v", expected, actual)
		}
	})

	t.Run("Submit signs proposal using client signing implementation", func(t *testing.T) {
		expected := []byte("SIGNATURE")
		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		var actual []byte
		mockClient := mock.NewGatewayClient()
		mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
			actual = in.Proposal.Signature

			preparedTransaction := &gateway.PreparedTransaction{
				Envelope: &common.Envelope{},
				Response: &gateway.Result{},
			}
			return preparedTransaction, nil
		}
		mockClient.MockSubmit = func(ctx context.Context, in *gateway.PreparedTransaction, opts ...grpc.CallOption) (gateway.Gateway_SubmitClient, error) {
			submitClient := mock.NewSubmitClient()
			submitClient.MockRecv = func() (*gateway.Event, error) {
				return nil, io.EOF
			}
			return submitClient, nil
		}
		contract := AssertNewDefaultTestContract(t, "contract", WithClient(mockClient), WithSign(sign))

		if _, err := contract.SubmitTransaction("transaction"); err != nil {
			t.Fatal(err)
		}

		if bytes.Compare(actual, expected) != 0 {
			t.Fatalf("Expected signature: %v\nGot: %v", expected, actual)
		}
	})

	t.Run("Submit signs transaction using client signing implementation", func(t *testing.T) {
		expected := []byte("SIGNATURE")
		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		var actual []byte
		mockClient := mock.NewGatewayClient()
		mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
			preparedTransaction := &gateway.PreparedTransaction{
				Envelope: &common.Envelope{},
				Response: &gateway.Result{},
			}
			return preparedTransaction, nil
		}
		mockClient.MockSubmit = func(ctx context.Context, in *gateway.PreparedTransaction, opts ...grpc.CallOption) (gateway.Gateway_SubmitClient, error) {
			actual = in.Envelope.Signature

			submitClient := mock.NewSubmitClient()
			submitClient.MockRecv = func() (*gateway.Event, error) {
				return nil, io.EOF
			}
			return submitClient, nil
		}
		contract := AssertNewDefaultTestContract(t, "contract", WithClient(mockClient), WithSign(sign))

		if _, err := contract.SubmitTransaction("transaction"); err != nil {
			t.Fatal(err)
		}

		if bytes.Compare(actual, expected) != 0 {
			t.Fatalf("Expected signature: %v\nGot: %v", expected, actual)
		}
	})

	t.Run("Default error implementation is used if no signing implementation supplied", func(t *testing.T) {
		mockClient := mock.NewGatewayClient()
		mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
			return &gateway.Result{}, nil
		}

		id, _ := GetTestCredentials()
		gateway, err := Connect(id, WithClient(mockClient))
		if err != nil {
			t.Fatal(err)
		}

		contract := gateway.GetNetwork("network").GetDefaultContract("chaincode")

		if _, err := contract.EvaluateTransaction("transaction"); nil == err {
			t.Fatal("Expected signing error but got nil")
		}
	})
}
