/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func NonExistentFileName(t *testing.T, dir string) string {
	file, err := os.CreateTemp(dir, "checkpointer")
	require.NoError(t, err, "CreateTemp")

	fileName := file.Name()

	require.NoError(t, file.Close(), "Close")
	require.NoError(t, os.Remove(fileName), "Remove")

	return fileName
}

func TestCheckpointer(t *testing.T) {
	type Checkpointer interface {
		CheckpointBlock(uint64) error
		CheckpointTransaction(uint64, string) error
		CheckpointChaincodeEvent(*ChaincodeEvent) error
		Close() error

		Checkpoint
	}

	type TestCase struct {
		description     string
		newCheckpointer func(*testing.T) Checkpointer
	}

	assertState := func(t *testing.T, checkpoint Checkpoint, blockNumber uint64, transactionID string) {
		require.Equal(t, blockNumber, checkpoint.BlockNumber(), "BlockNumber")
		require.Equal(t, transactionID, checkpoint.TransactionID(), "TransactionID")
	}

	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testCases := []TestCase{
		{
			description: "In-memory",
			newCheckpointer: func(t *testing.T) Checkpointer {
				return &InMemoryAdapter{
					checkpointer: &InMemoryCheckpointer{},
				}
			},
		},
		{
			description: "File",
			newCheckpointer: func(t *testing.T) Checkpointer {
				fileName := NonExistentFileName(t, tempDir)
				checkpointer, err := NewFileCheckpointer(fileName)
				require.NoError(t, err)

				return checkpointer
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			t.Run("Initial checkpointer state", func(t *testing.T) {
				checkpointer := tc.newCheckpointer(t)
				defer checkpointer.Close()

				assertState(t, checkpointer, uint64(0), "")
			})

			t.Run("CheckpointBlock() sets next block number and empty transaction ID", func(t *testing.T) {
				blockNumber := uint64(101)
				checkpointer := tc.newCheckpointer(t)
				defer checkpointer.Close()

				checkpointer.CheckpointBlock(blockNumber)

				assertState(t, checkpointer, blockNumber+1, "")
			})

			t.Run("CheckpointTransaction() sets block number and transaction ID", func(t *testing.T) {
				blockNumber := uint64(101)
				checkpointer := tc.newCheckpointer(t)
				defer checkpointer.Close()

				checkpointer.CheckpointTransaction(blockNumber, "txn1")

				assertState(t, checkpointer, blockNumber, "txn1")
			})

			t.Run("CheckpointiEvent() sets block number and transaction ID from event", func(t *testing.T) {
				event := &ChaincodeEvent{
					BlockNumber:   uint64(101),
					TransactionID: "txn1",
				}
				checkpointer := tc.newCheckpointer(t)
				defer checkpointer.Close()

				checkpointer.CheckpointChaincodeEvent(event)

				assertState(t, checkpointer, event.BlockNumber, event.TransactionID)
			})
		})
	}
}

type InMemoryAdapter struct {
	checkpointer *InMemoryCheckpointer
}

func (adapter *InMemoryAdapter) BlockNumber() uint64 {
	return adapter.checkpointer.BlockNumber()
}

func (adapter *InMemoryAdapter) TransactionID() string {
	return adapter.checkpointer.TransactionID()
}

func (adapter *InMemoryAdapter) CheckpointBlock(blockNumber uint64) error {
	adapter.checkpointer.CheckpointBlock(blockNumber)
	return nil
}

func (adapter *InMemoryAdapter) CheckpointTransaction(blockNumber uint64, transactionID string) error {
	adapter.checkpointer.CheckpointTransaction(blockNumber, transactionID)
	return nil
}

func (adapter *InMemoryAdapter) CheckpointChaincodeEvent(event *ChaincodeEvent) error {
	adapter.checkpointer.CheckpointChaincodeEvent(event)
	return nil
}

func (adapter *InMemoryAdapter) Close() error {
	return nil
}
