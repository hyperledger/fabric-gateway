/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { PreparedTransaction } from '@hyperledger/fabric-protos/lib/gateway/gateway_pb';
import { Identity } from './identity/identity';
import { Network, NetworkImpl } from './network';
import { Signer } from './identity/signer';
import { SigningIdentity } from './signingidentity';
import { Transaction, TransactionImpl } from './transaction';

/**
 * Options used when connecting to a Fabric Gateway.
 * @example
 * ```typescript
 * const options: ConnectOptions {
 *     identity: {
 *         mspId: 'myorg',
 *         credentials,
 *     },
 *     signer: async (message) => {
 *         const signature = await globalThis.crypto.subtle.sign(
 *             { name: 'ECDSA', hash: 'SHA-256' },
 *             privateKey,
 *             message,
 *         );
 *         return new Uint8Array(signature);
 *     },
 * };
 * ```
 */
export interface ConnectOptions {
    /**
     * Client identity used by the gateway.
     */
    identity: Identity;

    /**
     * Signing implementation used to sign messages sent by the gateway.
     */
    signer: Signer;
}

/**
 * Connect to a Fabric Gateway using a client identity and signing implementation.
 * @param options - Connection options.
 * @returns A connected gateway.
 */
export function connect(options: Readonly<ConnectOptions>): Gateway {
    assertDefined(options.identity, 'No identity supplied');
    assertDefined(options.signer, 'No signer supplied');

    const signingIdentity = new SigningIdentity(options);

    return new GatewayImpl({
        signingIdentity,
    });
}

/**
 * Gateway represents the connection of a specific client identity to a Fabric Gateway. A Gateway is obtained using the
 * {@link connect} function.
 */
export interface Gateway {
    /**
     * Get the identity used by this gateway.
     */
    getIdentity(): Identity;

    /**
     * Get a network representing the named Fabric channel.
     * @param channelName - Fabric channel name.
     */
    getNetwork(channelName: string): Network;

    /**
     * Recreate a transaction from serialized data.
     * @param bytes - Serialized proposal.
     * @returns A transaction.
     * @throws Error if the transaction creator does not match the client identity.
     */
    newTransaction(bytes: Uint8Array): Promise<Transaction>;
}

interface GatewayOptions {
    signingIdentity: SigningIdentity;
}

class GatewayImpl implements Gateway {
    readonly #signingIdentity: SigningIdentity;

    constructor(options: Readonly<GatewayOptions>) {
        this.#signingIdentity = options.signingIdentity;
    }

    getIdentity(): Identity {
        return this.#signingIdentity.getIdentity();
    }

    getNetwork(channelName: string): Network {
        return new NetworkImpl({
            signingIdentity: this.#signingIdentity,
            channelName,
        });
    }

    async newTransaction(bytes: Uint8Array): Promise<TransactionImpl> {
        const preparedTransaction = PreparedTransaction.deserializeBinary(bytes);

        const result = await TransactionImpl.newInstance({
            signingIdentity: this.#signingIdentity,
            preparedTransaction,
        });

        const identity = result.getIdentity();
        if (!equalIdentity(identity, this.getIdentity())) {
            throw new Error('Transaction creator does not match client identity');
        }

        return result;
    }
}

export function assertDefined<T>(value: T | null | undefined, message: string): T {
    if (value == undefined) {
        throw new Error(message);
    }

    return value;
}

function equalIdentity(a: Identity, b: Identity): boolean {
    return a.mspId === b.mspId && equalBytes(a.credentials, b.credentials);
}

function equalBytes(a: Uint8Array, b: Uint8Array): boolean {
    if (a.length !== b.length) {
        return false;
    }

    return a.every((value, i) => value === b[i]);
}
