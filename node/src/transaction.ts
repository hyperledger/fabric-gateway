/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { GatewayClient } from './client';
import { SigningIdentity } from './signingidentity';
import { common, gateway } from './protos/protos';
import * as util from 'util';

export interface Transaction {
    /**
     * Get the serialized transaction message.
     */
    getBytes(): Uint8Array;

    /**
     * Get the digest of the transaction. This is used to generate a digital signature.
     */
    getDigest(): Uint8Array;

    /**
     * Get the transaction result. This is obtained during the endorsement process when the transaction proposal is
     * run on endorsing peers.
     */
    getResult(): Uint8Array;

    /**
     * Submit the transaction to the orderer to be committed to the ledger.
     */
    submit(): Promise<Uint8Array>;
}

export interface TransactionImplOptions {
    readonly client: GatewayClient;
    readonly signingIdentity: SigningIdentity;
    readonly channelName: string;
    readonly preparedTransaction: gateway.IPreparedTransaction;
}

export class TransactionImpl implements Transaction {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #channelName: string;
    readonly #preparedTransaction: gateway.IPreparedTransaction;
    readonly #envelope: common.IEnvelope;

    constructor(options: TransactionImplOptions) {
        this.#client = options.client;
        this.#signingIdentity = options.signingIdentity;
        this.#channelName = options.channelName;
        this.#preparedTransaction = options.preparedTransaction;

        const envelope = options.preparedTransaction.envelope;
        if (!envelope) {
            throw new Error(`Envelope not defined: ${util.inspect(options.preparedTransaction)}`);
        }
        this.#envelope = envelope;
    }

    getBytes(): Uint8Array {
        return gateway.PreparedTransaction.encode(this.#preparedTransaction).finish();
    }

    getDigest(): Uint8Array {
        const payload = this.#envelope.payload;
        if (!payload) {
            throw new Error(`Payload not defined: ${util.inspect(this.#envelope)}`);
        }
        return this.#signingIdentity.hash(payload);
    }

    getResult(): Uint8Array {
        return this.#preparedTransaction?.result?.payload || new Uint8Array(0);
    }

    async submit(): Promise<Uint8Array> {
        await this.sign();
        await this.#client.submit(this.newSubmitRequest());
        return this.getResult();
    }

    setSignature(signature: Uint8Array): void {
        this.#envelope.signature = signature;
    }

    private async sign(): Promise<void> {
        if (this.isSigned()) {
            return;
        }

        const signature = await this.#signingIdentity.sign(this.getDigest());
        this.setSignature(signature);
    }

    private isSigned(): boolean {
        const signatureLength = this.#envelope.signature?.length ?? 0;
        return signatureLength > 0;
    }

    private newSubmitRequest(): gateway.ISubmitRequest {
        return {
            transaction_id: this.#preparedTransaction.transaction_id,
            channel_id: this.#channelName,
            prepared_transaction: this.#envelope,
        };
    }
}
