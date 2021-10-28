/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Client } from '@grpc/grpc-js';
import { GatewayClient, GatewayGrpcClient, newGatewayClient } from './client';
import { Hash } from './hash/hash';
import { Identity } from './identity/identity';
import { Signer } from './identity/signer';
import { Network, NetworkImpl } from './network';
import { SigningIdentity } from './signingidentity';

/**
 * Options used when connecting to a Fabric Gateway.
 */
export interface ConnectOptions {
    /**
     * A gRPC client connection to a Fabric Gateway. This should be shared by all gateway instances connecting to the
     * same Fabric Gateway. The client connection will not be closed when the gateway is closed.
     */
    client: Client;

    /**
     * Client identity used by the gateway.
     */
    identity: Identity;

    /**
     * Signing implementation used to sign messages sent by the gateway.
     */
    signer?: Signer;

    /**
     * Hash implementation used by the gateway to generate digital signatures.
     */
    hash?: Hash;
}

/**
 * Connect to a Fabric Gateway using a client identity, gRPC connection and signing implementation.
 * @param options - Connection options.
 * @returns A connected gateway.
 */
export function connect(options: ConnectOptions): Gateway {
    return internalConnect(options);
}

export interface InternalConnectOptions extends Omit<ConnectOptions, 'client'> {
    client: GatewayGrpcClient;
}

export function internalConnect(options: InternalConnectOptions): Gateway {
    if (!options.client) {
        throw new Error('No client connection supplied');
    }
    if (!options.identity) {
        throw new Error('No identity supplied');
    }

    const signingIdentity = new SigningIdentity(options);
    const gatewayClient = newGatewayClient(options.client);

    return new GatewayImpl(gatewayClient, signingIdentity);
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
     * Close the gateway when it is no longer required. This releases all resources associated with networks and
     * contracts obtained using the Gateway, including removing event listeners.
     */
    close(): void;
}

class GatewayImpl {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;

    constructor(client: GatewayClient, signingIdentity: SigningIdentity) {
        this.#client = client;
        this.#signingIdentity = signingIdentity;
    }

    getIdentity(): Identity {
        return this.#signingIdentity.getIdentity();
    }

    getNetwork(channelName: string): Network {
        return new NetworkImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            channelName
        });
    }

    close(): void {
        // Nothing for now
    }
}
