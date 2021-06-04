/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import crypto from 'crypto';
import { common } from './protos/protos';
import { SigningIdentity } from './signingidentity';

export class TransactionContext {
    readonly #transactionId: string;
    readonly #signatureHeader: common.ISignatureHeader;

    constructor(signingIdentity: SigningIdentity) {
        const nonce = crypto.randomBytes(24);
        const creator = signingIdentity.getCreator();

        const saltedCreator = Buffer.concat([nonce, creator]);
        const rawTransactionId = signingIdentity.hash(saltedCreator);
        this.#transactionId = Buffer.from(rawTransactionId).toString('hex');

        this.#signatureHeader = {
            creator,
            nonce: nonce,
        };
    }

    getTransactionId(): string {
        return this.#transactionId;
    }

    getSignatureHeader(): common.ISignatureHeader {
        return this.#signatureHeader;
    }
}