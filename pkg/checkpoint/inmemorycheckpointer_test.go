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

		checkPointerInstance :=  new(InMemoryCheckpointer)
		blockNumber := checkPointerInstance.BlockNumber()
		require.EqualValues(t, uint64(0), blockNumber)
		require.EqualValues(t, "", checkPointerInstance.TransactionID())

	});
	t.Run("Checkpointing only a block number in a fresh checkpointer gives block number & blank transaction Id", func(t *testing.T) {

		expectedBlockNumber := uint64(101)
		checkPointerInstance :=  new(InMemoryCheckpointer)
		checkPointerInstance.CheckpointBlockNumber(expectedBlockNumber)
		actualblockNumber := checkPointerInstance.BlockNumber()
		require.EqualValues(t, expectedBlockNumber + 1 , actualblockNumber)
		require.EqualValues(t, "", checkPointerInstance.TransactionID())
	});
	t.Run("Checkpointing only a transaction id in a fresh checkpointer gives valid transaction Id and no blocknumber ", func(t *testing.T) {


		checkPointerInstance :=  new(InMemoryCheckpointer)
		checkPointerInstance.CheckpointTransaction("txn1")
		blockNumber := checkPointerInstance.BlockNumber()
		require.EqualValues(t, uint64(0), blockNumber)
		require.EqualValues(t, "txn1", checkPointerInstance.TransactionID())
	});

	t.Run("Checkpointing zero block number in a fresh checkpointer gives valid block number 1", func(t *testing.T) {

		expectedBlockNumber := uint64(0)
		checkPointerInstance :=  new(InMemoryCheckpointer)
		checkPointerInstance.CheckpointBlockNumber(expectedBlockNumber)
		actualblockNumber := checkPointerInstance.BlockNumber()
		require.EqualValues(t, expectedBlockNumber + 1 , actualblockNumber)
		require.EqualValues(t, "", checkPointerInstance.TransactionID())
	});

	t.Run("Checkpointing  block number and new transaction in used checkpointer unsets the  block number and gives expected transaction", func(t *testing.T) {

		expectedBlockNumber := uint64(101)
		checkPointerInstance :=  new(InMemoryCheckpointer)
		checkPointerInstance.CheckpointBlockNumber(expectedBlockNumber)
		checkPointerInstance.CheckpointTransaction("txn1")
		actualblockNumber := checkPointerInstance.BlockNumber()
		require.EqualValues(t, uint64(0), actualblockNumber)
		require.EqualValues(t, "txn1", checkPointerInstance.TransactionID())

	});

	t.Run("Checkpointing only a new block in used checkpointer gives new block number and blank transaction", func(t *testing.T) {

		blockNumber1 := uint64(101)
		blockNumber2 := uint64(102)

		checkPointerInstance :=  new(InMemoryCheckpointer)
		checkPointerInstance.CheckpointBlockNumber(blockNumber1)
		checkPointerInstance.CheckpointBlockNumber(blockNumber2)
		actualblockNumber := checkPointerInstance.BlockNumber()
		require.EqualValues(t, blockNumber2 + 1, actualblockNumber)
		require.EqualValues(t, "", checkPointerInstance.TransactionID())
	});
}