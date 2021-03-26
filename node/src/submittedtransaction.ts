/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Commit } from "./commit";
import { protos } from "./protos/protos";

export interface SubmittedTransaction extends Commit {
    getResult(): Uint8Array;
}

export interface SubmittedTransactionImplOptions {
    readonly result: Uint8Array;
    readonly commit: Commit;
}

export class SubmittedTransactionImpl implements SubmittedTransaction {
    #result: Uint8Array;
    #commit: Commit;

    constructor(options: SubmittedTransactionImplOptions) {
        this.#result = options.result;
        this.#commit = options.commit;
    }

    getResult(): Uint8Array {
        return this.#result;
    }

    getStatus(): Promise<protos.TxValidationCode> {
        return this.#commit.getStatus();
    }

    getTransactionId(): string {
        return this.#commit.getTransactionId();
    }
}
