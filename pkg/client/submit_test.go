/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/gateway"
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

	newSuccessSubmitClient := func(controller *gomock.Controller) *MockGateway_SubmitClient {
		mock := NewMockGateway_SubmitClient(controller)
		mock.EXPECT().Recv().
			Return(nil, io.EOF).
			AnyTimes()
		return mock
	}

	newFailSubmitClient := func(controller *gomock.Controller, err error) *MockGateway_SubmitClient {
		mock := NewMockGateway_SubmitClient(controller)
		mock.EXPECT().Recv().
			Return(nil, err).
			AnyTimes()
		return mock
	}

	t.Run("Returns endorsement error", func(t *testing.T) {
		expectedError := "ENDORSE_ERROR"
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(nil, errors.New(expectedError))

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")

		if nil == err || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Expected error containing %s, got %v", expectedError, err)
		}
	})

	t.Run("Returns error sending to orderer", func(t *testing.T) {
		expectedError := "SUBMIT_ERROR"
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newPreparedTransaction("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, errors.New(expectedError))

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")

		if nil == err || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Expected error containing %s, got %v", expectedError, err)
		}
	})

	t.Run("Returns commit error", func(t *testing.T) {
		expectedError := "COMMIT_ERROR"
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newPreparedTransaction("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(newFailSubmitClient(mockController, errors.New(expectedError)), nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")

		if nil == err || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Expected error containing %s, got %v", expectedError, err)
		}
	})

	t.Run("Returns result for committed transaction", func(t *testing.T) {
		expected := []byte("TRANSACTION_RESULT")
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newPreparedTransaction("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(newSuccessSubmitClient(mockController), nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		actual, err := contract.SubmitTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(actual, expected) {
			t.Fatalf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("Includes channel name in proposal", func(t *testing.T) {
		var actual string
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.ProposedTransaction) {
				actual = test.AssertUnmarshallChannelheader(t, in).ChannelId
			}).
			Return(newPreparedTransaction("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(newSuccessSubmitClient(mockController), nil)

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
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.ProposedTransaction) {
				actual = test.AssertUnmarshallInvocationSpec(t, in).ChaincodeSpec.ChaincodeId.Name
			}).
			Return(newPreparedTransaction("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(newSuccessSubmitClient(mockController), nil)

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
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.ProposedTransaction) {
				args = test.AssertUnmarshallInvocationSpec(t, in).ChaincodeSpec.Input.Args
			}).
			Return(newPreparedTransaction("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(newSuccessSubmitClient(mockController), nil)

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
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.ProposedTransaction) {
				args = test.AssertUnmarshallInvocationSpec(t, in).ChaincodeSpec.Input.Args
			}).
			Return(newPreparedTransaction("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(newSuccessSubmitClient(mockController), nil)

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
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.ProposedTransaction) {
				args = test.AssertUnmarshallInvocationSpec(t, in).ChaincodeSpec.Input.Args
			}).
			Return(newPreparedTransaction("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(newSuccessSubmitClient(mockController), nil)

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

	t.Run("Includes channel name in proposed transaction", func(t *testing.T) {
		var actual string
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.ProposedTransaction) {
				actual = in.ChannelId
			}).
			Return(newPreparedTransaction("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(newSuccessSubmitClient(mockController), nil)

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

	t.Run("Includes transaction ID in proposed transaction", func(t *testing.T) {
		var actual string
		var expected string
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.ProposedTransaction) {
				actual = in.TxId
				expected = test.AssertUnmarshallChannelheader(t, in).TxId
			}).
			Return(newPreparedTransaction("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(newSuccessSubmitClient(mockController), nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		if actual != expected {
			t.Fatalf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("Uses signer for endorse", func(t *testing.T) {
		var actual []byte
		expected := []byte("MY_SIGNATURE")
		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.ProposedTransaction) {
				actual = in.Proposal.Signature
			}).
			Return(newPreparedTransaction("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(newSuccessSubmitClient(mockController), nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient), WithSign(sign))

		_, err := contract.SubmitTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(actual, expected) {
			t.Fatalf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("Uses signer for submit", func(t *testing.T) {
		var actual []byte
		expected := []byte("MY_SIGNATURE")
		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newPreparedTransaction("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.PreparedTransaction) {
				actual = in.Envelope.Signature
			}).
			Return(newSuccessSubmitClient(mockController), nil).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient), WithSign(sign))

		_, err := contract.SubmitTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(actual, expected) {
			t.Fatalf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("Uses hash", func(t *testing.T) {
		var actual [][]byte
		expected := []byte("MY_DIGEST")
		sign := func(digest []byte) ([]byte, error) {
			actual = append(actual, digest)
			return expected, nil
		}
		hash := func(message []byte) ([]byte, error) {
			return expected, nil
		}
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newPreparedTransaction("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(newSuccessSubmitClient(mockController), nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient), WithSign(sign), WithHash(hash))

		_, err := contract.SubmitTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		if len(actual) != 2 {
			t.Fatalf("Expected 2 signatures, got %v", len(actual))
		}
		for i, digest := range actual {
			if !bytes.Equal(digest, expected) {
				t.Fatalf("Expected %s for call %v, got %s", expected, i, digest)
			}
		}
	})
}
