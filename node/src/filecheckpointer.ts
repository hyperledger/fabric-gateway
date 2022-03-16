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
    transactionID?: string;
}

/**
 * Checkpointer class that uses the specified file to store persistent state.
 */
export class FileCheckPointer implements Checkpointer {
    #path: string;
    #blockNumber?: bigint;
    #transactionID?: string;

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
            this.#transactionID = undefined;
        }
        if (transactionId) {
            this.#transactionID = transactionId;
        }
        await this.#saveToFile();
    }

    getBlockNumber(): bigint | undefined {
        return this.#blockNumber;
    }

    getTransactionId(): string | undefined {
        return this.#transactionID;
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
        this.#transactionID = state.transactionID;
    }

    #getState(): CheckpointerState {
        return {
            blockNumber: (this.#blockNumber !== undefined) ? this.#blockNumber.toString() : undefined,
            transactionID: this.#transactionID,
        };
    }

    async #saveToFile(): Promise<void> {
        const state = this.#getState();
        const bufferState = Buffer.from(JSON.stringify(state));
        await fs.promises.writeFile(this.#path, bufferState);
    }
}
