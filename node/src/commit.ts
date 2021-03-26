/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { GatewayClient } from './client';
import { gateway, protos } from './protos/protos';

/**
 * Allows access to information about a transaction that is committed to the ledger.
 */
export interface Commit {
    /**
     * Get the committed transaction status code. If the transaction has not yet committed, this method blocks until
     * the commit occurs.
     */
    getStatus(): Promise<protos.TxValidationCode>;

    /**
     * Get the ID of the transaction.
     */
    getTransactionId(): string;
}

export interface CommitImplOptions {
    readonly client: GatewayClient;
    readonly channelName: string;
    readonly transactionId: string;
}

export class CommitImpl implements Commit {
    readonly #client: GatewayClient;
    readonly #channelName: string;
    readonly #transactionId: string;

    constructor(options: CommitImplOptions) {
        this.#client = options.client;
        this.#channelName = options.channelName;
        this.#transactionId = options.transactionId;
    }

    async getStatus(): Promise<protos.TxValidationCode> {
        const response = await this.#client.commitStatus(this.newCommitStatusRequest());
        return response.result ?? protos.TxValidationCode.INVALID_OTHER_REASON;
    }

    getTransactionId(): string {
        return this.#transactionId
    }

    private newCommitStatusRequest(): gateway.ICommitStatusRequest {
        return {
            channel_id: this.#channelName,
            transaction_id: this.#transactionId,
        }
    }
}
