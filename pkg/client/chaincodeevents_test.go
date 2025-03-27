// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func TestChaincodeEvents(t *testing.T) {
	newChaincodeEventsResponse := func(events []*ChaincodeEvent) *gateway.ChaincodeEventsResponse {
		blockNumber := uint64(0)
		var peerEvents []*peer.ChaincodeEvent

		for _, event := range events {
			blockNumber = event.BlockNumber
			peerEvents = append(peerEvents, &peer.ChaincodeEvent{
				ChaincodeId: event.ChaincodeName,
				TxId:        event.TransactionID,
				EventName:   event.EventName,
				Payload:     event.Payload,
			})
		}

		return &gateway.ChaincodeEventsResponse{
			BlockNumber: blockNumber,
			Events:      peerEvents,
		}
	}

	t.Run("Returns connect error", func(t *testing.T) {
		expected := NewStatusError(t, codes.Aborted, "CHAINCODE_EVENTS_ERROR")

		mockConnection := NewMockClientConnInterface(t)
		ExpectChaincodeEvents(mockConnection, WithNewStreamError(expected))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithClientConnection(mockConnection))
		_, err := network.ChaincodeEvents(ctx, "CHAINCODE")

		require.Equal(t, status.Code(expected), status.Code(err), "status code")
		require.ErrorIs(t, err, expected, "error type: %T", err)
		require.ErrorContains(t, err, expected.Error(), "message")
	})

	for testName, testCase := range map[string]struct {
		options  []ChaincodeEventsOption
		expected *gateway.ChaincodeEventsRequest
	}{
		"Sends valid request with default start position": {
			options: nil,
			expected: &gateway.ChaincodeEventsRequest{
				StartPosition: &orderer.SeekPosition{
					Type: &orderer.SeekPosition_NextCommit{
						NextCommit: &orderer.SeekNextCommit{},
					},
				},
			},
		},
		"Sends valid request with specified start block number": {
			options: []ChaincodeEventsOption{
				WithStartBlock(418),
			},
			expected: &gateway.ChaincodeEventsRequest{
				StartPosition: &orderer.SeekPosition{
					Type: &orderer.SeekPosition_Specified{
						Specified: &orderer.SeekSpecified{
							Number: 418,
						},
					},
				},
			},
		},
		"Sends valid request with specified start block number and fresh checkpointer": {
			options: []ChaincodeEventsOption{
				WithStartBlock(418),
				WithCheckpoint(new(InMemoryCheckpointer)),
			},
			expected: &gateway.ChaincodeEventsRequest{
				StartPosition: &orderer.SeekPosition{
					Type: &orderer.SeekPosition_Specified{
						Specified: &orderer.SeekSpecified{
							Number: 418,
						},
					},
				},
			},
		},
		"Sends valid request with specified start block and checkpoint block": {
			options: func() []ChaincodeEventsOption {
				checkpointer := new(InMemoryCheckpointer)
				checkpointer.CheckpointBlock(500)
				return []ChaincodeEventsOption{
					WithStartBlock(418),
					WithCheckpoint(checkpointer),
				}
			}(),
			expected: &gateway.ChaincodeEventsRequest{
				StartPosition: &orderer.SeekPosition{
					Type: &orderer.SeekPosition_Specified{
						Specified: &orderer.SeekSpecified{
							Number: 501,
						},
					},
				},
			},
		},
		"Sends valid request with specified start block and checkpoint transaction ID": {
			options: func() []ChaincodeEventsOption {
				checkpointer := new(InMemoryCheckpointer)
				checkpointer.CheckpointTransaction(500, "txn1")
				return []ChaincodeEventsOption{
					WithStartBlock(418),
					WithCheckpoint(checkpointer),
				}
			}(),
			expected: &gateway.ChaincodeEventsRequest{
				StartPosition: &orderer.SeekPosition{
					Type: &orderer.SeekPosition_Specified{
						Specified: &orderer.SeekSpecified{
							Number: 500,
						},
					},
				},
				AfterTransactionId: "txn1",
			},
		},
		"Sends valid request with no start block and fresh checkpointer": {
			options: []ChaincodeEventsOption{
				WithCheckpoint(new(InMemoryCheckpointer)),
			},
			expected: &gateway.ChaincodeEventsRequest{
				StartPosition: &orderer.SeekPosition{
					Type: &orderer.SeekPosition_NextCommit{
						NextCommit: &orderer.SeekNextCommit{},
					},
				},
			},
		},
		"Sends valid request with no start block and checkpoint transaction ID": {
			options: func() []ChaincodeEventsOption {
				checkpointer := new(InMemoryCheckpointer)
				checkpointer.CheckpointTransaction(500, "txn1")
				return []ChaincodeEventsOption{
					WithCheckpoint(checkpointer),
				}
			}(),
			expected: &gateway.ChaincodeEventsRequest{
				StartPosition: &orderer.SeekPosition{
					Type: &orderer.SeekPosition_Specified{
						Specified: &orderer.SeekSpecified{
							Number: 500,
						},
					},
				},
				AfterTransactionId: "txn1",
			},
		},
		"Sends valid request with with start block and checkpoint chaincode event": {
			options: func() []ChaincodeEventsOption {
				checkpointer := new(InMemoryCheckpointer)
				event := &ChaincodeEvent{
					BlockNumber:   1,
					TransactionID: "TRANSACTION_1",
				}
				checkpointer.CheckpointChaincodeEvent(event)
				return []ChaincodeEventsOption{
					WithCheckpoint(checkpointer),
				}
			}(),
			expected: &gateway.ChaincodeEventsRequest{
				StartPosition: &orderer.SeekPosition{
					Type: &orderer.SeekPosition_Specified{
						Specified: &orderer.SeekSpecified{
							Number: 1,
						},
					},
				},
				AfterTransactionId: "TRANSACTION_1",
			},
		},
	} {
		t.Run(testName, func(t *testing.T) {
			mockConnection := NewMockClientConnInterface(t)
			mockStream := NewMockClientStream(t)
			ExpectChaincodeEvents(mockConnection, WithNewStreamResult(mockStream))

			messages := make(chan *gateway.SignedChaincodeEventsRequest, 1)
			ExpectSendMsg(mockStream, CaptureSendMsg(messages))
			mockStream.EXPECT().CloseSend().Return(nil)
			ExpectRecvMsg(mockStream).Maybe().Return(io.EOF)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			network := AssertNewTestNetwork(t, "NETWORK", WithClientConnection(mockConnection))
			_, err := network.ChaincodeEvents(ctx, "CHAINCODE", testCase.options...)
			require.NoError(t, err)

			creator, err := network.signingID.Creator()
			require.NoError(t, err)

			expected := &gateway.ChaincodeEventsRequest{
				ChannelId:   "NETWORK",
				ChaincodeId: "CHAINCODE",
				Identity:    creator,
			}
			proto.Merge(expected, testCase.expected)
			actual := &gateway.ChaincodeEventsRequest{}
			AssertUnmarshal(t, (<-messages).Request, actual)
			AssertProtoEqual(t, expected, actual)
		})
	}

	t.Run("Closes event channel on receive error", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		mockStream := NewMockClientStream(t)
		ExpectChaincodeEvents(mockConnection, WithNewStreamResult(mockStream))

		ExpectSendMsg(mockStream)
		mockStream.EXPECT().CloseSend().Return(nil)
		ExpectRecvMsg(mockStream).Maybe().Return(errors.New("fake"))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithClientConnection(mockConnection))

		receive, err := network.ChaincodeEvents(ctx, "CHAINCODE")
		require.NoError(t, err)

		actual, ok := <-receive

		require.False(t, ok, "Expected event listening to be cancelled, got %v", actual)
	})

	t.Run("Receives events", func(t *testing.T) {
		expected := []*ChaincodeEvent{
			{
				BlockNumber:   1,
				ChaincodeName: "CHAINCODE",
				EventName:     "EVENT_1",
				Payload:       []byte("PAYLOAD_1"),
				TransactionID: "TRANSACTION_ID_1",
			},
			{
				BlockNumber:   1,
				ChaincodeName: "CHAINCODE",
				EventName:     "EVENT_2",
				Payload:       []byte("PAYLOAD_2"),
				TransactionID: "TRANSACTION_ID_2",
			},
			{
				BlockNumber:   2,
				ChaincodeName: "CHAINCODE",
				EventName:     "EVENT_3",
				Payload:       []byte("PAYLOAD_3"),
				TransactionID: "TRANSACTION_ID_3",
			},
		}

		mockConnection := NewMockClientConnInterface(t)
		mockStream := NewMockClientStream(t)
		options := make(chan []grpc.CallOption, 1)
		ExpectChaincodeEvents(mockConnection, CaptureNewStreamOptions(options), WithNewStreamResult(mockStream))

		messages := make(chan *gateway.SignedChaincodeEventsRequest, 1)
		ExpectSendMsg(mockStream, CaptureSendMsg(messages))
		mockStream.EXPECT().CloseSend().Return(nil)
		ExpectRecvMsg(mockStream, WithRecvMsgs(
			newChaincodeEventsResponse(expected[0:2]),
			newChaincodeEventsResponse(expected[2:]),
		))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithClientConnection(mockConnection))

		receive, err := network.ChaincodeEvents(ctx, "CHAINCODE")
		require.NoError(t, err)

		for _, event := range expected {
			actual := <-receive
			require.Equal(t, event, actual)
		}
	})

	t.Run("Uses specified gRPC call options", func(t *testing.T) {
		expected := grpc.WaitForReady(true)

		mockConnection := NewMockClientConnInterface(t)
		mockStream := NewMockClientStream(t)
		options := make(chan []grpc.CallOption, 1)
		ExpectChaincodeEvents(mockConnection, CaptureNewStreamOptions(options), WithNewStreamResult(mockStream))

		messages := make(chan *gateway.SignedChaincodeEventsRequest, 1)
		ExpectSendMsg(mockStream, CaptureSendMsg(messages))
		mockStream.EXPECT().CloseSend().Return(nil)
		ExpectRecvMsg(mockStream).Maybe().Return(io.EOF)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithClientConnection(mockConnection))

		request, err := network.NewChaincodeEventsRequest("CHAINCODE")
		require.NoError(t, err, "NewChaincodeEventsRequest")

		_, err = request.Events(ctx, expected)
		require.NoError(t, err, "Events")

		require.Contains(t, <-options, expected, "CallOptions")
	})
}
