/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"errors"
	"testing"

	"github.com/hyperledger/fabric-gateway/pkg/internal/test"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/ledger/rwset"
	"github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestBlockAndPrivateDataEvents(t *testing.T) {
	t.Run("Returns connect error", func(t *testing.T) {
		expected := NewStatusError(t, codes.Aborted, "BLOCK_EVENTS_ERROR")
		mockClient := NewMockDeliverClient(gomock.NewController(t))
		mockClient.EXPECT().DeliverWithPrivateData(gomock.Any(), gomock.Any()).
			Return(nil, expected)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithDeliverClient(mockClient))
		_, err := network.BlockAndPrivateDataEvents(ctx)

		require.Equal(t, status.Code(expected), status.Code(err), "status code")
		require.ErrorIs(t, err, expected, "error type: %T", err)
		require.ErrorContains(t, err, expected.Error(), "message")
	})

	t.Run("Sends valid request with default start position", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockDeliverClient(controller)
		mockEvents := NewMockDeliver_DeliverWithPrivateDataClient(controller)

		mockClient.EXPECT().DeliverWithPrivateData(gomock.Any(), gomock.Any()).
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
		_, err := network.BlockAndPrivateDataEvents(ctx)
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
		mockEvents := NewMockDeliver_DeliverWithPrivateDataClient(controller)

		mockClient.EXPECT().DeliverWithPrivateData(gomock.Any(), gomock.Any()).
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
		_, err := network.BlockAndPrivateDataEvents(ctx, WithStartBlock(418))
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
		mockEvents := NewMockDeliver_DeliverWithPrivateDataClient(controller)

		mockClient.EXPECT().DeliverWithPrivateData(gomock.Any(), gomock.Any()).
			Return(mockEvents, nil)

		mockEvents.EXPECT().Send(gomock.Any()).
			Return(nil)
		mockEvents.EXPECT().Recv().
			Return(nil, errors.New("fake"))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithDeliverClient(mockClient))
		receive, err := network.BlockAndPrivateDataEvents(ctx, WithStartBlock(418))
		require.NoError(t, err)

		actual, ok := <-receive

		require.False(t, ok, "Expected event listening to be cancelled, got %v", actual)
	})

	t.Run("Receives events", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockDeliverClient(controller)
		mockEvents := NewMockDeliver_DeliverWithPrivateDataClient(controller)

		mockClient.EXPECT().DeliverWithPrivateData(gomock.Any(), gomock.Any()).
			Return(mockEvents, nil)

		mockEvents.EXPECT().Send(gomock.Any()).
			Return(nil)

		blocksAndPrivateData := []*peer.BlockAndPrivateData{
			{
				Block: &common.Block{
					Header: &common.BlockHeader{
						Number: 1,
					},
					Data: &common.BlockData{
						Data: [][]byte{
							[]byte("data1"),
						},
					},
				},
				PrivateDataMap: map[uint64]*rwset.TxPvtReadWriteSet{
					0: {
						DataModel:  rwset.TxReadWriteSet_KV,
						NsPvtRwset: []*rwset.NsPvtReadWriteSet{},
					},
				},
			},
			{
				Block: &common.Block{
					Header: &common.BlockHeader{
						Number: 2,
					},
					Data: &common.BlockData{
						Data: [][]byte{
							[]byte("data2"),
						},
					},
				},
				PrivateDataMap: map[uint64]*rwset.TxPvtReadWriteSet{
					0: {
						DataModel:  rwset.TxReadWriteSet_KV,
						NsPvtRwset: []*rwset.NsPvtReadWriteSet{},
					},
				},
			},
		}
		responseIndex := 0
		mockEvents.EXPECT().Recv().
			DoAndReturn(func() (*peer.DeliverResponse, error) {
				if responseIndex >= len(blocksAndPrivateData) {
					return nil, errors.New("fake")
				}
				response := &peer.DeliverResponse{
					Type: &peer.DeliverResponse_BlockAndPrivateData{
						BlockAndPrivateData: blocksAndPrivateData[responseIndex],
					},
				}
				responseIndex++
				return response, nil
			}).
			AnyTimes()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithDeliverClient(mockClient))
		receive, err := network.BlockAndPrivateDataEvents(ctx)
		require.NoError(t, err)

		for _, event := range blocksAndPrivateData {
			actual := <-receive
			test.AssertProtoEqual(t, event, actual)
		}
	})

	t.Run("Closes event channel on non-block message", func(t *testing.T) {
		controller := gomock.NewController(t)
		mockClient := NewMockDeliverClient(controller)
		mockEvents := NewMockDeliver_DeliverWithPrivateDataClient(controller)

		mockClient.EXPECT().DeliverWithPrivateData(gomock.Any(), gomock.Any()).
			Return(mockEvents, nil)

		mockEvents.EXPECT().Send(gomock.Any()).
			Return(nil)

		blockAndPrivateData := &peer.BlockAndPrivateData{
			Block: &common.Block{
				Header: &common.BlockHeader{
					Number: 1,
				},
				Data: &common.BlockData{
					Data: [][]byte{
						[]byte("data1"),
					},
				},
			},
			PrivateDataMap: map[uint64]*rwset.TxPvtReadWriteSet{
				0: {
					DataModel:  rwset.TxReadWriteSet_KV,
					NsPvtRwset: []*rwset.NsPvtReadWriteSet{},
				},
			},
		}
		responses := []*peer.DeliverResponse{
			{
				Type: &peer.DeliverResponse_BlockAndPrivateData{
					BlockAndPrivateData: blockAndPrivateData,
				},
			},
			{
				Type: &peer.DeliverResponse_Status{
					Status: common.Status_SERVICE_UNAVAILABLE,
				},
			},
			{
				Type: &peer.DeliverResponse_BlockAndPrivateData{
					BlockAndPrivateData: blockAndPrivateData,
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
		receive, err := network.BlockAndPrivateDataEvents(ctx)
		require.NoError(t, err)

		expected := []*peer.BlockAndPrivateData{
			blockAndPrivateData,
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

		mockClient.EXPECT().DeliverWithPrivateData(gomock.Any(), gomock.Any()).
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
		request, err := network.NewBlockAndPrivateDataEventsRequest()
		require.NoError(t, err, "NewBlockAndPrivateDataEventsRequest")

		_, err = request.Events(ctx, expected)
		require.NoError(t, err, "Events")

		require.Contains(t, actual, expected, "CallOptions")
	})

	t.Run("Sends request with TLS client certificate hash", func(t *testing.T) {
		expected := []byte("TLS_CLIENT_CERTIFICATE_HASH")

		controller := gomock.NewController(t)
		mockClient := NewMockDeliverClient(controller)
		mockEvents := NewMockDeliver_DeliverClient(controller)

		mockClient.EXPECT().DeliverWithPrivateData(gomock.Any(), gomock.Any()).
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

		network := AssertNewTestNetwork(t, "NETWORK", WithDeliverClient(mockClient), WithTLSClientCertificateHash(expected))
		_, err := network.BlockAndPrivateDataEvents(ctx)
		require.NoError(t, err)

		channelHeader := &common.ChannelHeader{}
		test.AssertUnmarshal(t, payload.GetHeader().GetChannelHeader(), channelHeader)

		require.Equal(t, expected, channelHeader.GetTlsCertHash())
	})
}
