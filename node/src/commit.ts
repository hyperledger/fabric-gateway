/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { SigningIdentity } from 'signingidentity';
import * as util from 'util';
import { GatewayClient } from './client';
import { gateway, protos } from './protos/protos';

/**
 * Allows access to information about a transaction that is committed to the ledger.
 */
export interface Commit {
    /**
     * Get the serialized commit status request message.
     */
    getBytes(): Uint8Array;

    /**
     * Get the digest of the commit status request. This is used to generate a digital signature.
     */
    getDigest(): Uint8Array;

    /**
     * Get the committed transaction status code. If the transaction has not yet committed, this method blocks until
     * the commit occurs.
     */
    getStatus(): Promise<protos.TxValidationCode>;

    /**
     * Check whether the transaction committed successfully. If the transaction has not yet committed, this method
     * blocks until the commit occurs.
     */
    isSuccessful(): Promise<boolean>;

    /**
     * Get the ID of the transaction.
     */
    getTransactionId(): string;
}

export interface CommitImplOptions {
    readonly client: GatewayClient;
    readonly signingIdentity: SigningIdentity;
    readonly transactionId: string;
    readonly signedRequest: gateway.ISignedCommitStatusRequest;
}

export class CommitImpl implements Commit {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity
    readonly #transactionId: string;
    readonly #signedRequest: gateway.ISignedCommitStatusRequest;
    #status?: protos.TxValidationCode;

    constructor(options: CommitImplOptions) {
        this.#client = options.client;
        this.#signingIdentity = options.signingIdentity;
        this.#transactionId = options.transactionId;
        this.#signedRequest = options.signedRequest;
    }

    getBytes(): Uint8Array {
        return gateway.SignedCommitStatusRequest.encode(this.#signedRequest).finish();
    }

    getDigest(): Uint8Array {
        const request = this.#signedRequest.request;
        if (!request) {
            throw new Error(`Request not defined: ${util.inspect(this.#signedRequest)}`);
        }

        return this.#signingIdentity.hash(request);
    }

    async getStatus(): Promise<protos.TxValidationCode> {
        if (this.#status === undefined) {
            await this.sign();
            const response = await this.#client.commitStatus(this.#signedRequest);
            this.#status = response.result ?? protos.TxValidationCode.INVALID_OTHER_REASON;
        }

        return this.#status;
    }

    async isSuccessful(): Promise<boolean> {
        const status = await this.getStatus();
        return status === protos.TxValidationCode.VALID;
    }

    getTransactionId(): string {
        return this.#transactionId
    }

    setSignature(signature: Uint8Array): void {
        this.#signedRequest.signature = signature;
    }

    private async sign(): Promise<void> {
        if (this.isSigned()) {
            return;
        }

        const signature = await this.#signingIdentity.sign(this.getDigest());
        this.setSignature(signature);
    }

    private isSigned(): boolean {
        const signatureLength = this.#signedRequest.signature?.length ?? 0;
        return signatureLength > 0;
    }
}
