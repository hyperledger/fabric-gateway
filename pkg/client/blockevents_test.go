// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/orderer"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestBlockEvents(t *testing.T) {
	t.Run("Returns connect error", func(t *testing.T) {
		expected := NewStatusError(t, codes.Aborted, "BLOCK_EVENTS_ERROR")

		mockConnection := NewMockClientConnInterface(t)
		ExpectDeliver(mockConnection, WithNewStreamError(expected))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithClientConnection(mockConnection))
		_, err := network.BlockEvents(ctx)

		require.Equal(t, status.Code(expected), status.Code(err), "status code")
		require.ErrorIs(t, err, expected, "error type: %T", err)
		require.ErrorContains(t, err, expected.Error(), "message")
	})

	for testName, testCase := range map[string]struct {
		options  []BlockEventsOption
		expected *orderer.SeekInfo
	}{
		"Sends valid request with default start position": {
			options: nil,
			expected: &orderer.SeekInfo{
				Start: &orderer.SeekPosition{
					Type: &orderer.SeekPosition_NextCommit{
						NextCommit: &orderer.SeekNextCommit{},
					},
				},
				Stop: seekLargestBlockNumber(),
			},
		},
		"Sends valid request with specified start block number": {
			options: []BlockEventsOption{
				WithStartBlock(418),
			},
			expected: &orderer.SeekInfo{
				Start: &orderer.SeekPosition{
					Type: &orderer.SeekPosition_Specified{
						Specified: &orderer.SeekSpecified{
							Number: 418,
						},
					},
				},
				Stop: seekLargestBlockNumber(),
			},
		},
		"Uses specified start block instead of unset checkpoint": {
			options: []BlockEventsOption{
				WithStartBlock(418),
				WithCheckpoint(new(InMemoryCheckpointer)),
			},
			expected: &orderer.SeekInfo{
				Start: &orderer.SeekPosition{
					Type: &orderer.SeekPosition_Specified{
						Specified: &orderer.SeekSpecified{
							Number: 418,
						},
					},
				},
				Stop: seekLargestBlockNumber(),
			},
		},
		"Uses checkpoint block instead of specified start block": {
			options: func() []BlockEventsOption {
				checkpointer := new(InMemoryCheckpointer)
				checkpointer.CheckpointBlock(500)
				return []BlockEventsOption{
					WithStartBlock(418),
					WithCheckpoint(checkpointer),
				}
			}(),
			expected: &orderer.SeekInfo{
				Start: &orderer.SeekPosition{
					Type: &orderer.SeekPosition_Specified{
						Specified: &orderer.SeekSpecified{
							Number: 501,
						},
					},
				},
				Stop: seekLargestBlockNumber(),
			},
		},
		"Uses checkpoint block zero with set transaction ID instead of specified start block": {
			options: func() []BlockEventsOption {
				checkpointer := new(InMemoryCheckpointer)
				checkpointer.CheckpointTransaction(0, "transctionId")
				return []BlockEventsOption{
					WithStartBlock(418),
					WithCheckpoint(checkpointer),
				}
			}(),
			expected: &orderer.SeekInfo{
				Start: &orderer.SeekPosition{
					Type: &orderer.SeekPosition_Specified{
						Specified: &orderer.SeekSpecified{
							Number: 0,
						},
					},
				},
				Stop: seekLargestBlockNumber(),
			},
		},
		"Uses default start position with unset checkpoint and no start block": {
			options: []BlockEventsOption{
				WithCheckpoint(new(InMemoryCheckpointer)),
			},
			expected: &orderer.SeekInfo{
				Start: &orderer.SeekPosition{
					Type: &orderer.SeekPosition_NextCommit{
						NextCommit: &orderer.SeekNextCommit{},
					},
				},
				Stop: seekLargestBlockNumber(),
			},
		},
	} {
		t.Run(testName, func(t *testing.T) {
			mockConnection := NewMockClientConnInterface(t)
			mockStream := NewMockClientStream(t)
			ExpectDeliver(mockConnection, WithNewStreamResult(mockStream))

			messages := make(chan *common.Envelope, 1)
			ExpectSendMsg(mockStream, CaptureSendMsg(messages))
			mockStream.EXPECT().CloseSend().Maybe().Return(nil)
			ExpectRecvMsg(mockStream).Maybe().Return(io.EOF)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			network := AssertNewTestNetwork(t, "NETWORK", WithClientConnection(mockConnection))
			_, err := network.BlockEvents(ctx, testCase.options...)
			require.NoError(t, err)

			payload := &common.Payload{}
			AssertUnmarshal(t, (<-messages).GetPayload(), payload)
			AssertValidBlockEventRequestHeader(t, payload, network.Name())
			actual := &orderer.SeekInfo{}
			AssertUnmarshal(t, payload.GetData(), actual)

			AssertProtoEqual(t, testCase.expected, actual)
		})
	}

	t.Run("Closes event channel on receive error", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		mockStream := NewMockClientStream(t)
		ExpectDeliver(mockConnection, WithNewStreamResult(mockStream))

		ExpectSendMsg(mockStream)
		mockStream.EXPECT().CloseSend().Maybe().Return(nil)
		ExpectRecvMsg(mockStream).Return(errors.New("fake"))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithClientConnection(mockConnection))
		receive, err := network.BlockEvents(ctx, WithStartBlock(418))
		require.NoError(t, err)

		actual, ok := <-receive

		require.False(t, ok, "Expected event listening to be cancelled, got %v", actual)
	})

	t.Run("Receives events", func(t *testing.T) {
		expected := []*common.Block{
			{
				Header: &common.BlockHeader{
					Number: 1,
				},
				Data: &common.BlockData{
					Data: [][]byte{
						[]byte("data1"),
					},
				},
			},
			{
				Header: &common.BlockHeader{
					Number: 2,
				},
				Data: &common.BlockData{
					Data: [][]byte{
						[]byte("data2"),
					},
				},
			},
		}

		mockConnection := NewMockClientConnInterface(t)
		mockStream := NewMockClientStream(t)
		ExpectDeliver(mockConnection, WithNewStreamResult(mockStream))

		ExpectSendMsg(mockStream)
		mockStream.EXPECT().CloseSend().Maybe().Return(nil)

		var responses []*peer.DeliverResponse
		for _, block := range expected {
			responses = append(responses, &peer.DeliverResponse{
				Type: &peer.DeliverResponse_Block{
					Block: block,
				},
			})
		}
		ExpectRecvMsg(mockStream, WithRecvMsgs(responses...))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithClientConnection(mockConnection))
		receive, err := network.BlockEvents(ctx)
		require.NoError(t, err)

		for _, event := range expected {
			actual := <-receive
			AssertProtoEqual(t, event, actual)
		}
	})

	t.Run("Closes event channel on non-block message", func(t *testing.T) {
		mockConnection := NewMockClientConnInterface(t)
		mockStream := NewMockClientStream(t)
		ExpectDeliver(mockConnection, WithNewStreamResult(mockStream))

		ExpectSendMsg(mockStream)
		mockStream.EXPECT().CloseSend().Maybe().Return(nil)

		block := &common.Block{
			Header: &common.BlockHeader{
				Number: 1,
			},
			Data: &common.BlockData{
				Data: [][]byte{
					[]byte("data1"),
				},
			},
		}
		responses := []*peer.DeliverResponse{
			{
				Type: &peer.DeliverResponse_Block{
					Block: block,
				},
			},
			{
				Type: &peer.DeliverResponse_Status{
					Status: common.Status_SERVICE_UNAVAILABLE,
				},
			},
			{
				Type: &peer.DeliverResponse_Block{
					Block: block,
				},
			},
		}
		ExpectRecvMsg(mockStream, WithRecvMsgs(responses...))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithClientConnection(mockConnection))
		receive, err := network.BlockEvents(ctx)
		require.NoError(t, err)

		expected := []*common.Block{
			block,
			nil,
			nil,
		}
		for _, event := range expected {
			actual := <-receive
			AssertProtoEqual(t, event, actual)
		}
	})

	t.Run("Uses specified gRPC call options", func(t *testing.T) {
		expected := grpc.WaitForReady(true)

		mockConnection := NewMockClientConnInterface(t)
		mockStream := NewMockClientStream(t)
		options := make(chan []grpc.CallOption, 1)
		ExpectDeliver(mockConnection, CaptureNewStreamOptions(options), WithNewStreamResult(mockStream))

		ExpectSendMsg(mockStream)
		mockStream.EXPECT().CloseSend().Maybe().Return(nil)
		ExpectRecvMsg(mockStream).Maybe().Return(io.EOF)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithClientConnection(mockConnection))
		request, err := network.NewBlockEventsRequest()
		require.NoError(t, err, "NewBlockEventsRequest")

		_, err = request.Events(ctx, expected)
		require.NoError(t, err, "Events")

		require.Contains(t, (<-options), expected, "CallOptions")
	})

	t.Run("Sends request with TLS client certificate hash", func(t *testing.T) {
		expected := []byte("TLS_CLIENT_CERTIFICATE_HASH")

		mockConnection := NewMockClientConnInterface(t)
		mockStream := NewMockClientStream(t)
		ExpectDeliver(mockConnection, WithNewStreamResult(mockStream))

		requests := make(chan *common.Envelope, 1)
		ExpectSendMsg(mockStream, CaptureSendMsg(requests))
		mockStream.EXPECT().CloseSend().Maybe().Return(nil)
		ExpectRecvMsg(mockStream).Maybe().Return(io.EOF)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		network := AssertNewTestNetwork(t, "NETWORK", WithClientConnection(mockConnection), WithTLSClientCertificateHash(expected))
		_, err := network.BlockEvents(ctx)
		require.NoError(t, err)

		payload := &common.Payload{}
		AssertUnmarshal(t, (<-requests).GetPayload(), payload)
		channelHeader := &common.ChannelHeader{}
		AssertUnmarshal(t, payload.GetHeader().GetChannelHeader(), channelHeader)

		require.Equal(t, expected, channelHeader.GetTlsCertHash())
	})
}
