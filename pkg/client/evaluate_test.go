/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
	"github.com/hyperledger/fabric-protos-go/gateway"
	"github.com/hyperledger/fabric-protos-go/peer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewStatusError(t *testing.T, code codes.Code, message string, details ...proto.Message) error {
	s, err := status.New(code, message).WithDetails(details...)
	if err != nil {
		t.Fatal(err)
	}

	return s.Err()
}

func TestEvaluateTransaction(t *testing.T) {
	newEvaluateResponse := func(value []byte) *gateway.EvaluateResponse {
		return &gateway.EvaluateResponse{
			Result: &peer.Response{
				Payload: []byte(value),
			},
		}
	}

	t.Run("Returns evaluate error without wrapping to allow gRPC status to be interrogated", func(t *testing.T) {
		expected := NewStatusError(t, codes.Aborted, "EVALUATE_ERROR")
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
			Return(nil, expected)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.EvaluateTransaction("transaction")

		if err != expected {
			t.Fatalf("Expected unmodified invocation error, got: %v", err)
		}
	})

	t.Run("Returns result", func(t *testing.T) {
		expected := []byte("TRANSACTION_RESULT")
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
			Return(newEvaluateResponse(expected), nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		actual, err := contract.EvaluateTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(actual, expected) {
			t.Fatalf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("Includes channel name in proposal", func(t *testing.T) {
		var actual string
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EvaluateRequest, _ ...grpc.CallOption) {
				actual = test.AssertUnmarshallChannelheader(t, in.ProposedTransaction).ChannelId
			}).
			Return(newEvaluateResponse(nil), nil).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.EvaluateTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		expected := contract.channelName
		if actual != expected {
			t.Fatalf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("Includes chaincode ID in proposal", func(t *testing.T) {
		var actual string
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EvaluateRequest, _ ...grpc.CallOption) {
				actual = test.AssertUnmarshallInvocationSpec(t, in.ProposedTransaction).ChaincodeSpec.ChaincodeId.Name
			}).
			Return(newEvaluateResponse(nil), nil).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.EvaluateTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		expected := contract.chaincodeID
		if actual != expected {
			t.Fatalf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("Includes transaction name in proposal for default smart contract", func(t *testing.T) {
		var args [][]byte
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EvaluateRequest, _ ...grpc.CallOption) {
				args = test.AssertUnmarshallInvocationSpec(t, in.ProposedTransaction).ChaincodeSpec.Input.Args
			}).
			Return(newEvaluateResponse(nil), nil).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		expected := "TRANSACTION_NAME"
		_, err := contract.EvaluateTransaction(expected)
		if err != nil {
			t.Fatal(err)
		}

		actual := string(args[0])
		if actual != expected {
			t.Fatalf("Expected Args[0] to be %s, got Args: %s", expected, args)
		}
	})

	t.Run("Includes transaction name in proposal for named smart contract", func(t *testing.T) {
		var args [][]byte
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EvaluateRequest, _ ...grpc.CallOption) {
				args = test.AssertUnmarshallInvocationSpec(t, in.ProposedTransaction).ChaincodeSpec.Input.Args
			}).
			Return(newEvaluateResponse(nil), nil).
			Times(1)

		contract := AssertNewTestContractWithName(t, "chaincode", "CONTRACT_NAME", WithClient(mockClient))

		_, err := contract.EvaluateTransaction("TRANSACTION_NAME")
		if err != nil {
			t.Fatal(err)
		}

		actual := string(args[0])
		expected := "CONTRACT_NAME:TRANSACTION_NAME"
		if actual != expected {
			t.Fatalf("Expected Args[0] to be %s, got Args: %s", expected, args)
		}
	})

	t.Run("Includes arguments in proposal", func(t *testing.T) {
		var args [][]byte
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EvaluateRequest, _ ...grpc.CallOption) {
				args = test.AssertUnmarshallInvocationSpec(t, in.ProposedTransaction).ChaincodeSpec.Input.Args
			}).
			Return(newEvaluateResponse(nil), nil).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		expected := []string{"one", "two", "three"}
		_, err := contract.EvaluateTransaction("transaction", expected...)
		if err != nil {
			t.Fatal(err)
		}

		actual := bytesAsStrings(args[1:])
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("Expected Args[1:] to be %s, got Args: %s", expected, args)
		}
	})

	t.Run("Includes channel name in proposed transaction", func(t *testing.T) {
		var actual string
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EvaluateRequest, _ ...grpc.CallOption) {
				actual = in.ChannelId
			}).
			Return(newEvaluateResponse(nil), nil).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.EvaluateTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		expected := contract.channelName
		if actual != expected {
			t.Fatalf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("Includes transaction ID in proposed transaction", func(t *testing.T) {
		var actual string
		var expected string
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EvaluateRequest, _ ...grpc.CallOption) {
				actual = in.TransactionId
				expected = test.AssertUnmarshallChannelheader(t, in.ProposedTransaction).TxId
			}).
			Return(newEvaluateResponse(nil), nil).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.EvaluateTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		if actual != expected {
			t.Fatalf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("Uses sign", func(t *testing.T) {
		var actual []byte
		expected := []byte("MY_SIGNATURE")
		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EvaluateRequest, _ ...grpc.CallOption) {
				actual = in.ProposedTransaction.Signature
			}).
			Return(newEvaluateResponse(nil), nil).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient), WithSign(sign))

		_, err := contract.EvaluateTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(actual, expected) {
			t.Fatalf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("Uses hash", func(t *testing.T) {
		var actual []byte
		expected := []byte("MY_DIGEST")
		sign := func(digest []byte) ([]byte, error) {
			actual = digest
			return expected, nil
		}
		hash := func(message []byte) []byte {
			return expected
		}
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
			Return(newEvaluateResponse(nil), nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient), WithSign(sign), WithHash(hash))

		_, err := contract.EvaluateTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(actual, expected) {
			t.Fatalf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("Sends private data with evaluate", func(t *testing.T) {
		var actualOrgs []string
		expectedOrgs := []string{"MY_ORG"}
		var actualPrice []byte
		expectedPrice := []byte("3000")
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EvaluateRequest, _ ...grpc.CallOption) {
				actualOrgs = in.TargetOrganizations
				transient := test.AssertUnmarshallProposalPayload(t, in.ProposedTransaction).TransientMap
				actualPrice = transient["price"]
			}).
			Return(newEvaluateResponse(nil), nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		privateData := map[string][]byte{
			"price": []byte("3000"),
		}

		_, err := contract.Evaluate("transaction", WithTransient(privateData), WithEndorsingOrganizations("MY_ORG"))
		if err != nil {
			t.Fatal(err)
		}

		if len(actualOrgs) != 1 && expectedOrgs[0] != actualOrgs[0] {
			t.Fatalf("Expected %v, got %v", expectedOrgs, actualOrgs)
		}

		if !bytes.Equal(actualPrice, expectedPrice) {
			t.Fatalf("Expected %s, got %s", expectedPrice, actualPrice)
		}
	})
}
