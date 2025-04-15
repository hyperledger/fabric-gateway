// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"testing"

	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ReceiveAll[T any](channel <-chan T) []T {
	var results []T

	for {
		if value, ok := TryReceive(channel); !ok {
			return results
		} else {
			results = append(results, value)
		}
	}
}

func TryReceive[T any](channel <-chan T) (T, bool) {
	var result T
	select {
	case result = <-channel:
		return result, true
	default:
		return result, false
	}
}

func TestSubmitTransaction(t *testing.T) {
	defaultEndorseResponse := AssertNewEndorseResponse(t, "TRANSACTION_RESULT", "network")

	t.Run("Returns endorse error", func(t *testing.T) {
		expected := NewStatusError(t, codes.Aborted, "ENDORSE_ERROR")

		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithInvokeError(expected))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))
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

		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection, WithInvokeError(expected))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))
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

		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection, WithInvokeError(expected))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))
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
		run func(*testing.T, *Contract) ([]byte, error)
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

			mockConnection := NewMockClientConnInterface(t)
			ExpectEndorse(mockConnection, WithEndorseResponse(defaultEndorseResponse))
			ExpectSubmit(mockConnection)
			ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

			contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

			actual, err := testCase.run(t, contract)
			require.NoError(t, err)

			require.Equal(t, expected, actual)
		})
	}

	for testName, testCase := range map[string]struct {
		run func(*testing.T, *Contract) ([]byte, error)
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
			mockConnection := NewMockClientConnInterface(t)
			ExpectEndorse(mockConnection, WithEndorseResponse(defaultEndorseResponse))
			ExpectSubmit(mockConnection)
			ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_MVCC_READ_CONFLICT, 1))

			contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))
			_, err := testCase.run(t, contract)

			var actual *CommitError
			require.ErrorAs(t, err, &actual, "error type: %T", err)
			require.NotEmpty(t, actual.TransactionID, "transaction ID")
			require.Equal(t, peer.TxValidationCode_MVCC_READ_CONFLICT, actual.Code, "validation code")
		})
	}

	t.Run("Includes channel name in proposal", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EndorseRequest, 1)
		ExpectEndorse(mockConnection, CaptureInvokeRequest(requests), WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		actual := AssertUnmarshalChannelheader(t, (<-requests).ProposedTransaction).ChannelId
		expected := contract.channelName
		require.Equal(t, expected, actual)
	})

	t.Run("Includes chaincode name in proposal", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EndorseRequest, 1)
		ExpectEndorse(mockConnection, CaptureInvokeRequest(requests), WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		actual := AssertUnmarshalInvocationSpec(t, (<-requests).ProposedTransaction).ChaincodeSpec.ChaincodeId.Name
		expected := contract.chaincodeName
		require.Equal(t, expected, actual)
	})

	t.Run("Includes transaction name in proposal for default contract", func(t *testing.T) {
		expected := "TRANSACTION_NAME"

		mockConnection := NewMockClientConnInterface(t)

		requests := make(chan *gateway.EndorseRequest, 1)
		ExpectEndorse(mockConnection, CaptureInvokeRequest(requests), WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		_, err := contract.SubmitTransaction(expected)
		require.NoError(t, err)

		args := AssertUnmarshalInvocationSpec(t, (<-requests).ProposedTransaction).ChaincodeSpec.Input.Args
		actual := string(args[0])
		require.Equal(t, expected, actual)
	})

	t.Run("Includes transaction name in proposal for named contract", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EndorseRequest, 1)
		ExpectEndorse(mockConnection, CaptureInvokeRequest(requests), WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

		contract := AssertNewTestContractWithName(t, "chaincode", "CONTRACT_NAME", WithClientConnection(mockConnection))

		_, err := contract.SubmitTransaction("TRANSACTION_NAME")
		require.NoError(t, err)

		args := AssertUnmarshalInvocationSpec(t, (<-requests).ProposedTransaction).ChaincodeSpec.Input.Args
		actual := string(args[0])
		expected := "CONTRACT_NAME:TRANSACTION_NAME"
		require.Equal(t, expected, actual)
	})

	t.Run("Includes arguments in proposal", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EndorseRequest, 1)
		ExpectEndorse(mockConnection, CaptureInvokeRequest(requests), WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		expected := []string{"one", "two", "three"}
		_, err := contract.SubmitTransaction("transaction", expected...)
		require.NoError(t, err)

		args := AssertUnmarshalInvocationSpec(t, (<-requests).ProposedTransaction).ChaincodeSpec.Input.Args
		actual := bytesAsStrings(args[1:])
		require.Equal(t, expected, actual)
	})

	t.Run("Includes channel name in endorse request", func(t *testing.T) {
		expected := "CHANNEL_NAME"

		mockConnection := NewMockClientConnInterface(t)
		network := AssertNewTestNetwork(t, expected, WithClientConnection(mockConnection))

		requests := make(chan *gateway.EndorseRequest, 1)
		ExpectEndorse(mockConnection, CaptureInvokeRequest(requests), WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

		contract := network.GetContract("chaincode")

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		actual := (<-requests).ChannelId
		require.Equal(t, expected, actual)
	})

	t.Run("Includes transaction ID in proposal", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EndorseRequest, 1)
		ExpectEndorse(mockConnection, CaptureInvokeRequest(requests), WithEndorseResponse(defaultEndorseResponse))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")
		_, err = proposal.Endorse()
		require.NoError(t, err, "Endorse")

		actual := AssertUnmarshalChannelheader(t, (<-requests).ProposedTransaction).TxId
		require.Equal(t, proposal.TransactionID(), actual)
	})

	t.Run("Includes transaction ID in endorse request", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EndorseRequest, 1)
		ExpectEndorse(mockConnection, CaptureInvokeRequest(requests), WithEndorseResponse(defaultEndorseResponse))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")
		_, err = proposal.Endorse()
		require.NoError(t, err, "Endorse")

		actual := (<-requests).TransactionId
		require.Equal(t, proposal.TransactionID(), actual)
	})

	t.Run("Includes channel name in commit status request", func(t *testing.T) {
		expected := "CHANNEL_NAME"

		mockConnection := NewMockClientConnInterface(t)
		network := AssertNewTestNetwork(t, expected, WithClientConnection(mockConnection))
		endorseResponse := AssertNewEndorseResponse(t, "TRANSACTION_RESULT", expected)
		ExpectEndorse(mockConnection, WithEndorseResponse(endorseResponse))
		ExpectSubmit(mockConnection)
		requests := make(chan *gateway.SignedCommitStatusRequest, 1)
		ExpectCommitStatus(mockConnection, CaptureInvokeRequest(requests), WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

		contract := network.GetContract("chaincode")

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		request := &gateway.CommitStatusRequest{}
		AssertUnmarshal(t, (<-requests).Request, request)
		actual := request.ChannelId
		require.Equal(t, expected, actual)
	})

	t.Run("Includes transaction ID in commit status request", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		endorseRequests := make(chan *gateway.EndorseRequest, 1)
		ExpectEndorse(mockConnection, CaptureInvokeRequest(endorseRequests), WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		commitStatusRequests := make(chan *gateway.SignedCommitStatusRequest, 1)
		ExpectCommitStatus(mockConnection, CaptureInvokeRequest(commitStatusRequests), WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		expected := AssertUnmarshalChannelheader(t, (<-endorseRequests).ProposedTransaction).TxId
		request := &gateway.CommitStatusRequest{}
		AssertUnmarshal(t, (<-commitStatusRequests).Request, request)
		actual := request.TransactionId
		require.Equal(t, expected, actual)
	})

	t.Run("Uses signer for endorse", func(t *testing.T) {
		expected := []byte("MY_SIGNATURE")

		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EndorseRequest, 1)
		ExpectEndorse(mockConnection, CaptureInvokeRequest(requests), WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection), WithSign(sign))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		actual := (<-requests).ProposedTransaction.Signature
		require.Equal(t, expected, actual)
	})

	t.Run("Uses signer for submit", func(t *testing.T) {
		expected := []byte("MY_SIGNATURE")

		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithEndorseResponse(defaultEndorseResponse))
		requests := make(chan *gateway.SubmitRequest, 1)
		ExpectSubmit(mockConnection, CaptureInvokeRequest(requests))
		ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection), WithSign(sign))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		actual := (<-requests).PreparedTransaction.Signature
		require.Equal(t, expected, actual)
	})

	t.Run("Sends private data with submit", func(t *testing.T) {
		expectedOrg := "MY_ORG"
		expectedPrice := []byte("3000")

		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EndorseRequest, 1)
		ExpectEndorse(mockConnection, CaptureInvokeRequest(requests), WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		privateData := map[string][]byte{
			"price": expectedPrice,
		}
		_, err := contract.Submit("transaction", WithTransient(privateData), WithEndorsingOrganizations(expectedOrg))
		require.NoError(t, err)

		request := <-requests
		require.ElementsMatch(t, []string{expectedOrg}, request.EndorsingOrganizations)

		transient := AssertUnmarshalProposalPayload(t, request.ProposedTransaction).TransientMap
		require.Equal(t, expectedPrice, transient["price"])
	})

	t.Run("Uses signer for commit status", func(t *testing.T) {
		expected := []byte("MY_SIGNATURE")

		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		requests := make(chan *gateway.SignedCommitStatusRequest, 1)
		ExpectCommitStatus(mockConnection, CaptureInvokeRequest(requests), WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection), WithSign(sign))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		actual := (<-requests).Signature
		require.Equal(t, expected, actual)
	})

	t.Run("Uses hash", func(t *testing.T) {
		digests := make(chan []byte, 3)
		digest := []byte("MY_DIGEST")
		sign := func(digest []byte) ([]byte, error) {
			digests <- digest
			return digest, nil
		}
		hash := func(message []byte) []byte {
			return digest
		}
		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection), WithSign(sign), WithHash(hash))

		_, err := contract.SubmitTransaction("transaction")
		require.NoError(t, err)

		expected := [][]byte{digest, digest, digest}
		actual := ReceiveAll(digests)
		require.Equal(t, expected, actual)
	})

	t.Run("Commit returns transaction status", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_MVCC_READ_CONFLICT, 1))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		_, commit, err := contract.SubmitAsync("transaction")
		require.NoError(t, err)

		status, err := commit.Status()
		require.NoError(t, err)

		require.Equal(t, peer.TxValidationCode_MVCC_READ_CONFLICT, status.Code)
	})

	t.Run("Commit returns successful for successful transaction", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_VALID, 1))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		_, commit, err := contract.SubmitAsync("transaction")
		require.NoError(t, err, "submit")

		status, err := commit.Status()
		require.NoError(t, err, "commit status")

		require.True(t, status.Successful)
	})

	t.Run("Commit returns unsuccessful for failed transaction", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_MVCC_READ_CONFLICT, 1))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		_, commit, err := contract.SubmitAsync("transaction")
		require.NoError(t, err, "submit")

		status, err := commit.Status()
		require.NoError(t, err, "commit status")

		require.False(t, status.Successful)
	})

	t.Run("Commit returns block number", func(t *testing.T) {
		expectedBlockNumber := uint64(101)

		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection, WithCommitStatusResponse(peer.TxValidationCode_MVCC_READ_CONFLICT, expectedBlockNumber))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		_, commit, err := contract.SubmitAsync("transaction")
		require.NoError(t, err, "submit")

		status, err := commit.Status()
		require.NoError(t, err, "commit status")

		require.Equal(t, expectedBlockNumber, status.BlockNumber)
	})

	t.Run("Uses default context for endorse", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithInvokeContextErr(), WithEndorseResponse(defaultEndorseResponse))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection), WithEndorseTimeout(0))

		_, err := contract.Submit("transaction")

		require.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("Uses default context for submit", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection, WithInvokeContextErr())

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection), WithSubmitTimeout(0))

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
			endorseContexts := make(chan context.Context, 1)
			submitContexts := make(chan context.Context, 1)
			commitStatusContexts := make(chan context.Context, 1)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			mockConnection := NewMockClientConnInterface(t)
			ExpectEndorse(mockConnection,
				CaptureInvokeContext(endorseContexts),
				WithEndorseResponse(defaultEndorseResponse),
			)
			ExpectSubmit(mockConnection, CaptureInvokeContext(submitContexts))
			ExpectCommitStatus(mockConnection,
				CaptureInvokeContext(commitStatusContexts),
				WithCommitStatusResponse(peer.TxValidationCode_VALID, 1),
			)

			contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

			testCase.run(t, ctx, contract)

			require.ErrorIs(t, (<-endorseContexts).Err(), context.Canceled, "endorse context")
			require.ErrorIs(t, (<-submitContexts).Err(), context.Canceled, "submit context")
			require.ErrorIs(t, (<-commitStatusContexts).Err(), context.Canceled, "commit status context")
		})
	}

	t.Run("Uses default context for commit status", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection, WithInvokeContextErr())

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection), WithCommitStatusTimeout(0))

		_, err := contract.Submit("transaction")

		require.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("Endorse uses specified gRPC call options", func(t *testing.T) {
		callOptions := make(chan []grpc.CallOption, 1)
		expected := grpc.WaitForReady(true)

		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection,
			CaptureInvokeOptions(callOptions),
			WithEndorseResponse(defaultEndorseResponse),
		)

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")

		_, err = proposal.Endorse(expected)
		require.NoError(t, err, "Endorse")

		require.Contains(t, <-callOptions, expected, "CallOptions")
	})

	t.Run("Endorse uses specified gRPC call options with specified context", func(t *testing.T) {
		callOptions := make(chan []grpc.CallOption, 1)
		expected := grpc.WaitForReady(true)

		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection,
			CaptureInvokeOptions(callOptions),
			WithEndorseResponse(defaultEndorseResponse),
		)

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		_, err = proposal.EndorseWithContext(ctx, expected)
		require.NoError(t, err, "Endorse")

		require.Contains(t, <-callOptions, expected, "CallOptions")
	})

	t.Run("Submit uses specified gRPC call options", func(t *testing.T) {
		callOptions := make(chan []grpc.CallOption, 1)
		expected := grpc.WaitForReady(true)

		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection, CaptureInvokeOptions(callOptions))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")

		transaction, err := proposal.Endorse()
		require.NoError(t, err, "Endorse")

		_, err = transaction.Submit(expected)
		require.NoError(t, err, "Submit")

		require.Contains(t, <-callOptions, expected, "CallOptions")
	})

	t.Run("Submit uses specified gRPC call options with specified context", func(t *testing.T) {
		callOptions := make(chan []grpc.CallOption, 1)
		expected := grpc.WaitForReady(true)

		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection, CaptureInvokeOptions(callOptions))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")

		transaction, err := proposal.Endorse()
		require.NoError(t, err, "Endorse")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		_, err = transaction.SubmitWithContext(ctx, expected)
		require.NoError(t, err, "Submit")

		require.Contains(t, <-callOptions, expected, "CallOptions")
	})

	t.Run("CommisStatus uses specified gRPC call options", func(t *testing.T) {
		callOptions := make(chan []grpc.CallOption, 1)
		expected := grpc.WaitForReady(true)

		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection,
			CaptureInvokeOptions(callOptions),
			WithCommitStatusResponse(peer.TxValidationCode_VALID, 1),
		)

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")

		transaction, err := proposal.Endorse()
		require.NoError(t, err, "Endorse")

		commit, err := transaction.Submit(expected)
		require.NoError(t, err, "Submit")

		_, err = commit.Status(expected)
		require.NoError(t, err, "Status")

		require.Contains(t, <-callOptions, expected, "CallOptions")
	})

	t.Run("CommisStatus uses specified gRPC call options with specified context", func(t *testing.T) {
		callOptions := make(chan []grpc.CallOption, 1)
		expected := grpc.WaitForReady(true)

		mockConnection := NewMockClientConnInterface(t)
		ExpectEndorse(mockConnection, WithEndorseResponse(defaultEndorseResponse))
		ExpectSubmit(mockConnection)
		ExpectCommitStatus(mockConnection,
			CaptureInvokeOptions(callOptions),
			WithCommitStatusResponse(peer.TxValidationCode_VALID, 1),
		)

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

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

		require.Contains(t, <-callOptions, expected, "CallOptions")
	})
}
