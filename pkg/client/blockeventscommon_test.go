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
	"google.golang.org/protobuf/proto"
)

func TestCommonBlockEvents(t *testing.T) {
	for eventType, newTester := range map[string]func(*testing.T) blockEventsTester{
		"Block": func(t *testing.T) blockEventsTester {
			return NewBlockEventsTest(t)
		},
		"FilteredBlock": func(t *testing.T) blockEventsTester {
			return NewFilteredBlockEventsTest(t)
		},
		"BlockAndPrivateData": func(t *testing.T) blockEventsTester {
			return NewBlockAndPrivateDataEventsTest(t)
		},
	} {
		t.Run(eventType, func(t *testing.T) {
			t.Run("Returns connect error", func(t *testing.T) {
				expected := NewStatusError(t, codes.Aborted, "BLOCK_EVENTS_ERROR")

				tester := newTester(t)
				tester.SetConnectError(expected)

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				err := tester.Events(ctx)

				require.Equal(t, status.Code(expected), status.Code(err), "status code")
				require.ErrorIs(t, err, expected, "error type: %T", err)
				require.ErrorContains(t, err, expected.Error(), "message")
			})
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
				networkName := "NETWORK"

				tester := newTester(t)
				tester.SetNetworkName(networkName)

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				err := tester.Events(ctx, testCase.options...)
				require.NoError(t, err)

				payload := &common.Payload{}
				AssertUnmarshal(t, (<-tester.Requests()).GetPayload(), payload)
				AssertValidBlockEventRequestHeader(t, payload, networkName)
				actual := &orderer.SeekInfo{}
				AssertUnmarshal(t, payload.GetData(), actual)

				AssertProtoEqual(t, testCase.expected, actual)
			})
		}

		t.Run("Closes event channel on receive error", func(t *testing.T) {
			tester := newTester(t)
			tester.SetReceiveError(errors.New("fake"))

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := tester.Events(ctx)
			require.NoError(t, err)

			actual, ok := tester.Receive()

			require.False(t, ok, "Expected event listening to be cancelled, got %v", actual)
		})

		t.Run("Uses specified gRPC call options", func(t *testing.T) {
			expected := grpc.WaitForReady(true)

			tester := newTester(t)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := tester.EventsWithCallOptions(ctx, expected)
			require.NoError(t, err)

			require.Contains(t, (<-tester.CallOptions()), expected, "CallOptions")
		})

		t.Run("Sends request with TLS client certificate hash", func(t *testing.T) {
			expected := []byte("TLS_CLIENT_CERTIFICATE_HASH")

			tester := newTester(t)
			tester.SetGatewayOptions(WithTLSClientCertificateHash(expected))

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := tester.Events(ctx)
			require.NoError(t, err)

			payload := &common.Payload{}
			AssertUnmarshal(t, (<-tester.Requests()).GetPayload(), payload)
			channelHeader := &common.ChannelHeader{}
			AssertUnmarshal(t, payload.GetHeader().GetChannelHeader(), channelHeader)

			require.Equal(t, expected, channelHeader.GetTlsCertHash())
		})
	}
}

type blockEventsTester interface {
	SetConnectError(error)
	SetNetworkName(string)
	SetGatewayOptions(...ConnectOption)
	SetReceiveError(error)
	SetResponses(...*peer.DeliverResponse)
	Events(context.Context, ...BlockEventsOption) error
	Requests() <-chan *common.Envelope
	EventsWithCallOptions(context.Context, ...grpc.CallOption) error
	CallOptions() <-chan []grpc.CallOption
	Receive() (proto.Message, bool)
}

type baseBlockEventsTest struct {
	t              *testing.T
	connectError   error
	networkName    string
	gatewayOptions []ConnectOption
	receiveError   error
	responses      []*peer.DeliverResponse
	requests       chan *common.Envelope
	callOptions    chan []grpc.CallOption
}

func (b *baseBlockEventsTest) SetConnectError(err error) {
	b.connectError = err
}

func (b *baseBlockEventsTest) SetNetworkName(channel string) {
	b.networkName = channel
}

func (b *baseBlockEventsTest) SetGatewayOptions(options ...ConnectOption) {
	b.gatewayOptions = append(b.gatewayOptions, options...)
}

func (b *baseBlockEventsTest) SetReceiveError(err error) {
	b.receiveError = err
}

func (b *baseBlockEventsTest) SetResponses(responses ...*peer.DeliverResponse) {
	b.responses = append(b.responses, responses...)
}

func (b *baseBlockEventsTest) Requests() <-chan *common.Envelope {
	return b.requests
}

func (b *baseBlockEventsTest) CallOptions() <-chan []grpc.CallOption {
	return b.callOptions
}

func (b *baseBlockEventsTest) deliverOptions() []newStreamFunction {
	var result []newStreamFunction
	if b.connectError != nil {
		result = append(result, WithNewStreamError(b.connectError))
	} else {
		mockStream := NewMockClientStream(b.t)
		b.callOptions = make(chan []grpc.CallOption, 1)
		result = append(result, CaptureNewStreamOptions(b.callOptions), WithNewStreamResult(mockStream))

		b.requests = make(chan *common.Envelope, 1)
		ExpectSendMsg(mockStream, CaptureSendMsg(b.requests))
		mockStream.EXPECT().CloseSend().Maybe().Return(nil)

		if b.receiveError != nil {
			ExpectRecvMsg(mockStream).Return(b.receiveError)
		} else if len(b.responses) > 0 {
			ExpectRecvMsg(mockStream, WithRecvMsgs(b.responses...))
		} else {
			ExpectRecvMsg(mockStream).Maybe().Return(io.EOF)
		}
	}

	return result
}

func (b *baseBlockEventsTest) newNetwork(connection grpc.ClientConnInterface) *Network {
	options := []ConnectOption{
		WithClientConnection(connection),
	}
	options = append(options, b.gatewayOptions...)
	return AssertNewTestNetwork(b.t, b.networkName, options...)
}

type blockEventsTest struct {
	baseBlockEventsTest
	BlockEvents <-chan *common.Block
}

func NewBlockEventsTest(t *testing.T) *blockEventsTest {
	return &blockEventsTest{
		baseBlockEventsTest: baseBlockEventsTest{
			t: t,
		},
	}
}

func (b *blockEventsTest) Events(ctx context.Context, options ...BlockEventsOption) error {
	mockConnection := NewMockClientConnInterface(b.t)
	ExpectDeliver(mockConnection, b.deliverOptions()...)

	network := b.newNetwork(mockConnection)
	events, err := network.BlockEvents(ctx, options...)

	b.BlockEvents = events

	return err
}

func (b *blockEventsTest) EventsWithCallOptions(ctx context.Context, options ...grpc.CallOption) error {
	mockConnection := NewMockClientConnInterface(b.t)
	ExpectDeliver(mockConnection, b.deliverOptions()...)

	network := b.newNetwork(mockConnection)
	request, err := network.NewBlockEventsRequest()
	require.NoError(b.t, err, "NewBlockEventsRequest")

	events, err := request.Events(ctx, options...)

	b.BlockEvents = events

	return err
}

func (b *blockEventsTest) Receive() (proto.Message, bool) {
	event, ok := <-b.BlockEvents
	return event, ok
}

type filteredBlockEventsTest struct {
	baseBlockEventsTest
	FilteredBlockEvents <-chan *peer.FilteredBlock
}

func NewFilteredBlockEventsTest(t *testing.T) *filteredBlockEventsTest {
	return &filteredBlockEventsTest{
		baseBlockEventsTest: baseBlockEventsTest{
			t: t,
		},
	}
}

func (b *filteredBlockEventsTest) Events(ctx context.Context, options ...BlockEventsOption) error {
	mockConnection := NewMockClientConnInterface(b.t)
	ExpectDeliverFiltered(mockConnection, b.deliverOptions()...)

	network := b.newNetwork(mockConnection)
	events, err := network.FilteredBlockEvents(ctx, options...)

	b.FilteredBlockEvents = events

	return err
}

func (b *filteredBlockEventsTest) EventsWithCallOptions(ctx context.Context, options ...grpc.CallOption) error {
	mockConnection := NewMockClientConnInterface(b.t)
	ExpectDeliverFiltered(mockConnection, b.deliverOptions()...)

	network := b.newNetwork(mockConnection)
	request, err := network.NewFilteredBlockEventsRequest()
	require.NoError(b.t, err, "NewFilteredBlockEventsRequest")

	events, err := request.Events(ctx, options...)

	b.FilteredBlockEvents = events

	return err
}

func (b *filteredBlockEventsTest) Receive() (proto.Message, bool) {
	event, ok := <-b.FilteredBlockEvents
	return event, ok
}

type blockAndPrivateDataEventsTest struct {
	baseBlockEventsTest
	BlocksAndPrivateData <-chan *peer.BlockAndPrivateData
}

func NewBlockAndPrivateDataEventsTest(t *testing.T) *blockAndPrivateDataEventsTest {
	return &blockAndPrivateDataEventsTest{
		baseBlockEventsTest: baseBlockEventsTest{
			t: t,
		},
	}
}

func (b *blockAndPrivateDataEventsTest) Events(ctx context.Context, options ...BlockEventsOption) error {
	mockConnection := NewMockClientConnInterface(b.t)
	ExpectDeliverWithPrivateData(mockConnection, b.deliverOptions()...)

	network := b.newNetwork(mockConnection)
	events, err := network.BlockAndPrivateDataEvents(ctx, options...)

	b.BlocksAndPrivateData = events

	return err
}

func (b *blockAndPrivateDataEventsTest) EventsWithCallOptions(ctx context.Context, options ...grpc.CallOption) error {
	mockConnection := NewMockClientConnInterface(b.t)
	ExpectDeliverWithPrivateData(mockConnection, b.deliverOptions()...)

	network := b.newNetwork(mockConnection)
	request, err := network.NewBlockAndPrivateDataEventsRequest()
	require.NoError(b.t, err, "NewBlockAndPrivateDataEventsRequest")

	events, err := request.Events(ctx, options...)

	b.BlocksAndPrivateData = events

	return err
}

func (b *blockAndPrivateDataEventsTest) Receive() (proto.Message, bool) {
	event, ok := <-b.BlocksAndPrivateData
	return event, ok
}
