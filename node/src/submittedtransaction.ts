/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Commit, CommitImpl, CommitImplOptions } from './commit';

/**
 * Allows access to the transaction result and its commit status on the ledger.
 */
export interface SubmittedTransaction extends Commit {
    /**
     * Get the transaction result. This is obtained during the endorsement process when the transaction proposal is
     * run on endorsing peers and so is available immediately. The transaction might subsequently fail to commit
     * successfully.
     */
    getResult(): Uint8Array;
}

export interface SubmittedTransactionImplOptions extends CommitImplOptions {
    result: Uint8Array;
}

export class SubmittedTransactionImpl extends CommitImpl {
    #result: Uint8Array;

    constructor(options: Readonly<SubmittedTransactionImplOptions>) {
        super(options);
        this.#result = options.result;
    }

    getResult(): Uint8Array {
        return this.#result;
    }
}
