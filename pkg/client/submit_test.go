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

	gomock "github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/gateway"
	"github.com/hyperledger/fabric-protos-go/peer"
)

func TestSubmitTransaction(t *testing.T) {
	newEndorseResponse := func(value string) *gateway.EndorseResponse {
		return &gateway.EndorseResponse{
			PreparedTransaction: &common.Envelope{},
			Result: &peer.Response{
				Payload: []byte(value),
			},
		}
	}

	validStatusResponse := gateway.CommitStatusResponse{
		Result: peer.TxValidationCode_VALID,
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
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
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
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, errors.New(expectedError))

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
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(&validStatusResponse, nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		actual, err := contract.SubmitTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(actual, expected) {
			t.Fatalf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("Returns error with status code for commit failure", func(t *testing.T) {
		expectedError := peer.TxValidationCode_name[int32(peer.TxValidationCode_MVCC_READ_CONFLICT)]
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		readConflictResponse := &gateway.CommitStatusResponse{
			Result: peer.TxValidationCode_MVCC_READ_CONFLICT,
		}
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(readConflictResponse, nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")

		if nil == err || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Expected error containing %s, got %v", expectedError, err)
		}
	})

	t.Run("Returns error with details on communication failure getting transaction commit status", func(t *testing.T) {
		expectedError := "COMMIT_STATUS_ERROR"
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(&validStatusResponse, errors.New(expectedError))

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")

		if nil == err || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Expected error containing %s, got %v", expectedError, err)
		}
	})

	t.Run("Includes channel name in proposal", func(t *testing.T) {
		var actual string
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest) {
				actual = test.AssertUnmarshallChannelheader(t, in.ProposedTransaction).ChannelId
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(&validStatusResponse, nil)

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
			Do(func(_ context.Context, in *gateway.EndorseRequest) {
				actual = test.AssertUnmarshallInvocationSpec(t, in.ProposedTransaction).ChaincodeSpec.ChaincodeId.Name
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(&validStatusResponse, nil)

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
			Do(func(_ context.Context, in *gateway.EndorseRequest) {
				args = test.AssertUnmarshallInvocationSpec(t, in.ProposedTransaction).ChaincodeSpec.Input.Args
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(&validStatusResponse, nil)

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
			Do(func(_ context.Context, in *gateway.EndorseRequest) {
				args = test.AssertUnmarshallInvocationSpec(t, in.ProposedTransaction).ChaincodeSpec.Input.Args
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(&validStatusResponse, nil)

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
			Do(func(_ context.Context, in *gateway.EndorseRequest) {
				args = test.AssertUnmarshallInvocationSpec(t, in.ProposedTransaction).ChaincodeSpec.Input.Args
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(&validStatusResponse, nil)

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
			Do(func(_ context.Context, in *gateway.EndorseRequest) {
				actual = in.ChannelId
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(&validStatusResponse, nil)

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
			Do(func(_ context.Context, in *gateway.EndorseRequest) {
				actual = in.TransactionId
				expected = test.AssertUnmarshallChannelheader(t, in.ProposedTransaction).TxId
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(&validStatusResponse, nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		if actual != expected {
			t.Fatalf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("Includes channel name in commit status request", func(t *testing.T) {
		var actual string
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.CommitStatusRequest) {
				actual = in.ChannelId
			}).
			Return(&validStatusResponse, nil).
			Times(1)

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

	t.Run("Includes transaction ID in commit status request", func(t *testing.T) {
		var actual string
		var expected string
		mockController := gomock.NewController(t)
		defer mockController.Finish()

		mockClient := NewMockGatewayClient(mockController)
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest) {
				expected = test.AssertUnmarshallChannelheader(t, in.ProposedTransaction).TxId
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.CommitStatusRequest) {
				actual = in.TransactionId
			}).
			Return(&validStatusResponse, nil).
			Times(1)

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
			Do(func(_ context.Context, in *gateway.EndorseRequest) {
				actual = in.ProposedTransaction.Signature
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(&validStatusResponse, nil)

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
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SubmitRequest) {
				actual = in.PreparedTransaction.Signature
			}).
			Return(nil, nil).
			Times(1)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(&validStatusResponse, nil)

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
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(&validStatusResponse, nil)

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
