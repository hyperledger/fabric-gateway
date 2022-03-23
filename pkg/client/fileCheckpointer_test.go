/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const checkpointerFilePath = "file_checkpointer_test_file.json"

func after(t *testing.T) {
	fmt.Println("CheckFileExist", CheckFileExist(checkpointerFilePath))
	if CheckFileExist(checkpointerFilePath) {
		_ = os.Remove(checkpointerFilePath)
	}

}

func TestFileCheckpointer(t *testing.T) {
	t.Run("state is persisted", func(t *testing.T) {
		actualFileCheckpointer, _ := NewFileCheckpointer(checkpointerFilePath)
		actualFileCheckpointer.CheckpointTransaction(uint64(1), "txn")

		expectedFileCheckpointer, _ := NewFileCheckpointer(checkpointerFilePath)

		require.Equal(t, actualFileCheckpointer.BlockNumber(), expectedFileCheckpointer.BlockNumber())
		require.Equal(t, actualFileCheckpointer.TransactionID(), expectedFileCheckpointer.TransactionID())

		after(t)
	})

	t.Run("block number zero is persisted correctly", func(t *testing.T) {
		blockNumber := uint64(0)
		fileCheckpointer, _ := NewFileCheckpointer(checkpointerFilePath)

		fileCheckpointer.CheckpointBlock(uint64(0))

		require.Equal(t, blockNumber+uint64(1), fileCheckpointer.BlockNumber())
		require.Equal(t, "", fileCheckpointer.TransactionID())

		after(t)
	})

	t.Run("throws on unwritable file location", func(t *testing.T) {

		badPath := "MISSING_DIRECTORY/test_file.json"

		_, err := NewFileCheckpointer(badPath)

		require.EqualErrorf(t, err, "open MISSING_DIRECTORY/test_file.json: no such file or directory", "Error should be: %v, got: %v", "open MISSINGDIRECTORY/checkpointer.go: no such file or directory", err)

	})

}
