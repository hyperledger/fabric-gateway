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

const checkpointerFilePath = "file_checkpointer_test_file.json"

func after(t *testing.T) {
	if CheckFileExist(checkpointerFilePath) {
		err := os.Remove(checkpointerFilePath)
		require.NoError(t, err)
	}

}

func TestFileCheckpointer(t *testing.T) {
	defer after(t)

	t.Run("state is persisted", func(t *testing.T) {
		actualFileCheckpointer, err1 := NewFileCheckpointer(checkpointerFilePath)
		actualFileCheckpointer.CheckpointTransaction(uint64(1), "txn")

		expectedFileCheckpointer, err2 := NewFileCheckpointer(checkpointerFilePath)

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.Equal(t, actualFileCheckpointer.BlockNumber(), expectedFileCheckpointer.BlockNumber())
		require.Equal(t, actualFileCheckpointer.TransactionID(), expectedFileCheckpointer.TransactionID())

		after(t)
	})

	t.Run("block number zero is persisted correctly", func(t *testing.T) {
		blockNumber := uint64(0)
		fileCheckpointer, err := NewFileCheckpointer(checkpointerFilePath)

		fileCheckpointer.CheckpointBlock(uint64(0))

		require.NoError(t, err)
		require.Equal(t, blockNumber+uint64(1), fileCheckpointer.BlockNumber())
		require.Equal(t, "", fileCheckpointer.TransactionID())

		after(t)
	})

	t.Run("throws on reading invalid blockNumber type from the file", func(t *testing.T) {

		err := os.WriteFile(checkpointerFilePath, []byte(`{"blockNumber":false}`), 0600)

		_, err1 := NewFileCheckpointer(checkpointerFilePath)

		require.NoError(t, err)
		require.Containsf(t, err1.Error(), "cannot unmarshal", "Received unexpected error message")

		after(t)
	})

	t.Run("throws on unwritable file location", func(t *testing.T) {

		badPath := "MISSING_DIRECTORY/test_file.json"

		_, err := NewFileCheckpointer(badPath)

		require.EqualErrorf(t, err, "open MISSING_DIRECTORY/test_file.json: no such file or directory", "Error should be: %v, got: %v", "open MISSING_DIRECTORY/checkpointer.go: no such file or directory", err)
	})
}
