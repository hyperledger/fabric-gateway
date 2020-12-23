/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { GatewayClient } from 'client';
import { SigningIdentity } from 'signingidentity';
import { protos } from './protos/protos';

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
    readonly preparedTransaction: protos.IPreparedTransaction;
}

export class TransactionImpl implements Transaction {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #preparedTransaction: protos.IPreparedTransaction;

    constructor(options: TransactionImplOptions) {
        this.#client = options.client;
        this.#signingIdentity = options.signingIdentity;
        this.#preparedTransaction = options.preparedTransaction;
    }

    getBytes(): Uint8Array {
        return protos.PreparedTransaction.encode(this.#preparedTransaction).finish();
    }

    getDigest(): Uint8Array {
        return this.#signingIdentity.hash(this.#preparedTransaction.envelope!.payload!);
    }

    getResult(): Uint8Array {
        return this.#preparedTransaction.response?.value || new Uint8Array(0);
    }

    async submit(): Promise<Uint8Array> {
        if (!this.isSigned()) {
            this.sign();
        }
        await this.#client.submit(this.#preparedTransaction); // TODO: need to return before the commit
        return this.getResult();
    }

    private isSigned(): boolean {
        const signatureLength = this.#preparedTransaction.envelope!.signature?.length ?? 0;
        return signatureLength > 0;
    }

    private sign(): void {
        this.#preparedTransaction.envelope!.signature = this.#signingIdentity.sign(this.getDigest());
    }
}
