/*
Copyright 2021 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"errors"
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
	"github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		mockClient := NewMockGatewayClient(gomock.NewController(t))
		mockClient.EXPECT().ChaincodeEvents(gomock.Any(), gomock.Any()).
			Return(nil, expected)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithGatewayClient(mockClient))
		_, err := network.ChaincodeEvents(ctx, "CHAINCODE")

		require.Equal(t, status.Code(expected), status.Code(err), "status code")
		require.ErrorIs(t, err, expected, "error type: %T", err)
		require.ErrorContains(t, err, expected.Error(), "message")
	})

	t.Run("Sends valid request with default start position", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockGatewayClient(controller)
		mockEvents := NewMockGateway_ChaincodeEventsClient(controller)

		var actual *gateway.ChaincodeEventsRequest
		mockClient.EXPECT().ChaincodeEvents(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SignedChaincodeEventsRequest, _ ...grpc.CallOption) {
				request := &gateway.ChaincodeEventsRequest{}
				test.AssertUnmarshal(t, in.GetRequest(), request)
				actual = request
			}).
			Return(mockEvents, nil).
			Times(1)

		mockEvents.EXPECT().Recv().
			Return(nil, errors.New("fake")).
			AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithGatewayClient(mockClient))
		_, err := network.ChaincodeEvents(ctx, "CHAINCODE")
		require.NoError(t, err)

		creator, err := network.signingID.Creator()
		require.NoError(t, err)

		expected := &gateway.ChaincodeEventsRequest{
			ChannelId:   "NETWORK",
			ChaincodeId: "CHAINCODE",
			Identity:    creator,
			StartPosition: &orderer.SeekPosition{
				Type: &orderer.SeekPosition_NextCommit{
					NextCommit: &orderer.SeekNextCommit{},
				},
			},
		}
		test.AssertProtoEqual(t, expected, actual)
	})

	t.Run("Sends valid request with specified start block number", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockGatewayClient(controller)
		mockEvents := NewMockGateway_ChaincodeEventsClient(controller)

		var actual *gateway.ChaincodeEventsRequest
		mockClient.EXPECT().ChaincodeEvents(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SignedChaincodeEventsRequest, _ ...grpc.CallOption) {
				request := &gateway.ChaincodeEventsRequest{}
				test.AssertUnmarshal(t, in.GetRequest(), request)
				actual = request
			}).
			Return(mockEvents, nil).
			Times(1)

		mockEvents.EXPECT().Recv().
			Return(nil, errors.New("fake")).
			AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithGatewayClient(mockClient))
		_, err := network.ChaincodeEvents(ctx, "CHAINCODE", WithStartBlock(418))
		require.NoError(t, err)

		creator, err := network.signingID.Creator()
		require.NoError(t, err)

		expected := &gateway.ChaincodeEventsRequest{
			ChannelId:   "NETWORK",
			ChaincodeId: "CHAINCODE",
			Identity:    creator,
			StartPosition: &orderer.SeekPosition{
				Type: &orderer.SeekPosition_Specified{
					Specified: &orderer.SeekSpecified{
						Number: 418,
					},
				},
			},
		}
		test.AssertProtoEqual(t, expected, actual)
	})

	t.Run("Sends valid request with specified start block number and fresh checkpointer", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockGatewayClient(controller)
		mockEvents := NewMockGateway_ChaincodeEventsClient(controller)

		var actual *gateway.ChaincodeEventsRequest
		mockClient.EXPECT().ChaincodeEvents(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SignedChaincodeEventsRequest, _ ...grpc.CallOption) {
				request := &gateway.ChaincodeEventsRequest{}
				test.AssertUnmarshal(t, in.GetRequest(), request)
				actual = request
			}).
			Return(mockEvents, nil).
			Times(1)

		mockEvents.EXPECT().Recv().
			Return(nil, errors.New("fake")).
			AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithGatewayClient(mockClient))

		checkpointer := new(InMemoryCheckpointer)

		_, err := network.ChaincodeEvents(ctx, "CHAINCODE", WithStartBlock(418), WithCheckpoint(checkpointer))
		require.NoError(t, err)

		creator, err := network.signingID.Creator()
		require.NoError(t, err)

		expected := &gateway.ChaincodeEventsRequest{
			ChannelId:   "NETWORK",
			ChaincodeId: "CHAINCODE",
			Identity:    creator,
			StartPosition: &orderer.SeekPosition{
				Type: &orderer.SeekPosition_Specified{
					Specified: &orderer.SeekSpecified{
						Number: 418,
					},
				},
			},
		}
		test.AssertProtoEqual(t, expected, actual)
	})
	t.Run("Sends valid request with specified start block and checkpoint block", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockGatewayClient(controller)
		mockEvents := NewMockGateway_ChaincodeEventsClient(controller)

		var actual *gateway.ChaincodeEventsRequest
		mockClient.EXPECT().ChaincodeEvents(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SignedChaincodeEventsRequest, _ ...grpc.CallOption) {
				request := &gateway.ChaincodeEventsRequest{}
				test.AssertUnmarshal(t, in.GetRequest(), request)
				actual = request
			}).
			Return(mockEvents, nil).
			Times(1)

		mockEvents.EXPECT().Recv().
			Return(nil, errors.New("fake")).
			AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithGatewayClient(mockClient))

		checkpointer := new(InMemoryCheckpointer)
		checkpointer.CheckpointBlock(uint64(500))
		_, err := network.ChaincodeEvents(ctx, "CHAINCODE", WithStartBlock(418), WithCheckpoint(checkpointer))
		require.NoError(t, err)

		creator, err := network.signingID.Creator()
		require.NoError(t, err)

		expected := &gateway.ChaincodeEventsRequest{
			ChannelId:   "NETWORK",
			ChaincodeId: "CHAINCODE",
			Identity:    creator,
			StartPosition: &orderer.SeekPosition{
				Type: &orderer.SeekPosition_Specified{
					Specified: &orderer.SeekSpecified{
						Number: 501,
					},
				},
			},
		}
		test.AssertProtoEqual(t, expected, actual)
	})

	t.Run("Sends valid request with specified start block and checkpoint transaction ID", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockGatewayClient(controller)
		mockEvents := NewMockGateway_ChaincodeEventsClient(controller)

		var actual *gateway.ChaincodeEventsRequest
		mockClient.EXPECT().ChaincodeEvents(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SignedChaincodeEventsRequest, _ ...grpc.CallOption) {
				request := &gateway.ChaincodeEventsRequest{}
				test.AssertUnmarshal(t, in.GetRequest(), request)
				actual = request
			}).
			Return(mockEvents, nil).
			Times(1)

		mockEvents.EXPECT().Recv().
			Return(nil, errors.New("fake")).
			AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithGatewayClient(mockClient))

		checkpointer := new(InMemoryCheckpointer)
		checkpointer.CheckpointTransaction(uint64(500), "txn1")
		_, err := network.ChaincodeEvents(ctx, "CHAINCODE", WithStartBlock(418), WithCheckpoint(checkpointer))
		require.NoError(t, err)

		creator, err := network.signingID.Creator()
		require.NoError(t, err)

		expected := &gateway.ChaincodeEventsRequest{
			ChannelId:   "NETWORK",
			ChaincodeId: "CHAINCODE",
			Identity:    creator,
			StartPosition: &orderer.SeekPosition{
				Type: &orderer.SeekPosition_Specified{
					Specified: &orderer.SeekSpecified{
						Number: 500,
					},
				},
			},
			AfterTransactionId: "txn1",
		}
		test.AssertProtoEqual(t, expected, actual)
	})

	t.Run("Sends valid request with no start block and fresh checkpointer", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockGatewayClient(controller)
		mockEvents := NewMockGateway_ChaincodeEventsClient(controller)

		var actual *gateway.ChaincodeEventsRequest
		mockClient.EXPECT().ChaincodeEvents(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SignedChaincodeEventsRequest, _ ...grpc.CallOption) {
				request := &gateway.ChaincodeEventsRequest{}
				test.AssertUnmarshal(t, in.GetRequest(), request)
				actual = request
			}).
			Return(mockEvents, nil).
			Times(1)

		mockEvents.EXPECT().Recv().
			Return(nil, errors.New("fake")).
			AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithGatewayClient(mockClient))

		checkpointer := new(InMemoryCheckpointer)
		_, err := network.ChaincodeEvents(ctx, "CHAINCODE", WithCheckpoint(checkpointer))
		require.NoError(t, err)

		creator, err := network.signingID.Creator()
		require.NoError(t, err)
		expected := &gateway.ChaincodeEventsRequest{
			ChannelId:   "NETWORK",
			ChaincodeId: "CHAINCODE",
			Identity:    creator,
			StartPosition: &orderer.SeekPosition{
				Type: &orderer.SeekPosition_NextCommit{
					NextCommit: &orderer.SeekNextCommit{},
				},
			},
		}
		test.AssertProtoEqual(t, expected, actual)
	})

	t.Run("Sends valid request with no start block and checkpoint transaction ID", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockGatewayClient(controller)
		mockEvents := NewMockGateway_ChaincodeEventsClient(controller)

		var actual *gateway.ChaincodeEventsRequest
		mockClient.EXPECT().ChaincodeEvents(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SignedChaincodeEventsRequest, _ ...grpc.CallOption) {
				request := &gateway.ChaincodeEventsRequest{}
				test.AssertUnmarshal(t, in.GetRequest(), request)
				actual = request
			}).
			Return(mockEvents, nil).
			Times(1)

		mockEvents.EXPECT().Recv().
			Return(nil, errors.New("fake")).
			AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithGatewayClient(mockClient))

		checkpointer := new(InMemoryCheckpointer)
		checkpointer.CheckpointTransaction(uint64(500), "txn1")

		_, err := network.ChaincodeEvents(ctx, "CHAINCODE", WithCheckpoint(checkpointer))
		require.NoError(t, err)

		creator, err := network.signingID.Creator()
		require.NoError(t, err)
		expected := &gateway.ChaincodeEventsRequest{
			ChannelId:   "NETWORK",
			ChaincodeId: "CHAINCODE",
			Identity:    creator,
			StartPosition: &orderer.SeekPosition{
				Type: &orderer.SeekPosition_Specified{
					Specified: &orderer.SeekSpecified{
						Number: 500,
					},
				},
			},
			AfterTransactionId: "txn1",
		}
		test.AssertProtoEqual(t, expected, actual)
	})

	t.Run("Sends valid request with with start block and checkpoint chaincode event", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockGatewayClient(controller)
		mockEvents := NewMockGateway_ChaincodeEventsClient(controller)

		var actual *gateway.ChaincodeEventsRequest
		mockClient.EXPECT().ChaincodeEvents(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, in *gateway.SignedChaincodeEventsRequest, _ ...grpc.CallOption) {
				request := &gateway.ChaincodeEventsRequest{}
				test.AssertUnmarshal(t, in.GetRequest(), request)
				actual = request
			}).
			Return(mockEvents, nil).
			Times(1)

		mockEvents.EXPECT().Recv().
			Return(nil, errors.New("fake")).
			AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithGatewayClient(mockClient))

		checkpointer := new(InMemoryCheckpointer)
		event := &ChaincodeEvent{
			BlockNumber:   1,
			ChaincodeName: "CHAINCODE",
			EventName:     "EVENT_1",
			Payload:       []byte("PAYLOAD_1"),
			TransactionID: "TRANSACTION_1",
		}

		checkpointer.CheckpointChaincodeEvent(event)

		_, err := network.ChaincodeEvents(ctx, "CHAINCODE", WithStartBlock(418), WithCheckpoint(checkpointer))
		require.NoError(t, err)

		creator, err := network.signingID.Creator()
		require.NoError(t, err)
		expected := &gateway.ChaincodeEventsRequest{
			ChannelId:   "NETWORK",
			ChaincodeId: "CHAINCODE",
			Identity:    creator,
			StartPosition: &orderer.SeekPosition{
				Type: &orderer.SeekPosition_Specified{
					Specified: &orderer.SeekSpecified{
						Number: event.BlockNumber,
					},
				},
			},
			AfterTransactionId: event.TransactionID,
		}
		test.AssertProtoEqual(t, expected, actual)
	})

	t.Run("Closes event channel on receive error", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockGatewayClient(controller)
		mockEvents := NewMockGateway_ChaincodeEventsClient(controller)

		mockClient.EXPECT().ChaincodeEvents(gomock.Any(), gomock.Any()).
			Return(mockEvents, nil)

		mockEvents.EXPECT().Recv().
			Return(nil, errors.New("fake")).
			AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithGatewayClient(mockClient))
		receive, err := network.ChaincodeEvents(ctx, "CHAINCODE")
		require.NoError(t, err)

		actual, ok := <-receive

		require.False(t, ok, "Expected event listening to be cancelled, got %v", actual)
	})

	t.Run("Receives events", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockGatewayClient(controller)
		mockEvents := NewMockGateway_ChaincodeEventsClient(controller)

		mockClient.EXPECT().ChaincodeEvents(gomock.Any(), gomock.Any()).
			Return(mockEvents, nil)

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

		responses := []*gateway.ChaincodeEventsResponse{
			newChaincodeEventsResponse(expected[0:2]),
			newChaincodeEventsResponse(expected[2:]),
		}
		responseIndex := 0
		mockEvents.EXPECT().Recv().
			DoAndReturn(func() (*gateway.ChaincodeEventsResponse, error) {
				if responseIndex >= len(responses) {
					return nil, errors.New("fake")
				}
				response := responses[responseIndex]
				responseIndex++
				return response, nil
			}).
			AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithGatewayClient(mockClient))
		receive, err := network.ChaincodeEvents(ctx, "CHAINCODE")
		require.NoError(t, err)

		for _, event := range expected {
			actual := <-receive
			require.EqualValues(t, event, actual)
		}
	})

	t.Run("Uses specified gRPC call options", func(t *testing.T) {
		var actual []grpc.CallOption
		expected := grpc.WaitForReady(true)

		controller := gomock.NewController(t)
		mockClient := NewMockGatewayClient(controller)
		mockEvents := NewMockGateway_ChaincodeEventsClient(controller)

		mockClient.EXPECT().ChaincodeEvents(gomock.Any(), gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, _ *gateway.SignedChaincodeEventsRequest, opts ...grpc.CallOption) {
				actual = opts
			}).
			Return(mockEvents, nil).
			Times(1)

		mockEvents.EXPECT().Recv().
			Return(nil, errors.New("fake")).
			AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithGatewayClient(mockClient))
		request, err := network.NewChaincodeEventsRequest("CHAINCODE")
		require.NoError(t, err, "NewChaincodeEventsRequest")

		_, err = request.Events(ctx, expected)
		require.NoError(t, err, "Events")

		require.Contains(t, actual, expected, "CallOptions")
	})
}
