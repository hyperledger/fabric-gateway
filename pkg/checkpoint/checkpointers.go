/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package checkpoint

type Checkpointer interface {
    Checkpoint  (uint64,  ...string)
    GetBlockNumber() uint64
    GetTransactionID() string
}

func InMemory() Checkpointer {
	return &InMemoryCheckpointer{
		blockNumber: uint64(0),
		transactionID: "",
	}
}