/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"errors"
	"testing"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
)

func TestOfflineSign(t *testing.T) {
	var signature []byte

	newGatewayWithNoSign := func(t *testing.T, options ...ConnectOption) *Gateway {
		defaultOptions := []ConnectOption{
			WithDeliverClient(NewMockDeliverClient(gomock.NewController(t))),
		}
		options = append(defaultOptions, options...)
		gateway, err := Connect(TestCredentials.Identity(), options...)
		require.NoError(t, err)

		return gateway
	}

	newMockDeliverEvents := func(controller *gomock.Controller) *MockDeliver_DeliverClient {
		mockEvents := NewMockDeliver_DeliverClient(controller)

		mockEvents.EXPECT().Send(gomock.Any()).
			Do(func(in *common.Envelope) {
				signature = in.GetSignature()
			}).
			Return(nil).
			AnyTimes()
		mockEvents.EXPECT().Recv().
			Return(nil, errors.New("fake")).
			AnyTimes()

		return mockEvents
	}

	type Invocation struct {
		Invoke func() error
	}

	type Signable struct {
		Invocations map[string]Invocation
		OfflineSign func([]byte) *Signable
		State       interface{}
		Recreate    func() *Signable
	}

	var newSignableFromProposal func(t *testing.T, gateway *Gateway, proposal *Proposal) *Signable
	newSignableFromProposal = func(t *testing.T, gateway *Gateway, proposal *Proposal) *Signable {
		return &Signable{
			Invocations: map[string]Invocation{
				"Evaluate": {
					Invoke: func() error {
						_, err := proposal.Evaluate()
						return err
					},
				},
				"Endorse": {
					Invoke: func() error {
						_, err := proposal.Endorse()
						return err
					},
				},
			},
			OfflineSign: func(signature []byte) *Signable {
				bytes, err := proposal.Bytes()
				require.NoError(t, err, "Bytes")

				result, err := gateway.NewSignedProposal(bytes, signature)
				require.NoError(t, err, "NewSignedProposal")

				return newSignableFromProposal(t, gateway, result)
			},
			State: struct {
				Digest        []byte
				TransactionID string
				EndorsingOrgs []string
			}{
				Digest:        proposal.Digest(),
				TransactionID: proposal.TransactionID(),
				EndorsingOrgs: proposal.proposedTransaction.GetEndorsingOrganizations(),
			},
			Recreate: func() *Signable {
				signedBytes, err := proposal.Bytes()
				require.NoError(t, err, "NewSignedProposal")

				newProposal, err := gateway.NewProposal(signedBytes)
				require.NoError(t, err, "NewProposal")

				return newSignableFromProposal(t, gateway, newProposal)
			},
		}
	}

	var newSignableFromTransaction func(t *testing.T, gateway *Gateway, transaction *Transaction) *Signable
	newSignableFromTransaction = func(t *testing.T, gateway *Gateway, transaction *Transaction) *Signable {
		return &Signable{
			Invocations: map[string]Invocation{
				"Submit": {
					Invoke: func() error {
						_, err := transaction.Submit()
						return err
					},
				},
			},
			OfflineSign: func(signature []byte) *Signable {
				bytes, err := transaction.Bytes()
				require.NoError(t, err, "Bytes")

				result, err := gateway.NewSignedTransaction(bytes, signature)
				require.NoError(t, err, "NewSignedTransaction")

				return newSignableFromTransaction(t, gateway, result)
			},
			State: struct {
				Digest        []byte
				TransactionID string
			}{
				Digest:        transaction.Digest(),
				TransactionID: transaction.TransactionID(),
			},
			Recreate: func() *Signable {
				signedBytes, err := transaction.Bytes()
				require.NoError(t, err, "NewSignedTransactionBytes")

				newTransaction, err := gateway.NewTransaction(signedBytes)
				require.NoError(t, err, "NewTransaction")

				return newSignableFromTransaction(t, gateway, newTransaction)
			},
		}
	}

	var newSignableFromCommit func(t *testing.T, gateway *Gateway, commit *Commit) *Signable
	newSignableFromCommit = func(t *testing.T, gateway *Gateway, commit *Commit) *Signable {
		return &Signable{
			Invocations: map[string]Invocation{
				"Status": {
					Invoke: func() error {
						_, err := commit.Status()
						return err
					},
				},
			},
			OfflineSign: func(signature []byte) *Signable {
				bytes, err := commit.Bytes()
				require.NoError(t, err, "Bytes")

				result, err := gateway.NewSignedCommit(bytes, signature)
				require.NoError(t, err, "NewSignedCommit")

				return newSignableFromCommit(t, gateway, result)
			},
			State: struct {
				Digest        []byte
				TransactionID string
			}{
				Digest:        commit.Digest(),
				TransactionID: commit.TransactionID(),
			},
			Recreate: func() *Signable {
				signedBytes, err := commit.Bytes()
				require.NoError(t, err, "NewSignedCommitBytes")

				newCommit, err := gateway.NewCommit(signedBytes)
				require.NoError(t, err, "NewCommit")

				return newSignableFromCommit(t, gateway, newCommit)
			},
		}
	}

	var newSignableFromChaincodeEventsRequest func(t *testing.T, gateway *Gateway, request *ChaincodeEventsRequest) *Signable
	newSignableFromChaincodeEventsRequest = func(t *testing.T, gateway *Gateway, request *ChaincodeEventsRequest) *Signable {
		return &Signable{
			Invocations: map[string]Invocation{
				"Events": {
					Invoke: func() error {
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						_, err := request.Events(ctx)
						return err
					},
				},
			},
			OfflineSign: func(signature []byte) *Signable {
				bytes, err := request.Bytes()
				require.NoError(t, err, "Bytes")

				result, err := gateway.NewSignedChaincodeEventsRequest(bytes, signature)
				require.NoError(t, err, "NewSignedChaincodeEventsRequest")

				return newSignableFromChaincodeEventsRequest(t, gateway, result)
			},
			State: struct {
				Digest []byte
			}{
				Digest: request.Digest(),
			},
			Recreate: func() *Signable {
				signedBytes, err := request.Bytes()
				require.NoError(t, err, "NewSignedChaincodeEventsRequestBytes")

				newChaincodeRequest, err := gateway.NewChaincodeEventsRequest(signedBytes)
				require.NoError(t, err, "newChaincodeRequest")

				return newSignableFromChaincodeEventsRequest(t, gateway, newChaincodeRequest)
			},
		}
	}

	var newSignableFromBlockEventsRequest func(t *testing.T, gateway *Gateway, request *BlockEventsRequest) *Signable
	newSignableFromBlockEventsRequest = func(t *testing.T, gateway *Gateway, request *BlockEventsRequest) *Signable {
		return &Signable{
			Invocations: map[string]Invocation{
				"Events": {
					Invoke: func() error {
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						_, err := request.Events(ctx)
						return err
					},
				},
			},
			OfflineSign: func(signature []byte) *Signable {
				bytes, err := request.Bytes()
				require.NoError(t, err, "Bytes")

				result, err := gateway.NewSignedBlockEventsRequest(bytes, signature)
				require.NoError(t, err, "NewSignedBlockEventsRequest")

				return newSignableFromBlockEventsRequest(t, gateway, result)
			},
			State: struct {
				Digest []byte
			}{
				Digest: request.Digest(),
			},
			Recreate: func() *Signable {
				signedBytes, err1 := request.Bytes()
				require.NoError(t, err1, "NewSignedBlockEventsRequestBytes")

				newBlockRequest, err2 := gateway.NewBlockEventsRequest(signedBytes)
				require.NoError(t, err2, "newBlockRequest")

				return newSignableFromBlockEventsRequest(t, gateway, newBlockRequest)
			},
		}
	}

	var newSignableFromFilteredBlockEventsRequest func(t *testing.T, gateway *Gateway, request *FilteredBlockEventsRequest) *Signable
	newSignableFromFilteredBlockEventsRequest = func(t *testing.T, gateway *Gateway, request *FilteredBlockEventsRequest) *Signable {
		return &Signable{
			Invocations: map[string]Invocation{
				"Events": {
					Invoke: func() error {
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						_, err := request.Events(ctx)
						return err
					},
				},
			},
			OfflineSign: func(signature []byte) *Signable {
				bytes, err := request.Bytes()
				require.NoError(t, err, "Bytes")

				result, err := gateway.NewSignedFilteredBlockEventsRequest(bytes, signature)
				require.NoError(t, err, "NewSignedFilteredBlockEventsRequest")

				return newSignableFromFilteredBlockEventsRequest(t, gateway, result)
			},
			State: struct {
				Digest []byte
			}{
				Digest: request.Digest(),
			},
			Recreate: func() *Signable {
				signedBytes, err := request.Bytes()
				require.NoError(t, err, "NewSignedFilteredBlockEventsRequestBytes")

				newFilteredBlockRequest, err := gateway.NewFilteredBlockEventsRequest(signedBytes)
				require.NoError(t, err, "newRequest")

				return newSignableFromFilteredBlockEventsRequest(t, gateway, newFilteredBlockRequest)
			},
		}
	}

	var newSignableFromBlockAndPrivateDataEventsRequest func(t *testing.T, gateway *Gateway, request *BlockAndPrivateDataEventsRequest) *Signable
	newSignableFromBlockAndPrivateDataEventsRequest = func(t *testing.T, gateway *Gateway, request *BlockAndPrivateDataEventsRequest) *Signable {
		return &Signable{
			Invocations: map[string]Invocation{
				"Events": {
					Invoke: func() error {
						ctx, cancel := context.WithCancel(context.Background())
						defer cancel()

						_, err := request.Events(ctx)
						return err
					},
				},
			},
			OfflineSign: func(signature []byte) *Signable {
				bytes, err := request.Bytes()
				require.NoError(t, err, "Bytes")

				result, err := gateway.NewSignedBlockAndPrivateDataEventsRequest(bytes, signature)
				require.NoError(t, err, "NewSignedBlockEventsRequest")

				return newSignableFromBlockAndPrivateDataEventsRequest(t, gateway, result)
			},
			State: struct {
				Digest []byte
			}{
				Digest: request.Digest(),
			},
			Recreate: func() *Signable {
				signedBytes, err1 := request.Bytes()
				require.NoError(t, err1, "NewSignedBlockAndPrivateDataEventsRequestBytes")

				newBlockAndPrivateDataRequest, err2 := gateway.NewBlockAndPrivateDataEventsRequest(signedBytes)
				require.NoError(t, err2, "newBlockAndPrivateDataRequest")

				return newSignableFromBlockAndPrivateDataEventsRequest(t, gateway, newBlockAndPrivateDataRequest)
			},
		}
	}

	for testName, testCase := range map[string]struct {
		Create func(*testing.T) *Signable
	}{
		"Proposal": {
			Create: func(t *testing.T) *Signable {
				mockClient := NewMockGatewayClient(gomock.NewController(t))
				mockClient.EXPECT().Evaluate(gomock.Any(), gomock.Any()).
					Do(func(_ context.Context, in *gateway.EvaluateRequest, _ ...grpc.CallOption) {
						signature = in.GetProposedTransaction().GetSignature()
					}).
					Return(&gateway.EvaluateResponse{
						Result: &peer.Response{
							Payload: nil,
						},
					}, nil).
					AnyTimes()
				mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
					Do(func(_ context.Context, in *gateway.EndorseRequest, _ ...grpc.CallOption) {
						signature = in.GetProposedTransaction().GetSignature()
					}).
					Return(AssertNewEndorseResponse(t, "result", "network"), nil).
					AnyTimes()

				gateway := newGatewayWithNoSign(t, WithGatewayClient(mockClient))
				contract := gateway.GetNetwork("NETWORK").GetContract("CHAINCODE")

				proposal, err := contract.NewProposal("transaction")
				require.NoError(t, err)

				return newSignableFromProposal(t, gateway, proposal)
			},
		},
		"Transaction": {
			Create: func(t *testing.T) *Signable {
				mockClient := NewMockGatewayClient(gomock.NewController(t))
				mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
					Return(AssertNewEndorseResponse(t, "result", "network"), nil)
				mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
					Do(func(_ context.Context, in *gateway.SubmitRequest, _ ...grpc.CallOption) {
						signature = in.GetPreparedTransaction().GetSignature()
					}).
					Return(nil, nil).
					AnyTimes()

				gateway := newGatewayWithNoSign(t, WithGatewayClient(mockClient))
				contract := gateway.GetNetwork("NETWORK").GetContract("CHAINCODE")

				unsignedProposal, err := contract.NewProposal("transaction")
				require.NoError(t, err)

				proposalBytes, err := unsignedProposal.Bytes()
				require.NoError(t, err)

				signedProposal, err := gateway.NewSignedProposal(proposalBytes, []byte("SIGNATURE"))
				require.NoError(t, err)

				transaction, err := signedProposal.Endorse()
				require.NoError(t, err)

				return newSignableFromTransaction(t, gateway, transaction)
			},
		},
		"Commit": {
			Create: func(t *testing.T) *Signable {
				mockClient := NewMockGatewayClient(gomock.NewController(t))
				mockClient.EXPECT().Endorse(gomock.Any(), gomock.Any()).
					Return(AssertNewEndorseResponse(t, "result", "network"), nil)
				mockClient.EXPECT().Submit(gomock.Any(), gomock.Any()).
					Return(nil, nil).
					AnyTimes()
				mockClient.EXPECT().CommitStatus(gomock.Any(), gomock.Any()).
					Do(func(_ context.Context, in *gateway.SignedCommitStatusRequest, _ ...grpc.CallOption) {
						signature = in.GetSignature()
					}).
					Return(&gateway.CommitStatusResponse{
						Result: peer.TxValidationCode_VALID,
					}, nil).
					AnyTimes()

				gateway := newGatewayWithNoSign(t, WithGatewayClient(mockClient))
				contract := gateway.GetNetwork("NETWORK").GetContract("CHAINCODE")

				unsignedProposal, err := contract.NewProposal("transaction")
				require.NoError(t, err)

				proposalBytes, err := unsignedProposal.Bytes()
				require.NoError(t, err)

				signedProposal, err := gateway.NewSignedProposal(proposalBytes, []byte("SIGNATURE"))
				require.NoError(t, err)

				unsignedTransaction, err := signedProposal.Endorse()
				require.NoError(t, err)

				transactionBytes, err := unsignedTransaction.Bytes()
				require.NoError(t, err)

				signedTransaction, err := gateway.NewSignedTransaction(transactionBytes, []byte("SIGNATURE"))
				require.NoError(t, err)

				commit, err := signedTransaction.Submit()
				require.NoError(t, err)

				return newSignableFromCommit(t, gateway, commit)
			},
		},
		"Chaincode events": {
			Create: func(t *testing.T) *Signable {
				controller := gomock.NewController(t)
				mockClient := NewMockGatewayClient(controller)
				mockEvents := NewMockGateway_ChaincodeEventsClient(controller)

				mockClient.EXPECT().ChaincodeEvents(gomock.Any(), gomock.Any()).
					Do(func(_ context.Context, in *gateway.SignedChaincodeEventsRequest, _ ...grpc.CallOption) {
						signature = in.GetSignature()
					}).
					Return(mockEvents, nil).
					AnyTimes()

				mockEvents.EXPECT().Recv().
					Return(nil, errors.New("fake")).
					AnyTimes()

				gateway := newGatewayWithNoSign(t, WithGatewayClient(mockClient))
				network := gateway.GetNetwork("NETWORK")

				request, err := network.NewChaincodeEventsRequest("CHAINCODE")
				require.NoError(t, err)

				return newSignableFromChaincodeEventsRequest(t, gateway, request)
			},
		},
		"Block events": {
			Create: func(t *testing.T) *Signable {
				controller := gomock.NewController(t)
				mockClient := NewMockDeliverClient(controller)
				mockEvents := newMockDeliverEvents(controller)

				mockClient.EXPECT().Deliver(gomock.Any(), gomock.Any()).
					Return(mockEvents, nil).
					AnyTimes()

				gateway := newGatewayWithNoSign(t, WithGatewayClient(NewMockGatewayClient(controller)), WithDeliverClient(mockClient))
				network := gateway.GetNetwork("NETWORK")

				request, err := network.NewBlockEventsRequest()
				require.NoError(t, err)

				return newSignableFromBlockEventsRequest(t, gateway, request)
			},
		},
		"Filtered block events": {
			Create: func(t *testing.T) *Signable {
				controller := gomock.NewController(t)
				mockClient := NewMockDeliverClient(controller)
				mockEvents := newMockDeliverEvents(controller)

				mockClient.EXPECT().DeliverFiltered(gomock.Any(), gomock.Any()).
					Return(mockEvents, nil).
					AnyTimes()

				gateway := newGatewayWithNoSign(t, WithGatewayClient(NewMockGatewayClient(controller)), WithDeliverClient(mockClient))
				network := gateway.GetNetwork("NETWORK")

				request, err := network.NewFilteredBlockEventsRequest()
				require.NoError(t, err)

				return newSignableFromFilteredBlockEventsRequest(t, gateway, request)
			},
		},
		"Block and private data events": {
			Create: func(t *testing.T) *Signable {
				controller := gomock.NewController(t)
				mockClient := NewMockDeliverClient(controller)
				mockEvents := newMockDeliverEvents(controller)

				mockClient.EXPECT().DeliverWithPrivateData(gomock.Any(), gomock.Any()).
					Return(mockEvents, nil).
					AnyTimes()

				gateway := newGatewayWithNoSign(t, WithGatewayClient(NewMockGatewayClient(controller)), WithDeliverClient(mockClient))
				network := gateway.GetNetwork("NETWORK")

				request, err := network.NewBlockAndPrivateDataEventsRequest()
				require.NoError(t, err)

				return newSignableFromBlockAndPrivateDataEventsRequest(t, gateway, request)
			},
		},
	} {
		t.Run(testName, func(t *testing.T) {
			unsigned := testCase.Create(t)

			for invocationName, invocation := range unsigned.Invocations {
				t.Run(invocationName, func(t *testing.T) {
					t.Run("Returns error with no signer and no explicit signing", func(t *testing.T) {
						err := invocation.Invoke()
						require.Error(t, err)
					})

					t.Run("Uses off-line signature", func(t *testing.T) {
						signature = nil
						expected := []byte("SIGNATURE")

						signed := unsigned.OfflineSign(expected)
						err := signed.Invocations[invocationName].Invoke()
						require.NoError(t, err)

						require.EqualValues(t, expected, signature)
					})

					t.Run("retains signature", func(t *testing.T) {
						signature = nil
						expected := []byte("SIGNATURE")

						signed := unsigned.OfflineSign(expected)
						recreated := signed.Recreate()
						err := recreated.Invocations[invocationName].Invoke()
						require.NoError(t, err)

						require.EqualValues(t, expected, signature)

					})
				})
			}

			t.Run("Retains state after signing", func(t *testing.T) {
				signed := unsigned.OfflineSign([]byte("SIGNATURE"))
				require.EqualValues(t, unsigned.State, signed.State)
			})
		})
	}
}
