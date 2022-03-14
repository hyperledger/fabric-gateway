/*
Copyright 2022 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package checkpoint

	type InMemoryCheckpointer struct {
        blockNumber uint64
        transactionID string
	}

    func (c *InMemoryCheckpointer) Checkpoint (blockNumber uint64, transactionID ...string) {
        if (blockNumber != c.blockNumber) {
            c.blockNumber = blockNumber;
            c.transactionID = ""
        }
		if len(transactionID) > 0 {
			c.transactionID =  transactionID[0]
		}
    }

    func (c *InMemoryCheckpointer) GetBlockNumber() uint64 {
        return c.blockNumber;
    }

    func (c *InMemoryCheckpointer) GetTransactionID() string {
        return c.transactionID;
    }