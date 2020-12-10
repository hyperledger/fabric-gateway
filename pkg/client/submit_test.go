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
	gateway "github.com/hyperledger/fabric-gateway/protos"
	"github.com/hyperledger/fabric-protos-go/common"
	"google.golang.org/grpc"
)

func TestSubmitTransaction(t *testing.T) {
	newPreparedTransaction := func(value string) *gateway.PreparedTransaction {
		return &gateway.PreparedTransaction{
			Envelope: &common.Envelope{},
			Response: &gateway.Result{
				Value: []byte(value),
			},
		}
	}

	t.Run("Returns endorsement error", func(t *testing.T) {
		expectedError := "ENDORSE_ERROR"
		mockClient := mock.NewGatewayClient()
		mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
			return nil, errors.New(expectedError)
		}
		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")

		if nil == err || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Expected error containing %s, got %v", expectedError, err)
		}
	})

	t.Run("Returns error sending to orderer", func(t *testing.T) {
		expectedError := "SUBMIT_ERROR"
		mockClient := mock.NewGatewayClient()
		mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
			return newPreparedTransaction("TRANSACTION_RESULT"), nil
		}
		mockClient.MockSubmit = func(ctx context.Context, in *gateway.PreparedTransaction, opts ...grpc.CallOption) (gateway.Gateway_SubmitClient, error) {
			return nil, errors.New(expectedError)
		}
		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")

		if nil == err || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Expected error containing %s, got %v", expectedError, err)
		}
	})

	t.Run("Returns commit error", func(t *testing.T) {
		expectedError := "COMMIT_ERROR"
		mockClient := mock.NewGatewayClient()
		mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
			return newPreparedTransaction("TRANSACTION_RESULT"), nil
		}
		mockClient.MockSubmit = func(ctx context.Context, in *gateway.PreparedTransaction, opts ...grpc.CallOption) (gateway.Gateway_SubmitClient, error) {
			return mock.NewFailSubmitClient(errors.New(expectedError)), nil
		}
		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")

		if nil == err || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Expected error containing %s, got %v", expectedError, err)
		}
	})

	t.Run("Returns result for committed transaction", func(t *testing.T) {
		expected := []byte("TRANSACTION_RESULT")
		mockClient := mock.NewGatewayClient()
		mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
			return newPreparedTransaction("TRANSACTION_RESULT"), nil
		}
		mockClient.MockSubmit = func(ctx context.Context, in *gateway.PreparedTransaction, opts ...grpc.CallOption) (gateway.Gateway_SubmitClient, error) {
			return mock.NewSuccessSubmitClient(), nil
		}

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		actual, err := contract.SubmitTransaction("transaction")
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
		mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
			actual = test.AssertUnmarshallChannelheader(t, in).ChannelId
			return newPreparedTransaction("TRANSACTION_RESULT"), nil
		}
		mockClient.MockSubmit = func(ctx context.Context, in *gateway.PreparedTransaction, opts ...grpc.CallOption) (gateway.Gateway_SubmitClient, error) {
			return mock.NewSuccessSubmitClient(), nil
		}
		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")
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
		mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
			actual = test.AssertUnmarshallInvocationSpec(t, in).ChaincodeSpec.ChaincodeId.Name
			return newPreparedTransaction("TRANSACTION_RESULT"), nil
		}
		mockClient.MockSubmit = func(ctx context.Context, in *gateway.PreparedTransaction, opts ...grpc.CallOption) (gateway.Gateway_SubmitClient, error) {
			return mock.NewSuccessSubmitClient(), nil
		}
		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		expected := contract.chaincodeID
		if actual != expected {
			t.Fatalf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("Includes transaction name in proposal for default contract", func(t *testing.T) {
		var args [][]byte
		mockClient := mock.NewGatewayClient()
		mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
			args = test.AssertUnmarshallInvocationSpec(t, in).ChaincodeSpec.Input.Args
			return newPreparedTransaction("TRANSACTION_RESULT"), nil
		}
		mockClient.MockSubmit = func(ctx context.Context, in *gateway.PreparedTransaction, opts ...grpc.CallOption) (gateway.Gateway_SubmitClient, error) {
			return mock.NewSuccessSubmitClient(), nil
		}
		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		expected := "TRANSACTION_NAME"
		_, err := contract.SubmitTransaction(expected)
		if err != nil {
			t.Fatal(err)
		}

		actual := string(args[0])
		if actual != expected {
			t.Fatalf("Expected Args[0] to be %s, got Args: %s", expected, args)
		}
	})

	t.Run("Includes transaction name in proposal for named contract", func(t *testing.T) {
		var args [][]byte
		mockClient := mock.NewGatewayClient()
		mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
			args = test.AssertUnmarshallInvocationSpec(t, in).ChaincodeSpec.Input.Args
			return newPreparedTransaction("TRANSACTION_RESULT"), nil
		}
		mockClient.MockSubmit = func(ctx context.Context, in *gateway.PreparedTransaction, opts ...grpc.CallOption) (gateway.Gateway_SubmitClient, error) {
			return mock.NewSuccessSubmitClient(), nil
		}
		contract := AssertNewTestContractWithName(t, "chaincode", "CONTRACT_NAME", WithClient(mockClient))

		_, err := contract.SubmitTransaction("TRANSACTION_NAME")
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
		mockClient.MockEndorse = func(ctx context.Context, in *gateway.ProposedTransaction, opts ...grpc.CallOption) (*gateway.PreparedTransaction, error) {
			args = test.AssertUnmarshallInvocationSpec(t, in).ChaincodeSpec.Input.Args
			return newPreparedTransaction("TRANSACTION_RESULT"), nil
		}
		mockClient.MockSubmit = func(ctx context.Context, in *gateway.PreparedTransaction, opts ...grpc.CallOption) (gateway.Gateway_SubmitClient, error) {
			return mock.NewSuccessSubmitClient(), nil
		}
		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		expected := []string{"one", "two", "three"}
		_, err := contract.SubmitTransaction("transaction", expected...)
		if err != nil {
			t.Fatal(err)
		}

		actual := bytesAsStrings(args[1:])
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("Expected Args[1:] to be %s, got Args: %s", expected, args)
		}
	})

}
