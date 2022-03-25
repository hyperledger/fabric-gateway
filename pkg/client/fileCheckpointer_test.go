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

func beforeAll(t *testing.T) string {
	dname, err := os.MkdirTemp("", "temp")
	require.NoError(t, err)
	return dname
}

func afterAll(t *testing.T, dirName string) {
	err := os.RemoveAll(dirName)
	require.NoError(t, err)
}

func beforeEach(t *testing.T, dirName string, fileName string) string {
	filePath := filepath.Join(dirName, fileName)
	err := os.WriteFile(filePath, nil, 0666)
	require.NoError(t, err)
	return filePath
}

func TestFileCheckpointer(t *testing.T) {
	dirName := beforeAll(t)
	defer afterAll(t, dirName)

	t.Run("state is persisted", func(t *testing.T) {
		checkpointerFilePath := beforeEach(t, dirName, "file1.json")

		actualFileCheckpointer, err1 := NewFileCheckpointer(checkpointerFilePath)
		actualFileCheckpointer.CheckpointTransaction(uint64(1), "txn")

		expectedFileCheckpointer, err2 := NewFileCheckpointer(checkpointerFilePath)

		require.NoError(t, err1)
		require.NoError(t, err2)
		require.Equal(t, actualFileCheckpointer.BlockNumber(), expectedFileCheckpointer.BlockNumber())
		require.Equal(t, actualFileCheckpointer.TransactionID(), expectedFileCheckpointer.TransactionID())
	})

	t.Run("block number zero is persisted correctly", func(t *testing.T) {
		checkpointerFilePath := beforeEach(t, dirName, "file2.json")
		blockNumber := uint64(0)
		fileCheckpointer, err := NewFileCheckpointer(checkpointerFilePath)

		fileCheckpointer.CheckpointBlock(uint64(0))

		require.NoError(t, err)
		require.Equal(t, blockNumber+uint64(1), fileCheckpointer.BlockNumber())
		require.Equal(t, "", fileCheckpointer.TransactionID())
	})

	t.Run("throws on reading invalid blockNumber type from the file", func(t *testing.T) {
		checkpointerFilePath := beforeEach(t, dirName, "file3.json")
		err := os.WriteFile(checkpointerFilePath, []byte(`{"blockNumber":false}`), 0600)

		_, err1 := NewFileCheckpointer(checkpointerFilePath)

		require.NoError(t, err)
		require.Containsf(t, err1.Error(), "cannot unmarshal", "Received unexpected error message")

	})

	t.Run("throws on unwritable file location", func(t *testing.T) {

		badPath := "MISSING_DIRECTORY/test_file.json"

		_, err := NewFileCheckpointer(badPath)

		require.Errorf(t, err, "Error not received")
	})
}
