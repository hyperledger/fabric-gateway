/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CallOptions } from '@grpc/grpc-js';
import { gateway } from '@hyperledger/fabric-protos';
import { GatewayClient } from './client';
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
     * @param options - gRPC call options.
     * @throws {@link CommitStatusError}
     * Thrown if the gRPC service invocation fails.
     */
    getStatus(options?: CallOptions): Promise<Status>;

    /**
     * Get the ID of the transaction.
     */
    getTransactionId(): string;
}

export interface CommitImplOptions {
    client: GatewayClient;
    signingIdentity: SigningIdentity;
    transactionId: string;
    signedRequest: gateway.SignedCommitStatusRequest;
}

export class CommitImpl implements Commit {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #transactionId: string;
    readonly #signedRequest: gateway.SignedCommitStatusRequest;

    constructor(options: Readonly<CommitImplOptions>) {
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
        return this.#signingIdentity.hash(request);
    }

    async getStatus(options?: CallOptions): Promise<Status> {
        await this.#sign();
        const response = await this.#client.commitStatus(this.#signedRequest, options);
        return this.#newStatus(response);
    }

    getTransactionId(): string {
        return this.#transactionId;
    }

    setSignature(signature: Uint8Array): void {
        this.#signedRequest.setSignature(signature);
    }

    async #sign(): Promise<void> {
        if (this.#isSigned()) {
            return;
        }

        const signature = await this.#signingIdentity.sign(this.getDigest());
        this.setSignature(signature);
    }

    #isSigned(): boolean {
        const signatureLength = this.#signedRequest.getSignature().length || 0;
        return signatureLength > 0;
    }

    #newStatus(response: gateway.CommitStatusResponse): Status {
        const code = response.getResult();
        return {
            blockNumber: BigInt(response.getBlockNumber()),
            code,
            successful: code === StatusCode.VALID,
            transactionId: this.#transactionId,
        };
    }
}
