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

func TestFilteredBlockEvents(t *testing.T) {
	t.Run("Receives events", func(t *testing.T) {
		expected := []*peer.FilteredBlock{
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

		var responses []*peer.DeliverResponse
		for _, block := range expected {
			responses = append(responses, &peer.DeliverResponse{
				Type: &peer.DeliverResponse_FilteredBlock{
					FilteredBlock: block,
				},
			})
		}

		tester := NewFilteredBlockEventsTest(t)
		tester.SetResponses(responses...)

		ctx := t.Context()

		err := tester.Events(ctx)
		require.NoError(t, err)

		for _, event := range expected {
			actual := <-tester.FilteredBlockEvents
			AssertProtoEqual(t, event, actual)
		}
	})

	t.Run("Closes event channel on non-block message", func(t *testing.T) {
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

		tester := NewFilteredBlockEventsTest(t)
		tester.SetResponses(responses...)

		ctx := t.Context()

		err := tester.Events(ctx)
		require.NoError(t, err)

		expected := []*peer.FilteredBlock{
			block,
			nil,
			nil,
		}
		for _, event := range expected {
			actual := <-tester.FilteredBlockEvents
			AssertProtoEqual(t, event, actual)
		}
	})

	t.Run("Stops sending events after context is cancelled", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var responses []*peer.DeliverResponse
			for blockNumber := range uint64(2) {
				responses = append(responses, &peer.DeliverResponse{
					Type: &peer.DeliverResponse_FilteredBlock{
						FilteredBlock: &peer.FilteredBlock{
							Number: blockNumber,
						},
					},
				})
			}

			tester := NewFilteredBlockEventsTest(t)
			tester.SetResponses(responses...)

			ctx, cancel := context.WithCancel(t.Context())

			err := tester.Events(ctx)
			require.NoError(t, err)

			<-tester.FilteredBlockEvents
			cancel()

			// Unread events must not block the sending goroutine after the context is cancelled.
			synctest.Wait()
		})
	})
}
