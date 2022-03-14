/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package checkpoint

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInMemoryCheckpointer(t *testing.T) {
	t.Run("Initializes default checkpointer state when no checkpointer already exist", func(t *testing.T) {

		checkPointerInstance :=  InMemory()
		require.EqualValues(t, uint64(0), checkPointerInstance.GetBlockNumber())
		require.EqualValues(t, "", checkPointerInstance.GetTransactionID())

	});
	t.Run("Checkpointing only a block number in a fresh checkpointer gives block number & blank transaction Id", func(t *testing.T) {

		blockNumber := uint64(101)
		checkPointerInstance :=  InMemory()
		checkPointerInstance.Checkpoint(blockNumber);
		require.EqualValues(t, blockNumber, checkPointerInstance.GetBlockNumber())
		require.EqualValues(t, "", checkPointerInstance.GetTransactionID())
	});
	t.Run("Checkpointing same block number and new transaction in used checkpointer gives block number and expected transaction", func(t *testing.T) {

		blockNumber := uint64(101)
		checkPointerInstance :=  InMemory()
	    checkPointerInstance.Checkpoint(blockNumber);
		checkPointerInstance.Checkpoint(blockNumber, "txn1");
		require.EqualValues(t, "txn1", checkPointerInstance.GetTransactionID())

	});
	t.Run("Checkpointing block and transaction in a fresh checkpointer, gives block number and transaction", func(t *testing.T) {

		blockNumber := uint64(101);
		checkPointerInstance := InMemory()
		checkPointerInstance.Checkpoint(blockNumber, "txn1");
		require.EqualValues(t, blockNumber, checkPointerInstance.GetBlockNumber())
		require.EqualValues(t, "txn1", checkPointerInstance.GetTransactionID())

	});
	t.Run("Checkpointing only a new block in used checkpointer gives new block number and blank transaction", func(t *testing.T) {

		blockNumber1 := uint64(101);
		blockNumber2 := uint64(102);
		checkPointerInstance :=  InMemory();
		checkPointerInstance.Checkpoint(blockNumber1, "txn1");
		checkPointerInstance.Checkpoint(blockNumber2);
		require.EqualValues(t, blockNumber2, checkPointerInstance.GetBlockNumber())
		require.EqualValues(t, "", checkPointerInstance.GetTransactionID())

	});
}