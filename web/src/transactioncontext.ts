/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { SignatureHeader } from '@hyperledger/fabric-protos/lib/common/common_pb';
import { randomBytes, sha256 } from './crypto';
import { SigningIdentity } from './signingidentity';

export class TransactionContext {
    readonly #signatureHeader: SignatureHeader;
    readonly #transactionId: string;

    static async newInstance(signingIdentity: SigningIdentity): Promise<TransactionContext> {
        const nonce = randomBytes(24);
        const creator = signingIdentity.getCreator();

        const saltedCreator = concat(nonce, creator);
        const rawTransactionId = await sha256(saltedCreator);
        const transactionId = asHexString(rawTransactionId);

        const signatureHeader = new SignatureHeader();
        signatureHeader.setCreator(creator);
        signatureHeader.setNonce(nonce);

        return new TransactionContext(transactionId, signatureHeader);
    }

    private constructor(transactionId: string, signatureHeader: SignatureHeader) {
        this.#transactionId = transactionId;
        this.#signatureHeader = signatureHeader;
    }

    getTransactionId(): string {
        return this.#transactionId;
    }

    getSignatureHeader(): SignatureHeader {
        return this.#signatureHeader;
    }
}

function concat(...buffers: Uint8Array[]): Uint8Array {
    const length = buffers.reduce((total, buffer) => total + buffer.byteLength, 0);

    const result = new Uint8Array(length);
    let offset = 0;
    buffers.forEach((buffer) => {
        result.set(buffer, offset);
        offset += buffer.byteLength;
    });

    return result;
}

function asHexString(bytes: Uint8Array): string {
    return Array.from(bytes, (n) => n.toString(16).padStart(2, '0')).join('');
}
