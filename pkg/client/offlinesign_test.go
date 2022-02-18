/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-protos-go/gateway"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestOfflineSign(t *testing.T) {
	evaluateResponse := gateway.EvaluateResponse{
		Result: &peer.Response{
			Payload: nil,
		},
	}

	newContractWithNoSign := func(t *testing.T, options ...ConnectOption) (*Gateway, *Contract) {
		defaultOptions := []ConnectOption{
			WithDeliverClient(NewMockDeliverClient(gomock.NewController(t))),
		}
		options = append(defaultOptions, options...)
		gateway, err := Connect(TestCredentials.Identity(), options...)
		require.NoError(t, err)

		contract := gateway.GetNetwork("network").GetContract("contract")

		return gateway, contract
	}

	newCommitStatusResponse := func(value peer.TxValidationCode) *gateway.CommitStatusResponse {
		return &gateway.CommitStatusResponse{
			Result: value,
		}
	}

	t.Run("Evaluate", func(t *testing.T) {
		t.Run("Returns error with no signer and no explicit signing", func(t *testing.T) {
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
				Return(&evaluateResponse, nil).
				AnyTimes()

			_, contract := newContractWithNoSign(t, WithGatewayClient(mockClient))

			proposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			_, err = proposal.Evaluate()
			require.Error(t, err)
		})

		t.Run("Uses off-line signature", func(t *testing.T) {
			expected := []byte("SIGNATURE")
			var actual []byte
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, in *gateway.EvaluateRequest, _ ...grpc.CallOption) {
					actual = in.ProposedTransaction.Signature
				}).
				Return(&evaluateResponse, nil).
				Times(1)

			gateway, contract := newContractWithNoSign(t, WithGatewayClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := gateway.NewSignedProposal(proposalBytes, expected)
			require.NoError(t, err)

			_, err = signedProposal.Evaluate()
			require.NoError(t, err)

			require.EqualValues(t, expected, actual)
		})

		t.Run("Uses off-line signature with endorsing orgs", func(t *testing.T) {
			var actual []string
			expected := []string{"MY_ORG"}

			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, in *gateway.EvaluateRequest, _ ...grpc.CallOption) {
					actual = in.TargetOrganizations
				}).
				Return(&evaluateResponse, nil).
				Times(1)

			gateway, contract := newContractWithNoSign(t, WithGatewayClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction", WithEndorsingOrganizations("MY_ORG"))
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := gateway.NewSignedProposal(proposalBytes, []byte("SIGNATURE"))
			require.NoError(t, err)

			_, err = signedProposal.Evaluate()
			require.NoError(t, err)

			require.EqualValues(t, expected, actual)
		})
	})

	t.Run("Endorse", func(t *testing.T) {
		t.Run("Returns error with no signer and no explicit signing", func(t *testing.T) {
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(AssertNewEndorseResponse(t, "result", "network"), nil).
				AnyTimes()

			_, contract := newContractWithNoSign(t, WithGatewayClient(mockClient))

			proposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			_, err = proposal.Endorse()
			require.Error(t, err)
		})

		t.Run("Uses off-line signature", func(t *testing.T) {
			expected := []byte("SIGNATURE")
			var actual []byte
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
					actual = in.ProposedTransaction.Signature
				}).
				Return(AssertNewEndorseResponse(t, "result", "network"), nil).
				Times(1)

			gateway, contract := newContractWithNoSign(t, WithGatewayClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := gateway.NewSignedProposal(proposalBytes, expected)
			require.NoError(t, err)

			_, err = signedProposal.Endorse()
			require.NoError(t, err)

			require.EqualValues(t, expected, actual)
		})

		t.Run("Uses off-line signature with endorsing orgs", func(t *testing.T) {
			var actual []string
			expected := []string{"MY_ORG"}

			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
					actual = in.EndorsingOrganizations
				}).
				Return(AssertNewEndorseResponse(t, "result", "network"), nil).
				Times(1)

			gateway, contract := newContractWithNoSign(t, WithGatewayClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction", WithEndorsingOrganizations("MY_ORG"))
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := gateway.NewSignedProposal(proposalBytes, []byte("SIGNATURE"))
			require.NoError(t, err)

			_, err = signedProposal.Endorse()
			require.NoError(t, err)

			require.EqualValues(t, expected, actual)
		})
	})

	t.Run("Submit", func(t *testing.T) {
		t.Run("Returns error with no signer and no explicit signing", func(t *testing.T) {
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(AssertNewEndorseResponse(t, "result", "network"), nil).
				AnyTimes()
			mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				AnyTimes()

			gateway, contract := newContractWithNoSign(t, WithGatewayClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := gateway.NewSignedProposal(proposalBytes, []byte("signature"))
			require.NoError(t, err)

			transaction, err := signedProposal.Endorse()
			require.NoError(t, err)

			_, err = transaction.Submit()
			require.Error(t, err)
		})

		t.Run("Uses off-line signature", func(t *testing.T) {
			expected := []byte("SIGNATURE")
			var actual []byte
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(AssertNewEndorseResponse(t, "result", "network"), nil)
			mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, in *gateway.SubmitRequest, _ ...grpc.CallOption) {
					actual = in.PreparedTransaction.Signature
				}).
				Return(nil, nil).
				Times(1)

			gateway, contract := newContractWithNoSign(t, WithGatewayClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := gateway.NewSignedProposal(proposalBytes, expected)
			require.NoError(t, err)

			unsignedTransaction, err := signedProposal.Endorse()
			require.NoError(t, err)

			transactionBytes, err := unsignedTransaction.Bytes()
			require.NoError(t, err)

			signedTransaction, err := gateway.NewSignedTransaction(transactionBytes, expected)
			require.NoError(t, err)

			_, err = signedTransaction.Submit()
			require.NoError(t, err)

			require.EqualValues(t, expected, actual)
		})
	})

	t.Run("Commit", func(t *testing.T) {
		t.Run("Returns error with no signer and no explicit signing", func(t *testing.T) {
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(AssertNewEndorseResponse(t, "result", "network"), nil).
				AnyTimes()
			mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				AnyTimes()
			mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				AnyTimes()

			gateway, contract := newContractWithNoSign(t, WithGatewayClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := gateway.NewSignedProposal(proposalBytes, []byte("signature"))
			require.NoError(t, err)

			unsignedTransaction, err := signedProposal.Endorse()
			require.NoError(t, err)

			transactionBytes, err := unsignedTransaction.Bytes()
			require.NoError(t, err)

			signedTransaction, err := gateway.NewSignedTransaction(transactionBytes, []byte("signature"))
			require.NoError(t, err)

			commit, err := signedTransaction.Submit()
			require.NoError(t, err)

			_, err = commit.Status()
			require.Error(t, err)
		})

		t.Run("Uses off-line signature", func(t *testing.T) {
			expected := []byte("SIGNATURE")
			var actual []byte
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(AssertNewEndorseResponse(t, "result", "network"), nil)
			mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				AnyTimes()
			mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
				Do(func(_ context.Context, in *gateway.SignedCommitStatusRequest, _ ...grpc.CallOption) {
					actual = in.Signature
				}).
				Return(newCommitStatusResponse(peer.TxValidationCode_VALID), nil).
				Times(1)

			gateway, contract := newContractWithNoSign(t, WithGatewayClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := gateway.NewSignedProposal(proposalBytes, expected)
			require.NoError(t, err)

			unsignedTransaction, err := signedProposal.Endorse()
			require.NoError(t, err)

			transactionBytes, err := unsignedTransaction.Bytes()
			require.NoError(t, err)

			signedTransaction, err := gateway.NewSignedTransaction(transactionBytes, expected)
			require.NoError(t, err)

			unsignedCommit, err := signedTransaction.Submit()
			require.NoError(t, err)

			commitBytes, err := unsignedCommit.Bytes()
			require.NoError(t, err)

			signedCommit, err := gateway.NewSignedCommit(commitBytes, expected)
			require.NoError(t, err)

			_, err = signedCommit.Status()
			require.NoError(t, err)

			require.EqualValues(t, expected, actual)
		})
	})

	t.Run("Serialization", func(t *testing.T) {
		t.Run("Proposal keeps same digest", func(t *testing.T) {
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			gateway, contract := newContractWithNoSign(t, WithGatewayClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := gateway.NewSignedProposal(proposalBytes, []byte("signature"))
			require.NoError(t, err)

			expected := unsignedProposal.Digest()
			actual := signedProposal.Digest()

			require.EqualValues(t, expected, actual)
		})

		t.Run("Proposal keeps same transaction ID", func(t *testing.T) {
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			gateway, contract := newContractWithNoSign(t, WithGatewayClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := gateway.NewSignedProposal(proposalBytes, []byte("signature"))
			require.NoError(t, err)

			expected := unsignedProposal.TransactionID()
			actual := signedProposal.TransactionID()

			require.EqualValues(t, expected, actual)
		})

		t.Run("Transaction keeps same digest", func(t *testing.T) {
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(AssertNewEndorseResponse(t, "result", "network"), nil).
				Times(1)

			gateway, contract := newContractWithNoSign(t, WithGatewayClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := gateway.NewSignedProposal(proposalBytes, []byte("signature"))
			require.NoError(t, err)

			unsignedTransaction, err := signedProposal.Endorse()
			require.NoError(t, err)

			transactionBytes, err := unsignedTransaction.Bytes()
			require.NoError(t, err)

			signedTransaction, err := gateway.NewSignedTransaction(transactionBytes, []byte("signature"))
			require.NoError(t, err)

			expected := unsignedTransaction.Digest()
			actual := signedTransaction.Digest()

			require.EqualValues(t, expected, actual)
		})

		t.Run("Commit keeps same digest", func(t *testing.T) {
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(AssertNewEndorseResponse(t, "result", "network"), nil).
				Times(1)
			mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				Times(1)

			gateway, contract := newContractWithNoSign(t, WithGatewayClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := gateway.NewSignedProposal(proposalBytes, []byte("signature"))
			require.NoError(t, err)

			unsignedTransaction, err := signedProposal.Endorse()
			require.NoError(t, err)

			transactionBytes, err := unsignedTransaction.Bytes()
			require.NoError(t, err)

			signedTransaction, err := gateway.NewSignedTransaction(transactionBytes, []byte("signature"))
			require.NoError(t, err)

			unsignedCommit, err := signedTransaction.Submit()
			require.NoError(t, err)

			commitBytes, err := unsignedCommit.Bytes()
			require.NoError(t, err)

			signedCommit, err := gateway.NewSignedCommit(commitBytes, []byte("signature"))
			require.NoError(t, err)

			expected := unsignedCommit.Digest()
			actual := signedCommit.Digest()

			require.EqualValues(t, expected, actual)
		})
	})
}
