/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { SigningIdentity } from './signingidentity';
import util from 'util';
import { GatewayClient } from './client';
import { gateway, protos } from './protos/protos';
import Long from 'long';

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


    /**
     * Get the block number in which the transaction committed. If the transaction has not yet committed, this method
     * blocks until the commit occurs.
     */
    getBlockNumber(): Promise<Long>;
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
    #response?: gateway.ICommitStatusResponse;

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
        const response = await this.getCommitStatus();
        return response.result ?? protos.TxValidationCode.INVALID_OTHER_REASON;
    }

    async isSuccessful(): Promise<boolean> {
        const status = await this.getStatus();
        return status === protos.TxValidationCode.VALID;
    }

    getTransactionId(): string {
        return this.#transactionId
    }

    async getBlockNumber(): Promise<Long> {
        const response = await this.getCommitStatus();
        const blockNumber = response.block_number;

        if (blockNumber == undefined) {
            throw new Error('Missing block number');
        }

        if (Long.isLong(blockNumber)) {
            return blockNumber;
        }

        return Long.fromInt(blockNumber, true);
    }

    setSignature(signature: Uint8Array): void {
        this.#signedRequest.signature = signature;
    }

    private async getCommitStatus(): Promise<gateway.ICommitStatusResponse> {
        if (this.#response === undefined) {
            await this.sign();
            this.#response = await this.#client.commitStatus(this.#signedRequest);
        }

        return this.#response;
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
