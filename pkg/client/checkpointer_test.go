/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type testCase struct {
	description     string
	before          func(t *testing.T, fileName string) string
	newCheckpointer func(t *testing.T, filePath string) Checkpointer
}

func setupSuite(t *testing.T) string {
	dname, err := os.MkdirTemp("", "sample")
	require.NoError(t, err)
	return dname
}

func teardownSuite(t *testing.T, dirName string) {
	err := os.RemoveAll(dirName)
	require.NoError(t, err)
}

func assertState(t *testing.T, checkpointer Checkpoint, blocknumber uint64, transactionID string) {
	require.Equal(t, blocknumber, checkpointer.BlockNumber())
	require.Equal(t, transactionID, checkpointer.TransactionID())
}

func TestCheckpointer(t *testing.T) {

	dirName := setupSuite(t)
	defer teardownSuite(t, dirName)

	testCases := []testCase{
		{
			description: "In-memory",
			before: func(t *testing.T, fileName string) string {
				return ""
			},
			newCheckpointer: func(t *testing.T, filePath string) Checkpointer {
				inmemorCheckpointer := new(InMemoryCheckpointer)
				return inmemorCheckpointer
			},
		},
		{
			description: "File",
			before: func(t *testing.T, fileName string) string {
				filePath := filepath.Join(dirName, fileName)
				err := os.WriteFile(filePath, nil, 0666)
				require.NoError(t, err)
				return filePath
			},
			newCheckpointer: func(t *testing.T, filePath string) Checkpointer {
				fileCheckpointer, err := NewFileCheckpointer(filePath)
				require.NoError(t, err)
				return fileCheckpointer
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.description, func(t *testing.T) {

			t.Run("Initializes default checkpointer state when no checkpointer already exist", func(t *testing.T) {
				filePath := tc.before(t, "file1.json")

				checkpointer := tc.newCheckpointer(t, filePath)

				assertState(t, checkpointer, uint64(0), "")
			})

			t.Run("Checkpointing a block gives next block number & empty transaction Id", func(t *testing.T) {
				blockNumber := uint64(101)
				filePath := tc.before(t, "file2.json")

				checkpointer := tc.newCheckpointer(t, filePath)
				checkpointer.CheckpointBlock(blockNumber)

				assertState(t, checkpointer, blockNumber+1, "")
			})

			t.Run("Checkpointing a transaction gives valid transaction Id and blocknumber ", func(t *testing.T) {
				blockNumber := uint64(101)
				filePath := tc.before(t, "file3.json")

				checkpointer := tc.newCheckpointer(t, filePath)
				checkpointer.CheckpointTransaction(blockNumber, "txn1")

				assertState(t, checkpointer, blockNumber, "txn1")
			})

			t.Run("Checkpointing an event gives valid transaction Id and blocknumber ", func(t *testing.T) {
				filePath := tc.before(t, "file4.json")
				event := &ChaincodeEvent{
					BlockNumber:   uint64(101),
					TransactionID: "txn1",
					ChaincodeName: "Chaincode",
					EventName:     "event1",
					Payload:       []byte("payload"),
				}

				checkpointer := tc.newCheckpointer(t, filePath)
				checkpointer.CheckpointChaincodeEvent(event)

				assertState(t, checkpointer, event.BlockNumber, event.TransactionID)

			})
		})
	}
}
