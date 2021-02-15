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

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/gateway"
	"google.golang.org/grpc"
)

func TestSign(t *testing.T) {
	preparedTransaction := gateway.PreparedTransaction{
		Envelope: &common.Envelope{},
		Response: &gateway.Result{},
	}

	newMockSubmitClient := func(controller *gomock.Controller) *MockGateway_SubmitClient {
		mock := NewMockGateway_SubmitClient(controller)
		mock.EXPECT().Recv().
			Return(nil, io.EOF).
			AnyTimes()
		return mock
	}

	t.Run("Evaluate signs proposal using client signing implementation", func(t *testing.T) {
		expected := []byte("SIGNATURE")
		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		var actual []byte
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.ProposedTransaction, _ ...grpc.CallOption) {
				actual = in.Proposal.Signature
			}).
			Return(&gateway.Result{}, nil).
			Times(1)

		contract := AssertNewTestContract(t, "contract", WithClient(mockClient), WithSign(sign))

		if _, err := contract.EvaluateTransaction("transaction"); err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(actual, expected) {
			t.Fatalf("Expected signature: %v\nGot: %v", expected, actual)
		}
	})

	t.Run("Submit signs proposal using client signing implementation", func(t *testing.T) {
		expected := []byte("SIGNATURE")
		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		var actual []byte
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.ProposedTransaction, _ ...grpc.CallOption) {
				actual = in.Proposal.Signature
			}).
			Return(&preparedTransaction, nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(newMockSubmitClient(mockController), nil)

		contract := AssertNewTestContract(t, "contract", WithClient(mockClient), WithSign(sign))

		if _, err := contract.SubmitTransaction("transaction"); err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(actual, expected) {
			t.Fatalf("Expected signature: %v\nGot: %v", expected, actual)
		}
	})

	t.Run("Submit signs transaction using client signing implementation", func(t *testing.T) {
		expected := []byte("SIGNATURE")
		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		var actual []byte
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(&preparedTransaction, nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.PreparedTransaction, _ ...grpc.CallOption) {
				actual = in.Envelope.Signature
			}).
			Return(newMockSubmitClient(mockController), nil).
			Times(1)

		contract := AssertNewTestContract(t, "contract", WithClient(mockClient), WithSign(sign))

		if _, err := contract.SubmitTransaction("transaction"); err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(actual, expected) {
			t.Fatalf("Expected signature: %v\nGot: %v", expected, actual)
		}
	})

	t.Run("Default error implementation is used if no signing implementation supplied", func(t *testing.T) {
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
			Return(&gateway.Result{}, nil).
			AnyTimes()

		gateway, err := Connect(TestCredentials.identity, WithClient(mockClient))
		if err != nil {
			t.Fatal(err)
		}

		contract := gateway.GetNetwork("network").GetContract("chaincode")

		if _, err := contract.EvaluateTransaction("transaction"); nil == err {
			t.Fatal("Expected signing error but got nil")
		}
	})
}
