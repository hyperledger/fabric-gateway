/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/gateway"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestOfflineSign(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	evaluateResponse := gateway.EvaluateResponse{
		Result: &peer.Response{
			Payload: nil,
		},
	}

	newNetworkWithNoSign := func(t *testing.T, options ...ConnectOption) *Network {
		gateway, err := Connect(TestCredentials.identity, options...)
		require.NoError(t, err)

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
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
				Return(&evaluateResponse, nil).
				AnyTimes()

			contract := newContractWithNoSign(t, WithClient(mockClient))

			proposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			_, err = proposal.Evaluate(ctx)
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

			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := contract.NewSignedProposal(proposalBytes, expected)
			require.NoError(t, err)

			_, err = signedProposal.Evaluate(ctx)
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

			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction", WithEndorsingOrganizations("MY_ORG"))
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("SIGNATURE"))
			require.NoError(t, err)

			_, err = signedProposal.Evaluate(ctx)
			require.NoError(t, err)

			require.EqualValues(t, expected, actual)
		})
	})

	t.Run("Endorse", func(t *testing.T) {
		t.Run("Returns error with no signer and no explicit signing", func(t *testing.T) {
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(newEndorseResponse("result"), nil).
				AnyTimes()

			contract := newContractWithNoSign(t, WithClient(mockClient))

			proposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			_, err = proposal.Endorse(ctx)
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
				Return(newEndorseResponse("result"), nil).
				Times(1)

			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := contract.NewSignedProposal(proposalBytes, expected)
			require.NoError(t, err)

			_, err = signedProposal.Endorse(ctx)
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
				Return(newEndorseResponse("result"), nil).
				Times(1)

			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction", WithEndorsingOrganizations("MY_ORG"))
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("SIGNATURE"))
			require.NoError(t, err)

			_, err = signedProposal.Endorse(ctx)
			require.NoError(t, err)

			require.EqualValues(t, expected, actual)
		})
	})

	t.Run("Submit", func(t *testing.T) {
		t.Run("Returns error with no signer and no explicit signing", func(t *testing.T) {
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(newEndorseResponse("result"), nil).
				AnyTimes()
			mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				AnyTimes()

			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("signature"))
			require.NoError(t, err)

			transaction, err := signedProposal.Endorse(ctx)
			require.NoError(t, err)

			_, err = transaction.Submit(ctx)
			require.Error(t, err)
		})

		t.Run("Uses off-line signature", func(t *testing.T) {
			expected := []byte("SIGNATURE")
			var actual []byte
			mockClient := NewMockGatewayClient(gomock.NewController(t))
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
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := contract.NewSignedProposal(proposalBytes, expected)
			require.NoError(t, err)

			unsignedTransaction, err := signedProposal.Endorse(ctx)
			require.NoError(t, err)

			transactionBytes, err := unsignedTransaction.Bytes()
			require.NoError(t, err)

			signedTransaction, err := contract.NewSignedTransaction(transactionBytes, expected)
			require.NoError(t, err)

			_, err = signedTransaction.Submit(ctx)
			require.NoError(t, err)

			require.EqualValues(t, expected, actual)
		})
	})

	t.Run("Commit", func(t *testing.T) {
		t.Run("Returns error with no signer and no explicit signing", func(t *testing.T) {
			mockClient := NewMockGatewayClient(gomock.NewController(t))
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
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("signature"))
			require.NoError(t, err)

			unsignedTransaction, err := signedProposal.Endorse(ctx)
			require.NoError(t, err)

			transactionBytes, err := unsignedTransaction.Bytes()
			require.NoError(t, err)

			signedTransaction, err := contract.NewSignedTransaction(transactionBytes, []byte("signature"))
			require.NoError(t, err)

			commit, err := signedTransaction.Submit(ctx)
			require.NoError(t, err)

			_, err = commit.Status(ctx)
			require.Error(t, err)
		})

		t.Run("Uses off-line signature", func(t *testing.T) {
			expected := []byte("SIGNATURE")
			var actual []byte
			mockClient := NewMockGatewayClient(gomock.NewController(t))
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
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := contract.NewSignedProposal(proposalBytes, expected)
			require.NoError(t, err)

			unsignedTransaction, err := signedProposal.Endorse(ctx)
			require.NoError(t, err)

			transactionBytes, err := unsignedTransaction.Bytes()
			require.NoError(t, err)

			signedTransaction, err := contract.NewSignedTransaction(transactionBytes, expected)
			require.NoError(t, err)

			unsignedCommit, err := signedTransaction.Submit(ctx)
			require.NoError(t, err)

			commitBytes, err := unsignedCommit.Bytes()
			require.NoError(t, err)

			signedCommit, err := network.NewSignedCommit(commitBytes, expected)
			require.NoError(t, err)

			_, err = signedCommit.Status(ctx)
			require.NoError(t, err)

			require.EqualValues(t, expected, actual)
		})
	})

	t.Run("Serialization", func(t *testing.T) {
		t.Run("Proposal keeps same digest", func(t *testing.T) {
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("signature"))
			require.NoError(t, err)

			expected := unsignedProposal.Digest()
			actual := signedProposal.Digest()

			require.EqualValues(t, expected, actual)
		})

		t.Run("Proposal keeps same transaction ID", func(t *testing.T) {
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("signature"))
			require.NoError(t, err)

			expected := unsignedProposal.TransactionID()
			actual := signedProposal.TransactionID()

			require.EqualValues(t, expected, actual)
		})

		t.Run("Transaction keeps same digest", func(t *testing.T) {
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(newEndorseResponse("result"), nil).
				Times(1)

			contract := newContractWithNoSign(t, WithClient(mockClient))

			unsignedProposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("signature"))
			require.NoError(t, err)

			unsignedTransaction, err := signedProposal.Endorse(ctx)
			require.NoError(t, err)

			transactionBytes, err := unsignedTransaction.Bytes()
			require.NoError(t, err)

			signedTransaction, err := contract.NewSignedTransaction(transactionBytes, []byte("signature"))
			require.NoError(t, err)

			expected := unsignedTransaction.Digest()
			actual := signedTransaction.Digest()

			require.EqualValues(t, expected, actual)
		})

		t.Run("Commit keeps same digest", func(t *testing.T) {
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(newEndorseResponse("result"), nil).
				Times(1)
			mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
				Return(nil, nil).
				Times(1)

			network := newNetworkWithNoSign(t, WithClient(mockClient))
			contract := network.GetContract("chaincode")

			unsignedProposal, err := contract.NewProposal("transaction")
			require.NoError(t, err)

			proposalBytes, err := unsignedProposal.Bytes()
			require.NoError(t, err)

			signedProposal, err := contract.NewSignedProposal(proposalBytes, []byte("signature"))
			require.NoError(t, err)

			unsignedTransaction, err := signedProposal.Endorse(ctx)
			require.NoError(t, err)

			transactionBytes, err := unsignedTransaction.Bytes()
			require.NoError(t, err)

			signedTransaction, err := contract.NewSignedTransaction(transactionBytes, []byte("signature"))
			require.NoError(t, err)

			unsignedCommit, err := signedTransaction.Submit(ctx)
			require.NoError(t, err)

			commitBytes, err := unsignedCommit.Bytes()
			require.NoError(t, err)

			signedCommit, err := network.NewSignedCommit(commitBytes, []byte("signature"))
			require.NoError(t, err)

			expected := unsignedCommit.Digest()
			actual := signedCommit.Digest()

			require.EqualValues(t, expected, actual)
		})
	})
}
