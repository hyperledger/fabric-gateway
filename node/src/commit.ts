/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { inspect } from 'util';
import { GatewayClient } from './client';
import { CommitStatusResponse, SignedCommitStatusRequest } from './protos/gateway/gateway_pb';
import { Signable } from './signable';
import { SigningIdentity } from './signingidentity';
import { Status, StatusCode } from './status';

/**
 * Allows access to information about a transaction that is committed to the ledger.
 */
export interface Commit extends Signable {
    /**
     * Get status of the committed transaction. If the transaction has not yet committed, this method blocks until the
     * commit occurs.
     * @throws {@link GatewayError}
     * Thrown if the gRPC service invocation fails.
     */
    getStatus(): Promise<Status>;

    /**
     * Get the ID of the transaction.
     */
    getTransactionId(): string;
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
            throw new Error(`Request not defined: ${inspect(this.#signedRequest)}`);
        }

        return this.#signingIdentity.hash(request);
    }

    async getStatus(): Promise<Status> {
        await this.sign();
        const response = await this.#client.commitStatus(this.#signedRequest);
        return this.newStatus(response);
    }

    getTransactionId(): string {
        return this.#transactionId
    }

    setSignature(signature: Uint8Array): void {
        this.#signedRequest.setSignature(signature);
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


    private newStatus(response: CommitStatusResponse): Status {
        const code = response.getResult() ?? StatusCode.INVALID_OTHER_REASON;
        return {
            blockNumber: BigInt(response.getBlockNumber()),
            code,
            successful: code === StatusCode.VALID,
            transactionId: this.#transactionId,
        };
    }
}
