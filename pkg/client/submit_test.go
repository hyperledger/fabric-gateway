/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"testing"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func AssertMarshal(t *testing.T, message protoreflect.ProtoMessage, msgAndArgs ...interface{}) []byte {
	bytes, err := proto.Marshal(message)
	require.NoError(t, err, msgAndArgs...)
	return bytes
}

func AssertNewEndorseResponse(t *testing.T, result string, channelName string) *gateway.EndorseResponse {
	return &gateway.EndorseResponse{
		PreparedTransaction: &common.Envelope{
			Payload: AssertMarshal(t, &common.Payload{
				Header: &common.Header{
					ChannelHeader: AssertMarshal(t, &common.ChannelHeader{
						ChannelId: channelName,
					}),
				},
				Data: AssertMarshal(t, &peer.Transaction{
					Actions: []*peer.TransactionAction{
						{
							Payload: AssertMarshal(t, &peer.ChaincodeActionPayload{
								Action: &peer.ChaincodeEndorsedAction{
									ProposalResponsePayload: AssertMarshal(t, &peer.ProposalResponsePayload{
										Extension: AssertMarshal(t, &peer.ChaincodeAction{
											Response: &peer.Response{
												Payload: []byte(result),
											},
										}),
									}),
								},
							}),
						},
					},
				}),
			}),
		},
	}
}

func TestSubmitTransaction(t *testing.T) {
	newCommitStatusResponse := func(status peer.TxValidationCode, blockNumber uint64) *gateway.CommitStatusResponse {
		return &gateway.CommitStatusResponse{
			Result:      status,
			BlockNumber: blockNumber,
		}
	}

	t.Run("Returns endorse error", func(t *testing.T) {
		expected := NewStatusError(t, codes.Aborted, "ENDORSE_ERROR")
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(nil, expected)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))
		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")

		_, err = proposal.Endorse()

		require.Equal(t, status.Code(expected), status.Code(err), "status code")
		var actual *EndorseError
		require.ErrorAs(t, err, &actual, "error type: %T", err)
		require.Equal(t, proposal.TransactionID(), actual.TransactionID, "transaction ID")
		require.ErrorContains(t, err, expected.Error(), "message")
	})

	t.Run("Returns submit error", func(t *testing.T) {
		expected := NewStatusError(t, codes.Aborted, "SUBMIT_ERROR")
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, expected)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))
		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")
		transaction, err := proposal.Endorse()
		require.NoError(t, err, "Endorse")

		_, err = transaction.Submit()

		require.Equal(t, status.Code(expected), status.Code(err), "status code")
		var actual *SubmitError
		require.ErrorAs(t, err, &actual, "error type: %T", err)
		require.Equal(t, proposal.TransactionID(), actual.TransactionID, "transaction ID")
		require.ErrorContains(t, err, expected.Error(), "message")
	})

	t.Run("Returns commit status error", func(t *testing.T) {
		expected := NewStatusError(t, codes.Aborted, "COMMIT_ERROR")
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(nil, expected)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))
		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")
		transaction, err := proposal.Endorse()
		require.NoError(t, err, "Endorse")
		commit, err := transaction.Submit()
		require.NoError(t, err, "Submit")

		_, err = commit.Status()

		require.Equal(t, status.Code(expected), status.Code(err), "status code")
		var actual *CommitStatusError
		require.ErrorAs(t, err, &actual, "error type: %T", err)
		require.Equal(t, proposal.TransactionID(), actual.TransactionID, "transaction ID")
		require.ErrorContains(t, err, expected.Error(), "message")
	})

	for name, testCase := range map[string]struct {
		run func(t *testing.T, contract *Contract) ([]byte, error)
	}{
		"SubmitTransaction returns result for committed transaction": {
			run: func(t *testing.T, contract *Contract) ([]byte, error) {
				return contract.SubmitTransaction("transaction")
			},
		},
		"SubmitWithContext returns result for committed transaction": {
			run: func(t *testing.T, contract *Contract) ([]byte, error) {
				return contract.SubmitWithContext(context.Background(), "transactionName")
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			expected := []byte("TRANSACTION_RESULT")
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
			mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
				Return(nil, nil)
			mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
				Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

			contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

			actual, err := testCase.run(t, contract)
			require.NoError(t, err)

			require.Equal(t, expected, actual)
		})
	}

	for testName, testCase := range map[string]struct {
		run func(t *testing.T, contract *Contract) ([]byte, error)
	}{
		"SubmitTransaction returns commit error for invalid commit status": {
			run: func(t *testing.T, contract *Contract) ([]byte, error) {
				return contract.SubmitTransaction("transaction")
			},
		},
		"SubmitWithContext returns commit error for invalid commit status": {
			run: func(t *testing.T, contract *Contract) ([]byte, error) {
				return contract.SubmitWithContext(context.Background(), "transactionName")
			},
		},
	} {
		t.Run(testName, func(t *testing.T) {
			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
			mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
				Return(nil, nil)
			mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
				Return(newCommitStatusResponse(peer.TxValidationCode_MVCC_READ_CONFLICT, 1), nil)

			contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))
			_, err := testCase.run(t, contract)

			var actual *CommitError
			require.ErrorAs(t, err, &actual, "error type: %T", err)
			require.NotEmpty(t, actual.TransactionID, "transaction ID")
			require.Equal(t, peer.TxValidationCode_MVCC_READ_CONFLICT, actual.Code, "validation code")
		})
	}

	t.Run("Includes channel name in proposal", func(t *testing.T) {
		var actual string
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				actual = test.AssertUnmarshalChannelheader(t, in.ProposedTransaction).ChannelId
			}).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		expected := contract.channelName
		require.Equal(t, expected, actual)
	})

	t.Run("Includes chaincode name in proposal", func(t *testing.T) {
		var actual string
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				actual = test.AssertUnmarshalInvocationSpec(t, in.ProposedTransaction).ChaincodeSpec.ChaincodeId.Name
			}).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		expected := contract.chaincodeName
		require.Equal(t, expected, actual)
	})

	t.Run("Includes transaction name in proposal for default contract", func(t *testing.T) {
		var args [][]byte
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				args = test.AssertUnmarshalInvocationSpec(t, in.ProposedTransaction).ChaincodeSpec.Input.Args
			}).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		expected := "TRANSACTION_NAME"
		_, err := contract.SubmitTransaction(expected)
		require.NoError(t, err)

		actual := string(args[0])
		require.Equal(t, expected, actual)
	})

	t.Run("Includes transaction name in proposal for named contract", func(t *testing.T) {
		var args [][]byte
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				args = test.AssertUnmarshalInvocationSpec(t, in.ProposedTransaction).ChaincodeSpec.Input.Args
			}).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

		contract := AssertNewTestContractWithName(t, "chaincode", "CONTRACT_NAME", WithGatewayClient(mockClient))

		_, err := contract.SubmitTransaction("TRANSACTION_NAME")
		require.NoError(t, err)

		actual := string(args[0])
		expected := "CONTRACT_NAME:TRANSACTION_NAME"
		require.Equal(t, expected, actual)
	})

	t.Run("Includes arguments in proposal", func(t *testing.T) {
		var args [][]byte
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				args = test.AssertUnmarshalInvocationSpec(t, in.ProposedTransaction).ChaincodeSpec.Input.Args
			}).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		expected := []string{"one", "two", "three"}
		_, err := contract.SubmitTransaction("transaction", expected...)
		require.NoError(t, err)

		actual := bytesAsStrings(args[1:])
		require.EqualValues(t, expected, actual)
	})

	t.Run("Includes channel name in proposed transaction", func(t *testing.T) {
		var actual string
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				actual = in.ChannelId
			}).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		expected := contract.channelName
		require.Equal(t, expected, actual)
	})

	t.Run("Includes transaction ID in proposed transaction", func(t *testing.T) {
		var actual string

		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				actual = test.AssertUnmarshalChannelheader(t, in.ProposedTransaction).TxId
			}).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")
		_, err = proposal.Endorse()
		require.NoError(t, err, "Endorse")

		require.Equal(t, proposal.TransactionID(), actual)
	})

	t.Run("Includes transaction ID in endorse request", func(t *testing.T) {
		var actual string

		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				actual = in.TransactionId
			}).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")
		_, err = proposal.Endorse()
		require.NoError(t, err, "Endorse")

		require.Equal(t, proposal.TransactionID(), actual)
	})

	t.Run("Includes channel name in commit status request", func(t *testing.T) {
		var actual string
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SignedCommitStatusRequest, _ ...grpc.CallOption) {
				request := &gateway.CommitStatusRequest{}
				test.AssertUnmarshal(t, in.Request, request)
				actual = request.ChannelId
			}).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		expected := contract.channelName
		require.Equal(t, expected, actual)
	})

	t.Run("Includes transaction ID in commit status request", func(t *testing.T) {
		var actual string
		var expected string
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
				expected = test.AssertUnmarshalChannelheader(t, in.ProposedTransaction).TxId
			}).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SignedCommitStatusRequest, _ ...grpc.CallOption) {
				request := &gateway.CommitStatusRequest{}
				test.AssertUnmarshal(t, in.Request, request)
				actual = request.TransactionId
			}).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		require.Equal(t, expected, actual)
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
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient), WithSign(sign))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		require.EqualValues(t, expected, actual)
	})

	t.Run("Uses signer for submit", func(t *testing.T) {
		var actual []byte
		expected := []byte("MY_SIGNATURE")
		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SubmitRequest, _ ...grpc.CallOption) {
				actual = in.PreparedTransaction.Signature
			}).
			Return(nil, nil).
			Times(1)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient), WithSign(sign))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		require.EqualValues(t, expected, actual)
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
				transient := test.AssertUnmarshalProposalPayload(t, in.ProposedTransaction).TransientMap
				actualPrice = transient["price"]
			}).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil).
			Times(1)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		privateData := map[string][]byte{
			"price": []byte("3000"),
		}

		_, err := contract.Submit("transaction", WithTransient(privateData), WithEndorsingOrganizations("MY_ORG"))
		require.NoError(t, err)

		require.EqualValues(t, expectedOrgs, actualOrgs)
		require.EqualValues(t, expectedPrice, actualPrice)
	})

	t.Run("Uses signer for commit status", func(t *testing.T) {
		var actual []byte
		expected := []byte("MY_SIGNATURE")
		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SignedCommitStatusRequest, _ ...grpc.CallOption) {
				actual = in.Signature
			}).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient), WithSign(sign))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		require.EqualValues(t, expected, actual)
	})

	t.Run("Uses hash", func(t *testing.T) {
		var actual [][]byte
		digest := []byte("MY_DIGEST")
		sign := func(digest []byte) ([]byte, error) {
			actual = append(actual, digest)
			return digest, nil
		}
		hash := func(message []byte) []byte {
			return digest
		}
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient), WithSign(sign), WithHash(hash))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		expected := [][]byte{digest, digest, digest}
		require.EqualValues(t, expected, actual)
	})

	t.Run("Commit returns transaction status", func(t *testing.T) {
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_MVCC_READ_CONFLICT, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		_, commit, err := contract.SubmitAsync("transaction")
		require.NoError(t, err)

		status, err := commit.Status()
		require.NoError(t, err)

		require.Equal(t, peer.TxValidationCode_MVCC_READ_CONFLICT, status.Code)
	})

	t.Run("Commit returns successful for successful transaction", func(t *testing.T) {
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		_, commit, err := contract.SubmitAsync("transaction")
		require.NoError(t, err, "submit")

		status, err := commit.Status()
		require.NoError(t, err, "commit status")

		require.True(t, status.Successful)
	})

	t.Run("Commit returns unsuccessful for failed transaction", func(t *testing.T) {
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_MVCC_READ_CONFLICT, 1), nil)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		_, commit, err := contract.SubmitAsync("transaction")
		require.NoError(t, err, "submit")

		status, err := commit.Status()
		require.NoError(t, err, "commit status")

		require.False(t, status.Successful)
	})

	t.Run("Commit returns block number", func(t *testing.T) {
		expectedBlockNumber := uint64(101)
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_MVCC_READ_CONFLICT, expectedBlockNumber), nil)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		_, commit, err := contract.SubmitAsync("transaction")
		require.NoError(t, err, "submit")

		status, err := commit.Status()
		require.NoError(t, err, "commit status")

		require.Equal(t, expectedBlockNumber, status.BlockNumber)
	})

	t.Run("Uses default context for endorse", func(t *testing.T) {
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, _ *gateway.EndorseRequest, _ ...grpc.CallOption) (*gateway.EndorseResponse, error) {
				select {
				case <-time.After(1 * time.Second):
					return AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil
				case <-ctx.Done(): // Zero timeout context should cancel immediately, selecting this case
					return nil, ctx.Err()
				}
			})
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil).
			AnyTimes()
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil).
			AnyTimes()

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient), WithEndorseTimeout(0))

		_, err := contract.Submit("transaction")

		require.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("Uses default context for submit", func(t *testing.T) {
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, _ *gateway.SubmitRequest, _ ...grpc.CallOption) (*gateway.SubmitResponse, error) {
				select {
				case <-time.After(1 * time.Second):
					return nil, nil
				case <-ctx.Done(): // Zero timeout context should cancel immediately, selecting this case
					return nil, ctx.Err()
				}
			})
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil).
			AnyTimes()

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient), WithSubmitTimeout(0))

		_, err := contract.Submit("transaction")

		require.ErrorIs(t, err, context.DeadlineExceeded)
	})

	for name, testCase := range map[string]struct {
		run func(*testing.T, context.Context, *Contract)
	}{
		"SubmitWithContext uses specified context": {
			run: func(t *testing.T, ctx context.Context, contract *Contract) {
				_, err := contract.SubmitWithContext(ctx, "transaction")
				require.NoError(t, err, "SubmitWithContext")
			},
		},
		"SubmitAsyncWithContext uses specified context": {
			run: func(t *testing.T, ctx context.Context, contract *Contract) {
				_, commit, err := contract.SubmitAsyncWithContext(ctx, "transaction")
				require.NoError(t, err, "SubmitAsyncWithContext")

				_, err = commit.StatusWithContext(ctx)
				require.NoError(t, err, "StatusWithContext")
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			var endorseCtxErr error
			var submitCtxErr error
			var commitStatusCtxErr error

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			mockClient := NewMockGatewayClient(gomock.NewController(t))
			mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, _ *gateway.EndorseRequest, _ ...grpc.CallOption) {
					endorseCtxErr = ctx.Err()
				}).
				Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil).
				Times(1)
			mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, _ *gateway.SubmitRequest, _ ...grpc.CallOption) {
					submitCtxErr = ctx.Err()
				}).
				Return(nil, nil).
				Times(1)
			mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
				Do(func(ctx context.Context, _ *gateway.SignedCommitStatusRequest, _ ...grpc.CallOption) {
					commitStatusCtxErr = ctx.Err()
				}).
				Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil).
				Times(1)

			contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

			testCase.run(t, ctx, contract)

			require.ErrorIs(t, endorseCtxErr, context.Canceled, "endorse context")
			require.ErrorIs(t, submitCtxErr, context.Canceled, "submit context")
			require.ErrorIs(t, commitStatusCtxErr, context.Canceled, "commit status context")
		})
	}

	t.Run("Uses default context for commit status", func(t *testing.T) {
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil).
			DoAndReturn(func(ctx context.Context, _ *gateway.SignedCommitStatusRequest, _ ...grpc.CallOption) (*gateway.CommitStatusResponse, error) {
				select {
				case <-time.After(1 * time.Second):
					return nil, nil
				case <-ctx.Done(): // Zero timeout context should cancel immediately, selecting this case
					return nil, ctx.Err()
				}
			})

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient), WithCommitStatusTimeout(0))

		_, err := contract.Submit("transaction")

		require.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("Endorse uses specified gRPC call options", func(t *testing.T) {
		var actual []grpc.CallOption
		expected := grpc.WaitForReady(true)

		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, _ *gateway.EndorseRequest, opts ...grpc.CallOption) {
				actual = opts
			}).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")

		_, err = proposal.Endorse(expected)
		require.NoError(t, err, "Endorse")

		require.Contains(t, actual, expected, "CallOptions")
	})

	t.Run("Endorse uses specified gRPC call options with specified context", func(t *testing.T) {
		var actual []grpc.CallOption
		expected := grpc.WaitForReady(true)

		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, _ *gateway.EndorseRequest, opts ...grpc.CallOption) {
				actual = opts
			}).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		_, err = proposal.EndorseWithContext(ctx, expected)
		require.NoError(t, err, "Endorse")

		require.Contains(t, actual, expected, "CallOptions")
	})

	t.Run("Submit uses specified gRPC call options", func(t *testing.T) {
		var actual []grpc.CallOption
		expected := grpc.WaitForReady(true)

		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, _ *gateway.SubmitRequest, opts ...grpc.CallOption) {
				actual = opts
			}).
			Return(nil, nil).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")

		transaction, err := proposal.Endorse()
		require.NoError(t, err, "Endorse")

		_, err = transaction.Submit(expected)
		require.NoError(t, err, "Submit")

		require.Contains(t, actual, expected, "CallOptions")
	})

	t.Run("Submit uses specified gRPC call options with specified context", func(t *testing.T) {
		var actual []grpc.CallOption
		expected := grpc.WaitForReady(true)

		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, _ *gateway.SubmitRequest, opts ...grpc.CallOption) {
				actual = opts
			}).
			Return(nil, nil).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")

		transaction, err := proposal.Endorse()
		require.NoError(t, err, "Endorse")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		_, err = transaction.SubmitWithContext(ctx, expected)
		require.NoError(t, err, "Submit")

		require.Contains(t, actual, expected, "CallOptions")
	})

	t.Run("CommisStatus uses specified gRPC call options", func(t *testing.T) {
		var actual []grpc.CallOption
		expected := grpc.WaitForReady(true)

		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil).
			Do(func(ctx context.Context, _ *gateway.SignedCommitStatusRequest, opts ...grpc.CallOption) {
				actual = opts
			}).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")

		transaction, err := proposal.Endorse()
		require.NoError(t, err, "Endorse")

		commit, err := transaction.Submit(expected)
		require.NoError(t, err, "Submit")

		_, err = commit.Status(expected)
		require.NoError(t, err, "Status")

		require.Contains(t, actual, expected, "CallOptions")
	})

	t.Run("CommisStatus uses specified gRPC call options with specified context", func(t *testing.T) {
		var actual []grpc.CallOption
		expected := grpc.WaitForReady(true)

		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network"), nil)
		mockClient.EXPECT().Submit(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, nil)
		mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(newCommitStatusResponse(peer.TxValidationCode_VALID, 1), nil).
			Do(func(ctx context.Context, _ *gateway.SignedCommitStatusRequest, opts ...grpc.CallOption) {
				actual = opts
			}).
			Times(1)

		contract := AssertNewTestContract(t, "chaincode", WithGatewayClient(mockClient))

		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")

		transaction, err := proposal.Endorse()
		require.NoError(t, err, "Endorse")

		commit, err := transaction.Submit(expected)
		require.NoError(t, err, "Submit")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		_, err = commit.StatusWithContext(ctx, expected)
		require.NoError(t, err, "Status")

		require.Contains(t, actual, expected, "CallOptions")
	})
}
