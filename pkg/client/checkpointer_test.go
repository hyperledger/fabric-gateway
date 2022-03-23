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

var path = "checkpointer_test_file.json"

type checkpoint interface {
	CheckpointBlock(uint64) error
	CheckpointTransaction(uint64, string) error
	CheckpointChaincodeEvent(*ChaincodeEvent) error
}

type Checkpoint interface {
	checkpoint
	Checkpointer
}

type testCase struct {
	description     string
	after           func()
	newCheckpointer func() Checkpoint
}

func CheckFileExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func setupSuite(t *testing.T) func(t *testing.T) {
	if !CheckFileExist(path) {
		file, err := os.Create(path)
		if isError(err) {
			panic(err)
		}
		defer file.Close()
	}
	return func(t *testing.T) {
		if CheckFileExist(path) {
			err := os.Remove(path)
			if isError(err) {
				panic(err)
			}
		}
	}
}

func assertState(t *testing.T, checkpointer Checkpoint, blocknumber uint64, transactionID string) {
	require.Equal(t, blocknumber, checkpointer.BlockNumber())
	require.Equal(t, transactionID, checkpointer.TransactionID())
}

func TestCheckpointer(t *testing.T) {

	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	testCases := []testCase{
		{
			description: "In-memory",
			after:       func() {},
			newCheckpointer: func() Checkpoint {
				return new(InMemoryCheckpointer)
			},
		},
		{
			description: "File",
			after: func() {
				var err = os.Remove(path)
				if isError(err) {
					return
				}
			},
			newCheckpointer: func() Checkpoint {
				fileCheckpointer, _ := NewFileCheckpointer(path)
				return fileCheckpointer
			},
		},
	}

	for _, tc := range testCases {

		t.Run(tc.description+" :Initializes default checkpointer state when no checkpointer already exist", func(t *testing.T) {
			checkpointer := tc.newCheckpointer()

			assertState(t, checkpointer, uint64(0), "")

			tc.after()
		})

		t.Run(tc.description+" :Checkpointing a block gives next block number & empty transaction Id", func(t *testing.T) {

			blockNumber := uint64(101)
			checkpointer := tc.newCheckpointer()

			checkpointer.CheckpointBlock(blockNumber)

			assertState(t, checkpointer, blockNumber+1, "")

			tc.after()
		})

		t.Run(tc.description+" :Checkpointing a transaction gives valid transaction Id and blocknumber ", func(t *testing.T) {

			blockNumber := uint64(101)
			checkpointer := tc.newCheckpointer()

			checkpointer.CheckpointTransaction(blockNumber, "txn1")

			assertState(t, checkpointer, blockNumber, "txn1")

			tc.after()
		})

		t.Run(tc.description+" :Checkpointing an event gives valid transaction Id and blocknumber ", func(t *testing.T) {

			checkpointer := tc.newCheckpointer()
			event := &ChaincodeEvent{
				BlockNumber:   uint64(101),
				TransactionID: "txn1",
				ChaincodeName: "Chaincode",
				EventName:     "event1",
				Payload:       []byte("payload"),
			}

			checkpointer.CheckpointChaincodeEvent(event)
			assertState(t, checkpointer, event.BlockNumber, event.TransactionID)

			tc.after()
		})
	}
}
