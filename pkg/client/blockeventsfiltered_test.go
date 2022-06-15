/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFilteredBlockEvents(t *testing.T) {
	t.Run("Returns connect error", func(t *testing.T) {
		expected := NewStatusError(t, codes.Aborted, "BLOCK_EVENTS_ERROR")
		mockClient := NewMockDeliverClient(gomock.NewController(t))
		mockClient.EXPECT().DeliverFiltered(gomock.Any(), gomock.Any()).
			Return(nil, expected)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithDeliverClient(mockClient))
		_, err := network.FilteredBlockEvents(ctx)

		require.Equal(t, status.Code(expected), status.Code(err), "status code")
		require.Errorf(t, err, expected.Error(), "error message")
	})

	t.Run("Sends valid request with default start position", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockDeliverClient(controller)
		mockEvents := NewMockDeliver_DeliverFilteredClient(controller)

		mockClient.EXPECT().DeliverFiltered(gomock.Any(), gomock.Any()).
			Return(mockEvents, nil)

		payload := &common.Payload{}
		mockEvents.EXPECT().Send(gomock.Any()).
			Do(func(in *common.Envelope) {
				test.AssertUnmarshal(t, in.GetPayload(), payload)
			}).
			Return(nil).
			Times(1)
		mockEvents.EXPECT().Recv().
			Return(nil, errors.New("fake")).
			AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithDeliverClient(mockClient))
		_, err := network.FilteredBlockEvents(ctx)
		require.NoError(t, err)

		AssertValidBlockEventRequestHeader(t, payload, network.Name())
		actual := &orderer.SeekInfo{}
		test.AssertUnmarshal(t, payload.GetData(), actual)

		expected := &orderer.SeekInfo{
			Start: &orderer.SeekPosition{
				Type: &orderer.SeekPosition_NextCommit{
					NextCommit: &orderer.SeekNextCommit{},
				},
			},
			Stop: seekLargestBlockNumber(),
		}

		test.AssertProtoEqual(t, expected, actual)
	})

	t.Run("Sends valid request with specified start block number", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockDeliverClient(controller)
		mockEvents := NewMockDeliver_DeliverFilteredClient(controller)

		mockClient.EXPECT().DeliverFiltered(gomock.Any(), gomock.Any()).
			Return(mockEvents, nil)

		payload := &common.Payload{}
		mockEvents.EXPECT().Send(gomock.Any()).
			Do(func(in *common.Envelope) {
				test.AssertUnmarshal(t, in.GetPayload(), payload)
			}).
			Return(nil).
			Times(1)
		mockEvents.EXPECT().Recv().
			Return(nil, errors.New("fake")).
			AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithDeliverClient(mockClient))
		_, err := network.FilteredBlockEvents(ctx, WithStartBlock(418))
		require.NoError(t, err)

		AssertValidBlockEventRequestHeader(t, payload, network.Name())
		actual := &orderer.SeekInfo{}
		test.AssertUnmarshal(t, payload.GetData(), actual)

		expected := &orderer.SeekInfo{
			Start: &orderer.SeekPosition{
				Type: &orderer.SeekPosition_Specified{
					Specified: &orderer.SeekSpecified{
						Number: 418,
					},
				},
			},
			Stop: seekLargestBlockNumber(),
		}

		test.AssertProtoEqual(t, expected, actual)
	})

	t.Run("Closes event channel on receive error", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockDeliverClient(controller)
		mockEvents := NewMockDeliver_DeliverFilteredClient(controller)

		mockClient.EXPECT().DeliverFiltered(gomock.Any(), gomock.Any()).
			Return(mockEvents, nil)

		mockEvents.EXPECT().Send(gomock.Any()).
			Return(nil)
		mockEvents.EXPECT().Recv().
			Return(nil, errors.New("fake"))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithDeliverClient(mockClient))
		receive, err := network.FilteredBlockEvents(ctx, WithStartBlock(418))
		require.NoError(t, err)

		actual, ok := <-receive

		require.False(t, ok, "Expected event listening to be cancelled, got %v", actual)
	})

	t.Run("Receives events", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockDeliverClient(controller)
		mockEvents := NewMockDeliver_DeliverFilteredClient(controller)

		mockClient.EXPECT().DeliverFiltered(gomock.Any(), gomock.Any()).
			Return(mockEvents, nil)

		mockEvents.EXPECT().Send(gomock.Any()).
			Return(nil)

		blocks := []*peer.FilteredBlock{
			{
				ChannelId: "NETWORK",
				Number:    1,
				FilteredTransactions: []*peer.FilteredTransaction{
					{
						Txid: "TX1",
					},
				},
			},
			{
				ChannelId: "NETWORK",
				Number:    2,
				FilteredTransactions: []*peer.FilteredTransaction{
					{
						Txid: "TX2",
					},
				},
			},
		}
		responseIndex := 0
		mockEvents.EXPECT().Recv().
			DoAndReturn(func() (*peer.DeliverResponse, error) {
				if responseIndex >= len(blocks) {
					return nil, errors.New("fake")
				}
				response := &peer.DeliverResponse{
					Type: &peer.DeliverResponse_FilteredBlock{
						FilteredBlock: blocks[responseIndex],
					},
				}
				responseIndex++
				return response, nil
			}).
			AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithDeliverClient(mockClient))
		receive, err := network.FilteredBlockEvents(ctx)
		require.NoError(t, err)

		for _, event := range blocks {
			actual := <-receive
			test.AssertProtoEqual(t, event, actual)
		}
	})

	t.Run("Closes event channel on non-block message", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockDeliverClient(controller)
		mockEvents := NewMockDeliver_DeliverFilteredClient(controller)

		mockClient.EXPECT().DeliverFiltered(gomock.Any(), gomock.Any()).
			Return(mockEvents, nil)

		mockEvents.EXPECT().Send(gomock.Any()).
			Return(nil)

		block := &peer.FilteredBlock{
			ChannelId: "NETWORK",
			Number:    1,
			FilteredTransactions: []*peer.FilteredTransaction{
				{
					Txid: "TX1",
				},
			},
		}
		responses := []*peer.DeliverResponse{
			{
				Type: &peer.DeliverResponse_FilteredBlock{
					FilteredBlock: block,
				},
			},
			{
				Type: &peer.DeliverResponse_Status{
					Status: common.Status_SERVICE_UNAVAILABLE,
				},
			},
			{
				Type: &peer.DeliverResponse_FilteredBlock{
					FilteredBlock: block,
				},
			},
		}
		responseIndex := 0
		mockEvents.EXPECT().Recv().
			DoAndReturn(func() (*peer.DeliverResponse, error) {
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

		network := AssertNewTestNetwork(t, "NETWORK", WithDeliverClient(mockClient))
		receive, err := network.FilteredBlockEvents(ctx)
		require.NoError(t, err)

		expected := []*peer.FilteredBlock{
			block,
			nil,
			nil,
		}
		for _, event := range expected {
			actual := <-receive
			test.AssertProtoEqual(t, event, actual)
		}
	})

	t.Run("Uses specified gRPC call options", func(t *testing.T) {
		var actual []grpc.CallOption
		expected := grpc.WaitForReady(true)

		controller := gomock.NewController(t)
		mockClient := NewMockDeliverClient(controller)
		mockEvents := NewMockDeliver_DeliverClient(controller)

		mockClient.EXPECT().DeliverFiltered(gomock.Any(), gomock.Any()).
			Do(func(_ context.Context, opts ...grpc.CallOption) {
				actual = opts
			}).
			Return(mockEvents, nil).
			Times(1)

		mockEvents.EXPECT().Send(gomock.Any()).
			Return(nil).
			AnyTimes()
		mockEvents.EXPECT().Recv().
			Return(nil, errors.New("fake")).
			AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithDeliverClient(mockClient))
		request, err := network.NewFilteredBlockEventsRequest()
		require.NoError(t, err, "NewFilteredBlockEventsRequest")

		_, err = request.Events(ctx, expected)
		require.NoError(t, err, "Events")

		require.Contains(t, actual, expected, "CallOptions")
	})
}
