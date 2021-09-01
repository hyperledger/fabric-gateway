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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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

	newCommitStatusResponse := func(status peer.TxValidationCode, blockNumber uint64) *gateway.CommitStatusResponse {
		return &gateway.CommitStatusResponse{
			Result:      status,
			BlockNumber: blockNumber,
		}
	}

	t.Run("Returns endorse error without wrapping to allow gRPC status to be interrogated", func(t *testing.T) {
		expected := NewStatusError(t, codes.Aborted, "ENDORSE_ERROR")
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(nil, expected)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")

		if err != expected {
			t.Fatalf("Expected unmodified invocation error, got: %v", err)
		}
	})

	t.Run("Returns submit error without wrapping to allow gRPC status to be interrogated", func(t *testing.T) {
		expected := NewStatusError(t, codes.Aborted, "SUBMIT_ERROR")
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, expected)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")

		if err != expected {
			t.Fatalf("Expected unmodified invocation error, got: %v", err)
		}
	})

	t.Run("Returns commit error without wrapping to allow gRPC status to be interrogated", func(t *testing.T) {
		expected := NewStatusError(t, codes.Aborted, "COMMIT_ERROR")
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(nil, expected)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")

		if err != expected {
			t.Fatalf("Expected unmodified invocation error, got: %v", err)
		}
	})

	t.Run("Returns result for committed transaction", func(t *testing.T) {
		expected := []byte("TRANSACTION_RESULT")
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

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
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_MVCC_READ_CONFLICT, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")

		if nil == err || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Expected error containing %s, got %v", expectedError, err)
		}
	})

	t.Run("Returns error with details on communication failure getting transaction commit status", func(t *testing.T) {
		expectedError := "COMMIT_STATUS_ERROR"
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), errors.New(expectedError))

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")

		if nil == err || !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Expected error containing %s, got %v", expectedError, err)
		}
	})

	t.Run("Includes channel name in proposal", func(t *testing.T) {
		var actual string
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				actual = test.AssertUnmarshallChannelheader(t, in.ProposedTransaction).ChannelId
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

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
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				actual = test.AssertUnmarshallInvocationSpec(t, in.ProposedTransaction).ChaincodeSpec.ChaincodeId.Name
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

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
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				args = test.AssertUnmarshallInvocationSpec(t, in.ProposedTransaction).ChaincodeSpec.Input.Args
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

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
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				args = test.AssertUnmarshallInvocationSpec(t, in.ProposedTransaction).ChaincodeSpec.Input.Args
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

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
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				args = test.AssertUnmarshallInvocationSpec(t, in.ProposedTransaction).ChaincodeSpec.Input.Args
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

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
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				actual = in.ChannelId
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

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
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				actual = in.TransactionId
				expected = test.AssertUnmarshallChannelheader(t, in.ProposedTransaction).TxId
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

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
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SignedCommitStatusRequest, _ ...grpc.CallOption) {
				request := &gateway.CommitStatusRequest{}
				test.AssertUnmarshall(t, in.Request, request)
				actual = request.ChannelId
			}).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil).
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
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				expected = test.AssertUnmarshallChannelheader(t, in.ProposedTransaction).TxId
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SignedCommitStatusRequest, _ ...grpc.CallOption) {
				request := &gateway.CommitStatusRequest{}
				test.AssertUnmarshall(t, in.Request, request)
				actual = request.TransactionId
			}).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil).
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
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				actual = in.ProposedTransaction.Signature
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

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
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SubmitRequest, _ ...grpc.CallOption) {
				actual = in.PreparedTransaction.Signature
			}).
			Return(nil, nil).
			Times(1)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient), WithSign(sign))

		_, err := contract.SubmitTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(actual, expected) {
			t.Fatalf("Expected %s, got %s", expected, actual)
		}
	})

	t.Run("Sends private data with submit", func(t *testing.T) {
		var actualOrgs []string
		expectedOrgs := []string{"MY_ORG"}
		var actualPrice []byte
		expectedPrice := []byte("3000")
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				actualOrgs = in.EndorsingOrganizations
				transient := test.AssertUnmarshallProposalPayload(t, in.ProposedTransaction).TransientMap
				actualPrice = transient["price"]
			}).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		privateData := map[string][]byte{
			"price": []byte("3000"),
		}

		_, err := contract.Submit("transaction", WithTransient(privateData), WithEndorsingOrganizations("MY_ORG"))
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

	t.Run("Uses signer for commit status", func(t *testing.T) {
		var actual []byte
		expected := []byte("MY_SIGNATURE")
		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SignedCommitStatusRequest, _ ...grpc.CallOption) {
				actual = in.Signature
			}).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil).
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
		hash := func(message []byte) []byte {
			return expected
		}
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient), WithSign(sign), WithHash(hash))

		_, err := contract.SubmitTransaction("transaction")
		if err != nil {
			t.Fatal(err)
		}

		if len(actual) != 3 { // Endorse, Submit and CommitStatus
			t.Fatalf("Expected 3 signatures, got %v", len(actual))
		}
		for i, digest := range actual {
			if !bytes.Equal(digest, expected) {
				t.Fatalf("Expected %s for call %v, got %s", expected, i, digest)
			}
		}
	})

	t.Run("Commit returns transaction status", func(t *testing.T) {
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_MVCC_READ_CONFLICT, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, commit, err := contract.SubmitAsync("transaction")
		if err != nil {
			t.Fatal(err)
		}

		status, err := commit.Status()
		if err != nil {
			t.Fatal(err)
		}

		if status != peer.TxValidationCode_MVCC_READ_CONFLICT {
			t.Fatalf("Expected status %v, got %v", peer.TxValidationCode_MVCC_READ_CONFLICT, status)
		}
	})

	t.Run("Commit returns successful for successful transaction", func(t *testing.T) {
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, commit, err := contract.SubmitAsync("transaction")
		if err != nil {
			t.Fatal(err)
		}

		success, err := commit.Successful()
		if err != nil {
			t.Fatal(err)
		}

		if !success {
			t.Fatal("Expected successful, got unsuccessful")
		}
	})

	t.Run("Commit returns unsuccessful for failed transaction", func(t *testing.T) {
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_MVCC_READ_CONFLICT, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, commit, err := contract.SubmitAsync("transaction")
		if err != nil {
			t.Fatal(err)
		}

		success, err := commit.Successful()
		if err != nil {
			t.Fatal(err)
		}

		if success {
			t.Fatal("Expected unsuccessful, got successful")
		}
	})

	t.Run("Commit returns block number", func(t *testing.T) {
		expectedBlockNumber := uint64(101)
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(newEndorseResponse("TRANSACTION_RESULT"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_MVCC_READ_CONFLICT, expectedBlockNumber), nil)

		contract := AssertNewTestContract(t, "chaincode", WithClient(mockClient))

		_, commit, err := contract.SubmitAsync("transaction")
		if err != nil {
			t.Fatal(err)
		}

		blockNumber, err := commit.BlockNumber()
		if err != nil {
			t.Fatal(err)
		}

		if blockNumber != expectedBlockNumber {
			t.Fatalf("Expected block number %v, got %v", expectedBlockNumber, blockNumber)
		}
	})
}
