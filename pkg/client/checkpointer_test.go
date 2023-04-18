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

	assertState := func(t *testing.T, checkpoint Checkpoint, blockNumber uint64, transactionID string) {
		require.Equal(t, blockNumber, checkpoint.BlockNumber(), "BlockNumber")
		require.Equal(t, transactionID, checkpoint.TransactionID(), "TransactionID")
	}

	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	for testName, testCase := range map[string]struct {
		newCheckpointer func(*testing.T) Checkpointer
	}{
		"In-memory": {
			newCheckpointer: func(t *testing.T) Checkpointer {
				return &InMemoryAdapter{
					InMemoryCheckpointer{},
				}
			},
		},
		"File": {
			newCheckpointer: func(t *testing.T) Checkpointer {
				fileName := NonExistentFileName(t, tempDir)
				checkpointer, err := NewFileCheckpointer(fileName)
				require.NoError(t, err)

				return checkpointer
			},
		},
	} {
		t.Run(testName, func(t *testing.T) {
			t.Run("Initial checkpointer state", func(t *testing.T) {
				checkpointer := testCase.newCheckpointer(t)
				defer checkpointer.Close()

				assertState(t, checkpointer, uint64(0), "")
			})

			t.Run("CheckpointBlock() sets next block number and empty transaction ID", func(t *testing.T) {
				blockNumber := uint64(101)
				checkpointer := testCase.newCheckpointer(t)
				defer checkpointer.Close()

				err = checkpointer.CheckpointBlock(blockNumber)
				require.NoError(t, err)

				assertState(t, checkpointer, blockNumber+1, "")
			})

			t.Run("CheckpointTransaction() sets block number and transaction ID", func(t *testing.T) {
				blockNumber := uint64(101)
				checkpointer := testCase.newCheckpointer(t)
				defer checkpointer.Close()

				err = checkpointer.CheckpointTransaction(blockNumber, "txn1")
				require.NoError(t, err)

				assertState(t, checkpointer, blockNumber, "txn1")
			})

			t.Run("CheckpointiEvent() sets block number and transaction ID from event", func(t *testing.T) {
				event := &ChaincodeEvent{
					BlockNumber:   uint64(101),
					TransactionID: "txn1",
				}
				checkpointer := testCase.newCheckpointer(t)
				defer checkpointer.Close()

				err = checkpointer.CheckpointChaincodeEvent(event)
				require.NoError(t, err)

				assertState(t, checkpointer, event.BlockNumber, event.TransactionID)
			})
		})
	}
}

type InMemoryAdapter struct {
	InMemoryCheckpointer
}

func (adapter *InMemoryAdapter) CheckpointBlock(blockNumber uint64) error {
	adapter.InMemoryCheckpointer.CheckpointBlock(blockNumber)
	return nil
}

func (adapter *InMemoryAdapter) CheckpointTransaction(blockNumber uint64, transactionID string) error {
	adapter.InMemoryCheckpointer.CheckpointTransaction(blockNumber, transactionID)
	return nil
}

func (adapter *InMemoryAdapter) CheckpointChaincodeEvent(event *ChaincodeEvent) error {
	adapter.InMemoryCheckpointer.CheckpointChaincodeEvent(event)
	return nil
}

func (adapter *InMemoryAdapter) Close() error {
	return nil
}
