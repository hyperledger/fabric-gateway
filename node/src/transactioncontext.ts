/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { randomBytes } from 'crypto';
import { SignatureHeader } from './protos/common/common_pb';
import { SigningIdentity } from './signingidentity';

export class TransactionContext {
    readonly #transactionId: string;
    readonly #signatureHeader: SignatureHeader;

    constructor(signingIdentity: SigningIdentity) {
        const nonce = randomBytes(24);
        const creator = signingIdentity.getCreator();

        const saltedCreator = Buffer.concat([nonce, creator]);
        const rawTransactionId = signingIdentity.hash(saltedCreator);
        this.#transactionId = Buffer.from(rawTransactionId).toString('hex');

        this.#signatureHeader = new SignatureHeader();
        this.#signatureHeader.setCreator(creator);
        this.#signatureHeader.setNonce(nonce);
    }

    getTransactionId(): string {
        return this.#transactionId;
    }

    getSignatureHeader(): SignatureHeader {
        return this.#signatureHeader;
    }
}
