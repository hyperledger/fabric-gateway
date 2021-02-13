/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as grpc from '@grpc/grpc-js';
import { Hash } from 'hash/hash';
import { GatewayClient, newGatewayClient } from './client';
import { Identity } from './identity/identity';
import { Signer } from './identity/signer';
import { Network, NetworkImpl } from './network';
import { SigningIdentity } from './signingidentity';

export interface ConnectOptions {
    url?: string;
    client?: grpc.Client;
    identity: Identity;
    signer?: Signer;
    hash?: Hash;
}

export interface InternalConnectOptions extends ConnectOptions {
    gatewayClient?: GatewayClient;
}

type Closer = () => void;

const noOpCloser: Closer = () => {
    // Do nothing
};

export async function connect(options: ConnectOptions): Promise<Gateway> {
    if (!options.identity) {
        throw new Error('No identity supplied');
    }
    const signingIdentity = new SigningIdentity(options);

    const gatewayClient = (options as InternalConnectOptions).gatewayClient;
    if (gatewayClient) {
        return new GatewayImpl(gatewayClient, signingIdentity, noOpCloser);
    }

    if (options.client) {
        return new GatewayImpl(newGatewayClient(options.client), signingIdentity, noOpCloser);
    }

    if (typeof options.url === 'string') {
        const GrpcClient = grpc.makeGenericClientConstructor({}, '');
        const grpcClient = new GrpcClient(options.url, grpc.credentials.createInsecure());
        const closer = () => grpcClient.close();
        return new GatewayImpl(newGatewayClient(grpcClient), signingIdentity, closer);
    }

    throw new Error('No client connection supplied.')
}

export interface Gateway {
    getIdentity(): Identity;
    getNetwork(channelName: string): Network;
    close(): void;
}

class GatewayImpl {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #closer: Closer;

    constructor(client: GatewayClient, signingIdentity: SigningIdentity, closer: Closer) {
        this.#client = client;
        this.#signingIdentity = signingIdentity;
        this.#closer = closer;
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
        this.#closer();
    }
}
