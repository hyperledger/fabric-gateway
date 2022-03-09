/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import fs from 'fs';
import { Checkpointer } from './checkpointer';

/**
 * Interface to store checkpointer state during file read write operations .
 */
interface CheckpointerState {
    blockNumber?: string;
    transactionIDs: string[];
}

/**
 * Checkpointer class that uses the specified file to store persistent state.
 */
export class FileCheckPointer implements Checkpointer {
    #path: string;
    #blockNumber?: bigint;
    #transactionIDs: Set<string> = new Set();

    constructor(path: string) {
        this.#path = path;
    }

    async init(): Promise<void> {
        await this.#loadFromFile();
        await this.#saveToFile();
    }

    async checkpoint(blockNumber: bigint, transactionId?: string): Promise<void> {
        if (blockNumber !== this.#blockNumber) {
            this.#blockNumber = blockNumber;
            this.#transactionIDs.clear();
        }
        if (transactionId) {
            this.#transactionIDs.add(transactionId);
        }
        await this.#saveToFile();
    }

    getBlockNumber(): bigint | undefined {
        return this.#blockNumber;
    }

    getTransactionIds(): Set<string> {
        return this.#transactionIDs;
    }

    async #loadFromFile(): Promise<void> {
        const fileDataBuffer = await this.#readFile();
        if (fileDataBuffer) {
            const data = fileDataBuffer.toString();
            if (data.length !== 0) {
                const state = JSON.parse(data) as CheckpointerState;
                this.#setState(state);
            }
        }
    }

    async #readFile(): Promise<Buffer| undefined> {
        try {
            return await fs.promises.readFile(this.#path);
        } catch (e) {
            // ignore file not exist error.
        }
        return;
    }

    #setState(state: CheckpointerState): void {
        this.#blockNumber = state.blockNumber ? BigInt(state.blockNumber) : undefined;
        this.#transactionIDs = new Set(state.transactionIDs);
    }
    #getState(): CheckpointerState {
        return {
            blockNumber: (this.#blockNumber !== undefined) ? this.#blockNumber.toString() : undefined,
            transactionIDs: Array.from(this.#transactionIDs),
        };
    }

    async #saveToFile(): Promise<void> {
        const state = this.#getState();
        const bufferState = Buffer.from(JSON.stringify(state));
        await fs.promises.writeFile(this.#path, bufferState);
    }
}
