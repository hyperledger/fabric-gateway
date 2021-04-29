/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"bytes"
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/gateway"
	"github.com/hyperledger/fabric-protos-go/peer"
	"google.golang.org/grpc"
)

func TestOfflineSign(t *testing.T) {
	evaluateResponse := gateway.EvaluateResponse{
		Result: &peer.Response{
			Payload: nil,
		},
	}

	newNetworkWithNoSign := func(t *testing.T, options ...ConnectOption) *Network {
		gateway, err := Connect(TestCredentials.identity, options...)
		if err != nil {
			t.Fatal(err)
		}

		return gateway.GetNetwork("network")
	}

	newContractWithNoSign := func(t *testing.T, options ...ConnectOption) *Contract {
		return newNetworkWithNoSign(t, options...).GetContract("chaincode")
	}

	newEndorseResponse := func(value string) *gateway.EndorseResponse {
		return &gateway.EndorseResponse{
			PreparedTransaction: &common.Envelope{},
			Result: &peer.Response{
				Payload: []byte(value),
			},
		}
	}

	newCommitStatusResponse := func(value peer.TxValidationCode) *gateway.CommitStatusResponse {
		return &gateway.CommitStatusResponse{
			Result: value,
		}
	}

	t.Run("Evaluate", func(t *testing.T) {
		t.Run("Returns error with no signer and no explicit signing", func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
				Return(&evaluateResponse, nil).
				AnyTimes()

			contract := newContractWithNoSign(t, WithClient(mockClient))

			proposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			if _, err := proposal.Evaluate(); nil == err {
				t.Fatal("Expected signing error but got nil")
			}
		})

		t.Run("Uses off-line signature", func(t *testing.T) {
			expected := []byte("SIGNATURE")
			var actual []byte
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, in *gateway.EvaluateRequest, _ ...grpc.CallOption) {
					actual = in.ProposedTransaction.Signature
				}).
				Return(&evaluateResponse, nil).
				Times(1)

			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			proposalBytes, err := unsignedProposal.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedProposal, err := contract.NewSignedProposal(proposalBytes, expected)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := signedProposal.Evaluate(); err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(actual, expected) {
				t.Fatalf("Expected %s, got %s", expected, actual)
			}
		})
	})

	t.Run("Endorse", func(t *testing.T) {
		t.Run("Returns error with no signer and no explicit signing", func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(newEndorseResponse("result"), nil).
				AnyTimes()

			contract := newContractWithNoSign(t, WithClient(mockClient))

			proposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			if _, err := proposal.Endorse(); nil == err {
				t.Fatal("Expected signing error but got nil")
			}
		})

		t.Run("Uses off-line signature", func(t *testing.T) {
			expected := []byte("SIGNATURE")
			var actual []byte
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
					actual = in.ProposedTransaction.Signature
				}).
				Return(newEndorseResponse("result"), nil).
				Times(1)

			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			proposalBytes, err := unsignedProposal.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedProposal, err := contract.NewSignedProposal(proposalBytes, expected)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := signedProposal.Endorse(); err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(actual, expected) {
				t.Fatalf("Expected %s, got %s", expected, actual)
			}
		})
	})

	t.Run("Submit", func(t *testing.T) {
		t.Run("Returns error with no signer and no explicit signing", func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(newEndorseResponse("result"), nil).
				AnyTimes()
			mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				AnyTimes()

			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			proposalBytes, err := unsignedProposal.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("signature"))
			if err != nil {
				t.Fatal(err)
			}

			transaction, err := signedProposal.Endorse()
			if err != nil {
				t.Fatal(err)
			}

			if _, err := transaction.Submit(); nil == err {
				t.Fatal("Expected signing error but got nil")
			}
		})

		t.Run("Uses off-line signature", func(t *testing.T) {
			expected := []byte("SIGNATURE")
			var actual []byte
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(newEndorseResponse("result"), nil)
			mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, in *gateway.SubmitRequest, _ ...grpc.CallOption) {
					actual = in.PreparedTransaction.Signature
				}).
				Return(nil, nil).
				Times(1)

			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			proposalBytes, err := unsignedProposal.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedProposal, err := contract.NewSignedProposal(proposalBytes, expected)
			if err != nil {
				t.Fatal(err)
			}

			unsignedTransaction, err := signedProposal.Endorse()
			if err != nil {
				t.Fatal(err)
			}

			transactionBytes, err := unsignedTransaction.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedTransaction, err := contract.NewSignedTransaction(transactionBytes, expected)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := signedTransaction.Submit(); err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(actual, expected) {
				t.Fatalf("Expected %s, got %s", expected, actual)
			}
		})
	})

	t.Run("Commit", func(t *testing.T) {
		t.Run("Returns error with no signer and no explicit signing", func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(newEndorseResponse("result"), nil).
				AnyTimes()
			mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				AnyTimes()
			mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				AnyTimes()

			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			proposalBytes, err := unsignedProposal.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("signature"))
			if err != nil {
				t.Fatal(err)
			}

			unsignedTransaction, err := signedProposal.Endorse()
			if err != nil {
				t.Fatal(err)
			}

			transactionBytes, err := unsignedTransaction.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedTransaction, err := contract.NewSignedTransaction(transactionBytes, []byte("signature"))
			if err != nil {
				t.Fatal(err)
			}

			commit, err := signedTransaction.Submit()
			if err != nil {
				t.Fatal(err)
			}

			if _, err := commit.Status(); nil == err {
				t.Fatal("Expected signing error but got nil")
			}
		})

		t.Run("Uses off-line signature", func(t *testing.T) {
			expected := []byte("SIGNATURE")
			var actual []byte
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(newEndorseResponse("result"), nil)
			mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				AnyTimes()
			mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, in *gateway.SignedCommitStatusRequest, _ ...grpc.CallOption) {
					actual = in.Signature
				}).
				Return(newCommitStatusResponse(peer.TxValidationCode_VALID), nil).
				Times(1)

			network := newNetworkWithNoSign(t, WithClient(mockClient))
			contract := network.GetContract("chaincode")

			unsignedProposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			proposalBytes, err := unsignedProposal.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedProposal, err := contract.NewSignedProposal(proposalBytes, expected)
			if err != nil {
				t.Fatal(err)
			}

			unsignedTransaction, err := signedProposal.Endorse()
			if err != nil {
				t.Fatal(err)
			}

			transactionBytes, err := unsignedTransaction.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedTransaction, err := contract.NewSignedTransaction(transactionBytes, expected)
			if err != nil {
				t.Fatal(err)
			}

			unsignedCommit, err := signedTransaction.Submit()
			if err != nil {
				t.Fatal(err)
			}

			commitBytes, err := unsignedCommit.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedCommit, err := network.NewSignedCommit(commitBytes, expected)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := signedCommit.Status(); err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(actual, expected) {
				t.Fatalf("Expected %s, got %s", expected, actual)
			}
		})
	})

	t.Run("Serialization", func(t *testing.T) {
		t.Run("Proposal keeps same digest", func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			proposalBytes, err := unsignedProposal.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("signature"))
			if err != nil {
				t.Fatal(err)
			}

			expected := unsignedProposal.Digest()
			actual := signedProposal.Digest()

			if !bytes.Equal(actual, expected) {
				t.Fatalf("Expected %s, got %s", expected, actual)
			}
		})

		t.Run("Proposal keeps same transaction ID", func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			proposalBytes, err := unsignedProposal.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("signature"))
			if err != nil {
				t.Fatal(err)
			}

			expected := unsignedProposal.TransactionID()
			actual := signedProposal.TransactionID()

			if actual != expected {
				t.Fatalf("Expected %s, got %s", expected, actual)
			}
		})

		t.Run("Transaction keeps same digest", func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(newEndorseResponse("result"), nil).
				Times(1)

			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			proposalBytes, err := unsignedProposal.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("signature"))
			if err != nil {
				t.Fatal(err)
			}

			unsignedTransaction, err := signedProposal.Endorse()
			if err != nil {
				t.Fatal(err)
			}

			transactionBytes, err := unsignedTransaction.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedTransaction, err := contract.NewSignedTransaction(transactionBytes, []byte("signature"))
			if err != nil {
				t.Fatal(err)
			}

			expected := unsignedTransaction.Digest()
			actual := signedTransaction.Digest()

			if !bytes.Equal(actual, expected) {
				t.Fatalf("Expected %s, got %s", expected, actual)
			}
		})

		t.Run("Commit keeps same digest", func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockClient := NewMockGatewayClient(mockController)
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(newEndorseResponse("result"), nil).
				Times(1)
			mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				Times(1)

			network := newNetworkWithNoSign(t, WithClient(mockClient))
			contract := network.GetContract("chaincode")

			unsignedProposal, err := contract.NewProposal("transaction")
			if err != nil {
				t.Fatal(err)
			}

			proposalBytes, err := unsignedProposal.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("signature"))
			if err != nil {
				t.Fatal(err)
			}

			unsignedTransaction, err := signedProposal.Endorse()
			if err != nil {
				t.Fatal(err)
			}

			transactionBytes, err := unsignedTransaction.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedTransaction, err := contract.NewSignedTransaction(transactionBytes, []byte("signature"))
			if err != nil {
				t.Fatal(err)
			}

			unsignedCommit, err := signedTransaction.Submit()
			if err != nil {
				t.Fatal(err)
			}

			commitBytes, err := unsignedCommit.Bytes()
			if err != nil {
				t.Fatal(err)
			}

			signedCommit, err := network.NewSignedCommit(commitBytes, []byte("signature"))
			if err != nil {
				t.Fatal(err)
			}

			expected := unsignedCommit.Digest()
			actual := signedCommit.Digest()

			if !bytes.Equal(actual, expected) {
				t.Fatalf("Expected %s, got %s", expected, actual)
			}
		})
	})
}
