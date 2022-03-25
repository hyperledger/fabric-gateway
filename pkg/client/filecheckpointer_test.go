/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileCheckpointer(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	newCheckpointer := func(t *testing.T) (*FileCheckpointer, string) {
		fileName := NonExistentFileName(t, tempDir)
		checkpointer, err := NewFileCheckpointer(fileName)
		require.NoError(t, err)

		return checkpointer, fileName
	}

	t.Run("state is persisted", func(t *testing.T) {
		expected, fileName := newCheckpointer(t)
		defer expected.Close()

		expected.CheckpointTransaction(uint64(1), "TRANSACTION_ID")

		actual, err := NewFileCheckpointer(fileName)
		require.NoError(t, err, "NewFileCheckpointer")
		defer actual.Close()

		require.Equal(t, uint64(1), actual.BlockNumber())
		require.Equal(t, "TRANSACTION_ID", actual.TransactionID())
	})

	t.Run("block number zero is persisted correctly", func(t *testing.T) {
		expected, fileName := newCheckpointer(t)
		defer expected.Close()

		actual, err := NewFileCheckpointer(fileName)
		require.NoError(t, err, "NewFileCheckpointer")
		defer actual.Close()

		require.Equal(t, uint64(0), actual.BlockNumber())
		require.Equal(t, "", actual.TransactionID())
	})

	t.Run("error reading invalid checkpoint JSON from file", func(t *testing.T) {
		type BadState struct {
			BlockNumber   string `json:"blockNumber"` // Wrong data type
			TransactionID string `json:"transactionId"`
		}

		fileName := NonExistentFileName(t, tempDir)

		data, err := json.Marshal(&BadState{})
		require.NoError(t, err, "Marshal")
		require.NoError(t, os.WriteFile(fileName, data, 0644), "WriteFile")

		_, err = NewFileCheckpointer(fileName)

		require.Error(t, err)
		require.Contains(t, err.Error(), "blockNumber")
	})

	t.Run("error reading non-JSON file", func(t *testing.T) {
		file, err := os.CreateTemp(tempDir, "test")
		require.NoError(t, err, "CreateTemp")

		_, err = file.WriteString("I AM NOT JSON DATA")
		require.NoError(t, err, "WriteString")
		require.NoError(t, file.Close(), "Close")

		_, err = NewFileCheckpointer(file.Name())

		require.Error(t, err)
	})

	t.Run("error checkpointing to non-writable file location", func(t *testing.T) {
		_, err = NewFileCheckpointer(path.Join(tempDir, "NON_EXISTENT_DIR", "FILE"))

		require.Error(t, err)
	})
}
