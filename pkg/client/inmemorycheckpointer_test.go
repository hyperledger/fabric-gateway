/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInMemoryCheckpointer(t *testing.T) {
	t.Run("Initializes default checkpointer state when no checkpointer already exist", func(t *testing.T) {
		checkpointer := new(InMemoryCheckpointer)

		blockNumber := checkpointer.BlockNumber()

		require.Equal(t, uint64(0), blockNumber)
		require.Equal(t, "", checkpointer.TransactionID())
	})

	t.Run("Checkpointing a block gives next block number & empty transaction Id", func(t *testing.T) {
		blockNumber := uint64(101)
		checkpointer := new(InMemoryCheckpointer)

		checkpointer.CheckpointBlock(blockNumber)

		require.Equal(t, blockNumber+1, checkpointer.BlockNumber())
		require.Equal(t, "", checkpointer.TransactionID())
	})

	t.Run("Checkpointing a transaction gives valid transaction Id and blocknumber ", func(t *testing.T) {
		blockNumber := uint64(101)
		checkpointer := new(InMemoryCheckpointer)

		checkpointer.CheckpointTransaction(blockNumber, "txn1")

		require.Equal(t, blockNumber, checkpointer.BlockNumber())
		require.Equal(t, "txn1", checkpointer.TransactionID())
	})

}
