/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CallOptions } from '@grpc/grpc-js';
import { common, gateway } from '@hyperledger/fabric-protos';
import { GatewayClient } from './client';
import { assertDefined } from './gateway';
import { Signable } from './signable';
import { SigningIdentity } from './signingidentity';
import { SubmittedTransaction, SubmittedTransactionImpl } from './submittedtransaction';
import { parseTransactionEnvelope } from './transactionparser';

/**
 * Represents an endorsed transaction that can be submitted to the orderer for commit to the ledger.
 */
export interface Transaction extends Signable {
    /**
     * Get the transaction result. This is obtained during the endorsement process when the transaction proposal is
     * run on endorsing peers.
     */
    getResult(): Uint8Array;

    /**
     * Get the transaction ID.
     */
    getTransactionId(): string;

    /**
     * Submit the transaction to the orderer to be committed to the ledger.
     * @param options - gRPC call options.
     * @throws {@link SubmitError}
     * Thrown if the gRPC service invocation fails.
     */
    submit(options?: CallOptions): Promise<SubmittedTransaction>;
}

export interface TransactionImplOptions {
    client: GatewayClient;
    signingIdentity: SigningIdentity;
    preparedTransaction: gateway.PreparedTransaction;
}

export class TransactionImpl implements Transaction {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #channelName: string;
    readonly #preparedTransaction: gateway.PreparedTransaction;
    readonly #envelope: common.Envelope;
    readonly #result: Uint8Array;

    constructor(options: Readonly<TransactionImplOptions>) {
        this.#client = options.client;
        this.#signingIdentity = options.signingIdentity;
        this.#preparedTransaction = options.preparedTransaction;

        const envelope = assertDefined(options.preparedTransaction.getEnvelope(), 'Missing envelope');
        this.#envelope = envelope;

        const { channelName, result } = parseTransactionEnvelope(envelope);
        this.#channelName = channelName;
        this.#result = result;
    }

    getBytes(): Uint8Array {
        return this.#preparedTransaction.serializeBinary();
    }

    getDigest(): Uint8Array {
        const payload = this.#envelope.getPayload_asU8();
        return this.#signingIdentity.hash(payload);
    }

    getResult(): Uint8Array {
        return this.#result;
    }

    getTransactionId(): string {
        return this.#preparedTransaction.getTransactionId();
    }

    async submit(options?: Readonly<CallOptions>): Promise<SubmittedTransaction> {
        await this.#sign();
        await this.#client.submit(this.#newSubmitRequest(), options);

        return new SubmittedTransactionImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            transactionId: this.getTransactionId(),
            signedRequest: this.#newSignedCommitStatusRequest(),
            result: this.getResult(),
        });
    }

    setSignature(signature: Uint8Array): void {
        this.#envelope.setSignature(signature);
    }

    async #sign(): Promise<void> {
        if (this.#isSigned()) {
            return;
        }

        const signature = await this.#signingIdentity.sign(this.getDigest());
        this.setSignature(signature);
    }

    #isSigned(): boolean {
        const signatureLength = this.#envelope.getSignature_asU8()?.length || 0;
        return signatureLength > 0;
    }

    #newSubmitRequest(): gateway.SubmitRequest {
        const result = new gateway.SubmitRequest();
        result.setTransactionId(this.getTransactionId());
        result.setChannelId(this.#channelName);
        result.setPreparedTransaction(this.#envelope);
        return result;
    }

    #newSignedCommitStatusRequest(): gateway.SignedCommitStatusRequest {
        const result = new gateway.SignedCommitStatusRequest();
        result.setRequest(this.#newCommitStatusRequest().serializeBinary());
        return result;
    }

    #newCommitStatusRequest(): gateway.CommitStatusRequest {
        const result = new gateway.CommitStatusRequest();
        result.setChannelId(this.#channelName);
        result.setTransactionId(this.getTransactionId());
        result.setIdentity(this.#signingIdentity.getCreator());
        return result;
    }
}
