/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package checkpoint

    type BlockNumber struct {
        Value uint64
        Exist bool
    }
	type InMemoryCheckpointer struct {
        blockNumber *BlockNumber
        transactionID string
	}

    func (c *InMemoryCheckpointer) Checkpoint (blockNumber func() ( uint64, bool), transactionID ...string) {
        if blockNumber, exist := blockNumber(); exist {
            c.blockNumber.Exist = true;
            if (blockNumber != c.blockNumber.Value ) {
                c.blockNumber.Value = blockNumber;
                c.transactionID = "";
            }
        }
            if len(transactionID) > 0 {
                c.transactionID =  transactionID[0]
            }
    }

    func (c *InMemoryCheckpointer) GetBlockNumber() *BlockNumber {
        return c.blockNumber;
    }

    func (c *InMemoryCheckpointer) GetTransactionID() string {
        return c.transactionID;
    }