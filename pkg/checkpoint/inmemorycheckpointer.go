/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package checkpoint

type InMemoryCheckpointer struct {
	blockNumber uint64
	transactionID string
}

func (c *InMemoryCheckpointer) CheckpointBlock(blockNumber uint64) {
	c.blockNumber = blockNumber + 1
	c.transactionID = ""
}

func (c *InMemoryCheckpointer) CheckpointTransaction(blockNumber uint64 , transactionID string) {
	c.blockNumber = blockNumber
	c.transactionID = transactionID
}

func (c *InMemoryCheckpointer) BlockNumber() uint64 {
	return c.blockNumber
}

func (c *InMemoryCheckpointer) TransactionID() string {
	return c.transactionID
}