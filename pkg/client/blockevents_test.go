// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"testing"
	"testing/synctest"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/require"
)

func TestBlockEvents(t *testing.T) {
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

		var responses []*peer.DeliverResponse
		for _, block := range expected {
			responses = append(responses, &peer.DeliverResponse{
				Type: &peer.DeliverResponse_Block{
					Block: block,
				},
			})
		}

		tester := NewBlockEventsTest(t)
		tester.SetResponses(responses...)

		ctx := t.Context()

		err := tester.Events(ctx)
		require.NoError(t, err)

		for _, event := range expected {
			actual := <-tester.BlockEvents
			AssertProtoEqual(t, event, actual)
		}
	})

	t.Run("Closes event channel on non-block message", func(t *testing.T) {
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

		tester := NewBlockEventsTest(t)
		tester.SetResponses(responses...)

		ctx := t.Context()

		err := tester.Events(ctx)
		require.NoError(t, err)

		expected := []*common.Block{
			block,
			nil,
			nil,
		}
		for _, event := range expected {
			actual := <-tester.BlockEvents
			AssertProtoEqual(t, event, actual)
		}
	})

	t.Run("Stops sending events after context is cancelled", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var responses []*peer.DeliverResponse
			for blockNumber := range uint64(2) {
				responses = append(responses, &peer.DeliverResponse{
					Type: &peer.DeliverResponse_Block{
						Block: &common.Block{
							Header: &common.BlockHeader{
								Number: blockNumber,
							},
						},
					},
				})
			}

			tester := NewBlockEventsTest(t)
			tester.SetResponses(responses...)

			ctx, cancel := context.WithCancel(t.Context())

			err := tester.Events(ctx)
			require.NoError(t, err)

			<-tester.BlockEvents
			cancel()

			// Unread events must not block the sending goroutine after the context is cancelled.
			synctest.Wait()
		})
	})
}
