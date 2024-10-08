// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"testing"

	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestEvaluateTransaction(t *testing.T) {
	t.Run("Returns evaluate error", func(t *testing.T) {
		expected := NewStatusError(t, codes.Aborted, "EVALUATE_ERROR")

		mockConnection := NewMockClientConnInterface(t)
		ExpectEvaluate(mockConnection, WithInvokeError(expected))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))
		_, err := contract.EvaluateTransaction("transaction")

		require.ErrorIs(t, err, expected, "error type: %T", err)
		require.ErrorContains(t, err, expected.Error(), "message")
		require.Equal(t, status.Code(expected), status.Code(err), "status code")
	})

	for name, testCase := range map[string]struct {
		run func(*testing.T, *Contract) ([]byte, error)
	}{
		"EvaluateTransaction returns result": {
			run: func(t *testing.T, contract *Contract) ([]byte, error) {
				return contract.EvaluateTransaction("transaction")
			},
		},
		"EvaluateWithContext returns result": {
			run: func(t *testing.T, contract *Contract) ([]byte, error) {
				return contract.EvaluateWithContext(context.Background(), "transaction")
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			expected := []byte("TRANSACTION_RESULT")

			mockConnection := NewMockClientConnInterface(t)
			ExpectEvaluate(mockConnection, WithEvaluateResponse(expected))
			contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

			actual, err := testCase.run(t, contract)
			require.NoError(t, err)

			require.EqualValues(t, expected, actual)
		})
	}

	t.Run("Includes channel name in proposal", func(t *testing.T) {
		expected := "CHANNEL_NAME"

		mockConnection := NewMockClientConnInterface(t)
		network := AssertNewTestNetwork(t, expected, WithClientConnection(mockConnection))

		requests := make(chan *gateway.EvaluateRequest, 1)
		ExpectEvaluate(mockConnection, CaptureInvokeRequest(requests), WithEvaluateResponse(nil))
		contract := network.GetContract("chaincode")

		_, err := contract.EvaluateTransaction("transaction")
		require.NoError(t, err)

		actual := AssertUnmarshalChannelheader(t, (<-requests).ProposedTransaction).ChannelId
		require.Equal(t, expected, actual)
	})

	t.Run("Includes chaincode name in proposal", func(t *testing.T) {
		expected := "CHAINCODE_NAME"

		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EvaluateRequest, 1)
		ExpectEvaluate(mockConnection, CaptureInvokeRequest(requests), WithEvaluateResponse(nil))
		contract := AssertNewTestContract(t, expected, WithClientConnection(mockConnection))

		_, err := contract.EvaluateTransaction("transaction")
		require.NoError(t, err)

		actual := AssertUnmarshalInvocationSpec(t, (<-requests).ProposedTransaction).ChaincodeSpec.ChaincodeId.Name
		require.Equal(t, expected, actual)
	})

	t.Run("Includes transaction name in proposal for default smart contract", func(t *testing.T) {
		expected := "TRANSACTION_NAME"

		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EvaluateRequest, 1)
		ExpectEvaluate(mockConnection, CaptureInvokeRequest(requests), WithEvaluateResponse(nil))
		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		_, err := contract.EvaluateTransaction(expected)
		require.NoError(t, err)

		args := AssertUnmarshalInvocationSpec(t, (<-requests).ProposedTransaction).ChaincodeSpec.Input.Args
		actual := string(args[0])
		require.Equal(t, expected, actual, "got Args: %s", args)
	})

	t.Run("Includes transaction name in proposal for named smart contract", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EvaluateRequest, 1)
		ExpectEvaluate(mockConnection, CaptureInvokeRequest(requests), WithEvaluateResponse(nil))
		contract := AssertNewTestContractWithName(t, "chaincode", "CONTRACT_NAME", WithClientConnection(mockConnection))

		_, err := contract.EvaluateTransaction("TRANSACTION_NAME")
		require.NoError(t, err)

		args := AssertUnmarshalInvocationSpec(t, (<-requests).ProposedTransaction).ChaincodeSpec.Input.Args
		actual := string(args[0])
		expected := "CONTRACT_NAME:TRANSACTION_NAME"
		require.Equal(t, expected, actual, "got Args: %s", args)
	})

	t.Run("Includes arguments in proposal", func(t *testing.T) {
		expected := []string{"one", "two", "three"}

		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EvaluateRequest, 1)
		ExpectEvaluate(mockConnection, CaptureInvokeRequest(requests), WithEvaluateResponse(nil))
		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		_, err := contract.EvaluateTransaction("transaction", expected...)
		require.NoError(t, err)

		args := AssertUnmarshalInvocationSpec(t, (<-requests).ProposedTransaction).ChaincodeSpec.Input.Args
		actual := bytesAsStrings(args[1:])
		require.EqualValues(t, expected, actual, "got Args: %s", args)
	})

	t.Run("Includes channel name in proposed transaction", func(t *testing.T) {
		expected := "CHANNEL_NAME"

		mockConnection := NewMockClientConnInterface(t)
		network := AssertNewTestNetwork(t, expected, WithClientConnection(mockConnection))
		contract := network.GetContract("chaincode")

		requests := make(chan *gateway.EvaluateRequest, 1)
		ExpectEvaluate(mockConnection, CaptureInvokeRequest(requests), WithEvaluateResponse(nil))

		_, err := contract.EvaluateTransaction("transaction")
		require.NoError(t, err)

		actual := (<-requests).ChannelId
		require.Equal(t, contract.channelName, actual)
	})

	t.Run("Includes transaction ID in proposed transaction", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EvaluateRequest, 1)
		ExpectEvaluate(mockConnection, CaptureInvokeRequest(requests), WithEvaluateResponse(nil))
		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")
		_, err = proposal.Evaluate()
		require.NoError(t, err, "Evaluate")

		actual := AssertUnmarshalChannelheader(t, (<-requests).ProposedTransaction).TxId
		require.Equal(t, proposal.TransactionID(), actual)
	})

	t.Run("Includes transaction ID in evaluate request", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EvaluateRequest, 1)
		ExpectEvaluate(mockConnection, CaptureInvokeRequest(requests), WithEvaluateResponse(nil))
		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		proposal, err := contract.NewProposal("transaction")
		require.NoError(t, err, "NewProposal")
		_, err = proposal.Evaluate()
		require.NoError(t, err, "Evaluate")

		actual := (<-requests).TransactionId
		require.Equal(t, proposal.TransactionID(), actual)
	})

	t.Run("Uses sign", func(t *testing.T) {
		expected := []byte("MY_SIGNATURE")

		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EvaluateRequest, 1)
		ExpectEvaluate(mockConnection, CaptureInvokeRequest(requests), WithEvaluateResponse(nil))

		sign := func(digest []byte) ([]byte, error) {
			return expected, nil
		}
		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection), WithSign(sign))

		_, err := contract.EvaluateTransaction("transaction")
		require.NoError(t, err)

		actual := (<-requests).ProposedTransaction.Signature
		require.EqualValues(t, expected, actual)
	})

	t.Run("Uses hash", func(t *testing.T) {
		expected := []byte("MY_DIGEST")

		mockConnection := NewMockClientConnInterface(t)
		ExpectEvaluate(mockConnection, WithEvaluateResponse(nil))

		digests := make(chan []byte, 1)
		sign := func(digest []byte) ([]byte, error) {
			digests <- digest
			return digest, nil
		}
		hash := func(message []byte) []byte {
			return expected
		}
		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection), WithSign(sign), WithHash(hash))

		_, err := contract.EvaluateTransaction("transaction")
		require.NoError(t, err)

		actual := <-digests
		require.EqualValues(t, expected, actual)
	})

	t.Run("Sends private data with evaluate", func(t *testing.T) {
		expectedOrgs := []string{"MY_ORG"}
		expectedPrice := []byte("3000")

		mockConnection := NewMockClientConnInterface(t)
		requests := make(chan *gateway.EvaluateRequest, 1)
		ExpectEvaluate(mockConnection, CaptureInvokeRequest(requests), WithEvaluateResponse(nil))
		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

		privateData := map[string][]byte{
			"price": []byte("3000"),
		}
		_, err := contract.Evaluate("transaction", WithTransient(privateData), WithEndorsingOrganizations("MY_ORG"))
		require.NoError(t, err)

		request := <-requests
		actualOrgs := request.TargetOrganizations
		require.EqualValues(t, expectedOrgs, actualOrgs)

		transient := AssertUnmarshalProposalPayload(t, request.ProposedTransaction).TransientMap
		actualPrice := transient["price"]
		require.EqualValues(t, expectedPrice, actualPrice)
	})

	for name, testCase := range map[string]struct {
		run func(*testing.T, context.Context, *Contract) ([]byte, error)
	}{
		"Proposal uses specified context": {
			run: func(t *testing.T, ctx context.Context, contract *Contract) ([]byte, error) {
				proposal, err := contract.NewProposal("transaction")
				require.NoError(t, err, "NewProposal")
				return proposal.EvaluateWithContext(ctx)
			},
		},
		"Contract uses specified context": {
			run: func(t *testing.T, ctx context.Context, contract *Contract) ([]byte, error) {
				return contract.EvaluateWithContext(ctx, "transaction")
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			mockConnection := NewMockClientConnInterface(t)
			ExpectEvaluate(mockConnection, WithInvokeContextErr())
			contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

			_, err := testCase.run(t, ctx, contract)
			require.ErrorIs(t, err, context.Canceled)
		})
	}

	t.Run("Uses default context", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		ExpectEvaluate(mockConnection, WithInvokeContextErr(), WithEvaluateResponse(nil))

		contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection), WithEvaluateTimeout(0))

		_, err := contract.Evaluate("transaction")

		require.ErrorIs(t, err, context.DeadlineExceeded)
	})

	for testName, testCase := range map[string]struct {
		run func(*testing.T, *Contract, []grpc.CallOption) ([]byte, error)
	}{
		"Uses specified gRPC call options": {
			run: func(t *testing.T, contract *Contract, expected []grpc.CallOption) ([]byte, error) {
				proposal, err := contract.NewProposal("transaction")
				require.NoError(t, err, "NewProposal")
				return proposal.Evaluate(expected...)
			},
		},
		"Uses specified gRPC call options with specified context": {
			run: func(t *testing.T, contract *Contract, expected []grpc.CallOption) ([]byte, error) {
				proposal, err := contract.NewProposal("transaction")
				require.NoError(t, err, "NewProposal")
				return proposal.EvaluateWithContext(context.Background(), expected...)
			},
		},
	} {
		t.Run(testName, func(t *testing.T) {
			expected := grpc.WaitForReady(true)

			options := make(chan []grpc.CallOption, 1)
			mockConnection := NewMockClientConnInterface(t)
			ExpectEvaluate(mockConnection, CaptureInvokeOptions(options), WithEvaluateResponse(nil))
			contract := AssertNewTestContract(t, "chaincode", WithClientConnection(mockConnection))

			_, err := testCase.run(t, contract, []grpc.CallOption{expected})
			require.NoError(t, err, "Evaluate")

			actual := <-options
			require.Contains(t, actual, expected, "CallOptions")
		})
	}
}
