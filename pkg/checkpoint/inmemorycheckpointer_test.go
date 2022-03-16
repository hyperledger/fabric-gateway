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
		require.EqualValues(t, uint64(0), checkPointerInstance.GetBlockNumber().Value)
		require.EqualValues(t, false, checkPointerInstance.GetBlockNumber().Exist)
		require.EqualValues(t, "", checkPointerInstance.GetTransactionID())

	});
	t.Run("Checkpointing only a block number in a fresh checkpointer gives block number & blank transaction Id", func(t *testing.T) {

		blockNumber := uint64(101)
		checkPointerInstance :=  InMemory()
		checkPointerInstance.Checkpoint(func() (uint64, bool) {return blockNumber ,true});
		require.EqualValues(t, blockNumber, checkPointerInstance.GetBlockNumber().Value)
		require.EqualValues(t, true, checkPointerInstance.GetBlockNumber().Exist)
		require.EqualValues(t, "", checkPointerInstance.GetTransactionID())
	});

	t.Run("Checkpointing zero block number in a fresh checkpointer gives valid block number zero", func(t *testing.T) {

		blockNumber := uint64(0)
		checkPointerInstance :=  InMemory()
		checkPointerInstance.Checkpoint(func() (uint64, bool) {return blockNumber ,true});
		require.EqualValues(t, blockNumber, checkPointerInstance.GetBlockNumber().Value)
		require.EqualValues(t, true, checkPointerInstance.GetBlockNumber().Exist)
		require.EqualValues(t, "", checkPointerInstance.GetTransactionID())
	});

	t.Run("Checkpointing same block number and new transaction in used checkpointer gives block number and expected transaction", func(t *testing.T) {

		blockNumber := uint64(101)
		setBlockNumber := func() (uint64, bool) {return blockNumber ,true}
		checkPointerInstance :=  InMemory()
	    checkPointerInstance.Checkpoint(setBlockNumber);
		checkPointerInstance.Checkpoint(setBlockNumber, "txn1");
		require.EqualValues(t, "txn1", checkPointerInstance.GetTransactionID())

	});
	t.Run("Checkpointing block and transaction in a fresh checkpointer, gives block number and transaction", func(t *testing.T) {

		blockNumber := uint64(101);
		checkPointerInstance := InMemory()
		checkPointerInstance.Checkpoint(func() (uint64, bool) {return blockNumber ,true}, "txn1");
		require.EqualValues(t, blockNumber, checkPointerInstance.GetBlockNumber().Value)
		require.EqualValues(t, true, checkPointerInstance.GetBlockNumber().Exist)
		require.EqualValues(t, "txn1", checkPointerInstance.GetTransactionID())

	});
	t.Run("Checkpointing only a new block in used checkpointer gives new block number and blank transaction", func(t *testing.T) {

		blockNumber1 := uint64(101);
		blockNumber2 := uint64(102);
		setBlockNumber1 := func() (uint64, bool) {return blockNumber1 ,true}
		setBlockNumber2 := func() (uint64, bool) {return blockNumber2 ,true}
		checkPointerInstance :=  InMemory();
		checkPointerInstance.Checkpoint(setBlockNumber1, "txn1");
		checkPointerInstance.Checkpoint(setBlockNumber2);
		require.EqualValues(t, blockNumber2, checkPointerInstance.GetBlockNumber().Value)
		require.EqualValues(t, true, checkPointerInstance.GetBlockNumber().Exist)
		require.EqualValues(t, "", checkPointerInstance.GetTransactionID())

	});
}