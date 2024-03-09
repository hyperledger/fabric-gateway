/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import fs from 'node:fs';
import { ChaincodeEvent } from './chaincodeevent';
import { Checkpointer } from './checkpointer';

/**
 * Interface to store checkpointer state during file read write operations .
 */
interface CheckpointerState {
    blockNumber?: string;
    transactionId?: string;
}

export class FileCheckPointer implements Checkpointer {
    #path: string;
    #blockNumber?: bigint;
    #transactionId?: string;

    constructor(path: string) {
        this.#path = path;
    }

    async init(): Promise<void> {
        await this.#loadFromFile();
        await this.#saveToFile();
    }

    async checkpointBlock(blockNumber: bigint): Promise<void> {
        this.#blockNumber = blockNumber + BigInt(1);
        this.#transactionId = undefined;
        await this.#saveToFile();
    }

    async checkpointTransaction(blockNumber: bigint, transactionId: string): Promise<void> {
        this.#blockNumber = blockNumber;
        this.#transactionId = transactionId;
        await this.#saveToFile();
    }

    async checkpointChaincodeEvent(event: ChaincodeEvent): Promise<void> {
        await this.checkpointTransaction(event.blockNumber, event.transactionId);
    }

    getBlockNumber(): bigint | undefined {
        return this.#blockNumber;
    }

    getTransactionId(): string | undefined {
        return this.#transactionId;
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

    async #readFile(): Promise<Buffer | undefined> {
        try {
            return await fs.promises.readFile(this.#path);
        } catch (e) {
            // ignore file not exist error.
        }
        return;
    }

    #setState(state: CheckpointerState): void {
        this.#blockNumber = state.blockNumber != undefined ? BigInt(state.blockNumber) : state.blockNumber;
        this.#transactionId = state.transactionId;
    }

    #getState(): CheckpointerState {
        return {
            blockNumber: this.#blockNumber?.toString(),
            transactionId: this.#transactionId,
        };
    }

    async #saveToFile(): Promise<void> {
        const state = this.#getState();
        const bufferState = Buffer.from(JSON.stringify(state));
        await fs.promises.writeFile(this.#path, bufferState);
    }
}
