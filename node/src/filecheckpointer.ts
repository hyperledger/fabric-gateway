/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import fs from 'fs';
import { CheckPointer, CheckPointerState } from './checkpointer';

export class FileCheckPointer implements CheckPointer {
    path: string;
    checkPointerState: CheckPointerState = {
        blockNumber: undefined,
        transactionIDs: [],
    };

    constructor(path: string) {
        this.path = path;
    }

    async init(): Promise<void> {
        const state: CheckPointerState | undefined = await this.getStateFromFile();
        if (state) {
            this.setState(state);
        }else{
            this.saveDate();
        }
    }

    async checkpoint(blockNumber: bigint, transactionId?: string): Promise<void> {
        if (blockNumber !== this.checkPointerState.blockNumber) {
            this.checkPointerState.blockNumber = blockNumber;
            this.checkPointerState.transactionIDs = [];
        }
        if (transactionId) {
            this.checkPointerState.transactionIDs.push(transactionId);
        }
        await this.saveDate();
    }

    getBlockNumber(): Promise<bigint | undefined> {
        return Promise.resolve(this.checkPointerState.blockNumber);
    }

    getTransactionIds(): Promise<string[]> {
        return Promise.resolve(this.checkPointerState.transactionIDs);
    }

    async getStateFromFile(): Promise<CheckPointerState | undefined> {
        if(fs.existsSync(this.path)){
            const fileDataBuffer = await fs.promises.readFile(this.path);
            if (fileDataBuffer) {
                const data = fileDataBuffer.toString();
                if(data.length !== 0 ){
                    const state = JSON.parse(data,(_key, value) => {
                        if (typeof value === "string" && /^\d+n$/.test(value)) {
                          return BigInt(value.substr(0, value.length - 1));
                        }
                        return value;
                      }) as CheckPointerState;
                    return state;
                }
            }
        }
        return;
    }

    setState(checkPointerState: CheckPointerState): void {
        this.checkPointerState.blockNumber = checkPointerState.blockNumber;
        this.checkPointerState.transactionIDs = checkPointerState.transactionIDs;
    }

    async saveDate(): Promise<void> {
        const data = JSON.stringify(this.checkPointerState,(_key, value) => {
            value = typeof value === "bigint" ? value.toString() + "n" : value;
            return value;
            });
        await fs.promises.writeFile(this.path, data);
    }
}