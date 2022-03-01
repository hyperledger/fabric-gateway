/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CheckPointer, CheckPointerState } from './checkpointer';

export class InmemoryCheckPointer implements CheckPointer {

    checkPointerState: CheckPointerState = {
        blockNumber: undefined,
        transactionIDs: [],
    };

    checkpoint(blockNumber: bigint, transactionId?: string): Promise<void> {
        if (blockNumber !== this.checkPointerState.blockNumber) {
            this.checkPointerState.blockNumber = blockNumber;
            this.checkPointerState.transactionIDs = [];
        }
        if (transactionId) {
            this.checkPointerState.transactionIDs.push(transactionId);
        }
        return Promise.resolve();
    }

    getBlockNumber(): Promise<bigint | undefined> {
        return Promise.resolve(this.checkPointerState.blockNumber);
    }

    getTransactionIds(): Promise<string[]> {
        return Promise.resolve(this.checkPointerState.transactionIDs);
    }
}
