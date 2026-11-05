// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"testing"
	"testing/synctest"

	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"github.com/hyperledger/fabric-protos-go-apiv2/ledger/rwset"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/require"
)

func TestBlockAndPrivateDataEvents(t *testing.T) {
	t.Run("Receives events", func(t *testing.T) {
		expected := []*peer.BlockAndPrivateData{
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

		var responses []*peer.DeliverResponse
		for _, block := range expected {
			responses = append(responses, &peer.DeliverResponse{
				Type: &peer.DeliverResponse_BlockAndPrivateData{
					BlockAndPrivateData: block,
				},
			})
		}

		tester := NewBlockAndPrivateDataEventsTest(t)
		tester.SetResponses(responses...)

		ctx := t.Context()

		err := tester.Events(ctx)
		require.NoError(t, err)

		for _, event := range expected {
			actual := <-tester.BlocksAndPrivateData
			AssertProtoEqual(t, event, actual)
		}
	})

	t.Run("Closes event channel on non-block message", func(t *testing.T) {
		block := &peer.BlockAndPrivateData{
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
					BlockAndPrivateData: block,
				},
			},
			{
				Type: &peer.DeliverResponse_Status{
					Status: common.Status_SERVICE_UNAVAILABLE,
				},
			},
			{
				Type: &peer.DeliverResponse_BlockAndPrivateData{
					BlockAndPrivateData: block,
				},
			},
		}

		tester := NewBlockAndPrivateDataEventsTest(t)
		tester.SetResponses(responses...)

		ctx := t.Context()

		err := tester.Events(ctx)
		require.NoError(t, err)

		expected := []*peer.BlockAndPrivateData{
			block,
			nil,
			nil,
		}
		for _, event := range expected {
			actual := <-tester.BlocksAndPrivateData
			AssertProtoEqual(t, event, actual)
		}
	})

	t.Run("Stops sending events after context is cancelled", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var responses []*peer.DeliverResponse
			for blockNumber := range uint64(2) {
				responses = append(responses, &peer.DeliverResponse{
					Type: &peer.DeliverResponse_BlockAndPrivateData{
						BlockAndPrivateData: &peer.BlockAndPrivateData{
							Block: &common.Block{
								Header: &common.BlockHeader{
									Number: blockNumber,
								},
							},
						},
					},
				})
			}

			tester := NewBlockAndPrivateDataEventsTest(t)
			tester.SetResponses(responses...)

			ctx, cancel := context.WithCancel(t.Context())

			err := tester.Events(ctx)
			require.NoError(t, err)

			<-tester.BlocksAndPrivateData
			cancel()

			// Unread events must not block the sending goroutine after the context is cancelled.
			synctest.Wait()
		})
	})
}
