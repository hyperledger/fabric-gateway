/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

type InMemoryCheckpointer struct {
	blockNumber   uint64
	transactionID string
}

func (c *InMemoryCheckpointer) CheckpointBlock(blockNumber uint64) error {
	c.blockNumber = blockNumber + 1
	c.transactionID = ""
	return nil
}

func (c *InMemoryCheckpointer) CheckpointTransaction(blockNumber uint64, transactionID string) error {
	c.blockNumber = blockNumber
	c.transactionID = transactionID
	return nil
}

func (c *InMemoryCheckpointer) CheckpointChaincodeEvent(event *ChaincodeEvent) error {
	err := c.CheckpointTransaction(event.BlockNumber, event.TransactionID)
	return err
}

func (c *InMemoryCheckpointer) BlockNumber() uint64 {
	return c.blockNumber
}

func (c *InMemoryCheckpointer) TransactionID() string {
	return c.transactionID
}
