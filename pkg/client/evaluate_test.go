/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
	"github.com/hyperledger/fabric-gateway/pkg/internal/test/mock"
	gateway "github.com/hyperledger/fabric-gateway/protos/gateway"
	"google.golang.org/grpc"
)

func TestEvaluateTransaction(t *testing.T) {
	t.Run("Returns gRPC invocation error", func(t *testing.T) {
		expectedError := "EVALUATE_ERROR"
		mockClient := mock.NewGatewayClient()
		mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
			return nil, errors.New(expectedError)
		}
		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.EvaluateTransaction("transaction")

		if nil == err || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Expected error containing %s, got %v", expectedError, err)
		}
	})

	t.Run("Returns result", func(t *testing.T) {
		expected := []byte("TRANSACTION_RESULT")
		mockClient := mock.NewGatewayClient()
		mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
			value := &gateway.Result{
				Value: expected,
			}
			return value, nil
		}
		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		actual, err := contract.EvaluateTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		if bytes.Compare(actual, expected) != 0 {
			t.Fatalf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("Includes channel name in proposal", func(t *testing.T) {
		var actual string
		mockClient := mock.NewGatewayClient()
		mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
			actual = test.AssertUnmarshallChannelheader(t, in).ChannelId
			value := &gateway.Result{}
			return value, nil
		}
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
		mockClient := mock.NewGatewayClient()
		mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
			actual = test.AssertUnmarshallInvocationSpec(t, in).ChaincodeSpec.ChaincodeId.Name
			value := &gateway.Result{}
			return value, nil
		}
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
		mockClient := mock.NewGatewayClient()
		mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
			args = test.AssertUnmarshallInvocationSpec(t, in).ChaincodeSpec.Input.Args
			value := &gateway.Result{}
			return value, nil
		}
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
		mockClient := mock.NewGatewayClient()
		mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
			args = test.AssertUnmarshallInvocationSpec(t, in).ChaincodeSpec.Input.Args
			value := &gateway.Result{}
			return value, nil
		}
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
		mockClient := mock.NewGatewayClient()
		mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
			args = test.AssertUnmarshallInvocationSpec(t, in).ChaincodeSpec.Input.Args
			value := &gateway.Result{}
			return value, nil
		}
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
		mockClient := mock.NewGatewayClient()
		mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
			actual = in.ChannelId
			value := &gateway.Result{}
			return value, nil
		}
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
		mockClient := mock.NewGatewayClient()
		mockClient.MockEvaluate = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.Result, error) {
			actual = in.TxId
			expected = test.AssertUnmarshallChannelheader(t, in).TxId
			value := &gateway.Result{}
			return value, nil
		}
		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.EvaluateTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		if actual != expected {
			t.Fatalf("Expected %s, got %s", expected, actual)
		}
	})
}
