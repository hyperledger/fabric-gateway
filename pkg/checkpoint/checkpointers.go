/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package checkpoint

type Checkpointer interface {
    Checkpoint  (func() ( uint64, bool),  ...string)
    GetBlockNumber() *BlockNumber
    GetTransactionID() string
}
type CheckpointData struct {
	StartBlock *BlockNumber
	AfterTransactionID string
}

func InMemory() Checkpointer {
	return &InMemoryCheckpointer{
		blockNumber: &BlockNumber{
			Value: uint64(0),
			Exist: false,
		},
		transactionID: "",
	}
}