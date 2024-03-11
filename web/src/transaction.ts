/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { PreparedTransaction } from '@hyperledger/fabric-protos/lib/gateway/gateway_pb';
import { assertDefined } from './gateway';
import { Identity } from './identity';
import { SigningIdentity } from './signingidentity';
import { parseTransactionEnvelope } from './transactionparser';
import { Envelope } from '@hyperledger/fabric-protos/lib/common/common_pb';

/**
 * Represents an endorsed transaction that can be submitted to the orderer for commit to the ledger.
 */
export interface Transaction {
    /**
     * Get the serialized bytes of the object. This is used to transfer the object state to a remote service.
     */
    getBytes(): Uint8Array;

    /**
     * Get the transaction result. This is obtained during the endorsement process when the transaction proposal is
     * run on endorsing peers.
     */
    getResult(): Uint8Array;

    /**
     * Get the transaction ID.
     */
    getTransactionId(): string;
}

export interface TransactionImplOptions {
    signingIdentity: SigningIdentity;
    preparedTransaction: PreparedTransaction;
}

export class TransactionImpl implements Transaction {
    readonly #signingIdentity: SigningIdentity;
    readonly #preparedTransaction: PreparedTransaction;
    readonly #envelope: Envelope;
    readonly #result: Uint8Array;
    readonly #identity: Identity;

    static async newInstance(options: Readonly<TransactionImplOptions>): Promise<TransactionImpl> {
        const result = new TransactionImpl(options);
        await result.#sign();
        return result;
    }

    private constructor(options: Readonly<TransactionImplOptions>) {
        this.#signingIdentity = options.signingIdentity;
        this.#preparedTransaction = options.preparedTransaction;

        const envelope = assertDefined(options.preparedTransaction.getEnvelope(), 'Missing envelope');
        this.#envelope = envelope;

        const { identity, result, transactionId } = parseTransactionEnvelope(envelope);
        this.#identity = identity;
        this.#result = result;
        this.#preparedTransaction.setTransactionId(transactionId);
    }

    getBytes(): Uint8Array {
        return this.#preparedTransaction.serializeBinary();
    }

    getResult(): Uint8Array {
        return this.#result;
    }

    getTransactionId(): string {
        return this.#preparedTransaction.getTransactionId();
    }

    getIdentity(): Identity {
        return this.#identity;
    }

    #setSignature(signature: Uint8Array): void {
        this.#envelope.setSignature(signature);
    }

    async #sign(): Promise<void> {
        const signature = await this.#signingIdentity.sign(this.#getMessage());
        this.#setSignature(signature);
    }

    #getMessage(): Uint8Array {
        return this.#envelope.getPayload_asU8();
    }
}
