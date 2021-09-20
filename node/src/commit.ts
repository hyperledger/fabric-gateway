/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import util from 'util';
import { GatewayClient } from './client';
import { CommitStatusResponse, SignedCommitStatusRequest } from './protos/gateway/gateway_pb';
import { TxValidationCode, TxValidationCodeMap } from './protos/peer/transaction_pb';
import { Signable } from './signable';
import { SigningIdentity } from './signingidentity';

/**
 * Enumeration of transaction status codes.
 */
export const CommitStatus = Object.freeze(TxValidationCode);

export const CommitStatusNames = Object.freeze(
    Object.fromEntries(
        Object.entries(CommitStatus)
            .filter(([_, code]) => typeof code === 'number') // eslint-disable-line @typescript-eslint/no-unused-vars
            .map(([name, code]) => [code, name])
    ) as { [P in keyof typeof CommitStatus as typeof CommitStatus[P]]: P }
);
 
/**
 * Allows access to information about a transaction that is committed to the ledger.
 */
export interface Commit extends Signable {
    /**
     * Get the committed transaction status code. If the transaction has not yet committed, this method blocks until
     * the commit occurs. The return value corresponds to one of the values enumerated by {@link CommitStatus}.
     */
    getStatus(): Promise<TxValidationCodeMap[keyof TxValidationCodeMap]>;

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
    getBlockNumber(): Promise<bigint>;
}

export interface CommitImplOptions {
    readonly client: GatewayClient;
    readonly signingIdentity: SigningIdentity;
    readonly transactionId: string;
    readonly signedRequest: SignedCommitStatusRequest;
}

export class CommitImpl implements Commit {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity
    readonly #transactionId: string;
    readonly #signedRequest: SignedCommitStatusRequest;
    #response?: CommitStatusResponse;

    constructor(options: CommitImplOptions) {
        this.#client = options.client;
        this.#signingIdentity = options.signingIdentity;
        this.#transactionId = options.transactionId;
        this.#signedRequest = options.signedRequest;
    }

    getBytes(): Uint8Array {
        return this.#signedRequest.serializeBinary();
    }

    getDigest(): Uint8Array {
        const request = this.#signedRequest.getRequest_asU8();
        if (!request) {
            throw new Error(`Request not defined: ${util.inspect(this.#signedRequest)}`);
        }

        return this.#signingIdentity.hash(request);
    }

    async getStatus(): Promise<TxValidationCodeMap[keyof TxValidationCodeMap]> {
        const response = await this.getCommitStatus();
        return response.getResult() ?? CommitStatus.INVALID_OTHER_REASON;
    }

    async isSuccessful(): Promise<boolean> {
        const status = await this.getStatus();
        return status === CommitStatus.VALID;
    }

    getTransactionId(): string {
        return this.#transactionId
    }

    async getBlockNumber(): Promise<bigint> {
        const response = await this.getCommitStatus();
        const blockNumber = response.getBlockNumber();

        if (blockNumber == undefined) {
            throw new Error('Missing block number');
        }

        return BigInt(blockNumber);
    }

    setSignature(signature: Uint8Array): void {
        this.#signedRequest.setSignature(signature);
    }

    private async getCommitStatus(): Promise<CommitStatusResponse> {
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
        const signatureLength = this.#signedRequest.getSignature()?.length || 0;
        return signatureLength > 0;
    }
}
