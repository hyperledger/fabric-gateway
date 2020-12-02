/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { Identity } from './identity/identity'
import { Signer } from './identity/signer'
import { Network, NetworkImpl } from './network';
import { Client, ClientImpl } from './impl/client';
import { SigningIdentity } from './signingidentity';

export interface ConnectOptions {
    url: string;
    identity: Identity;
    signer?: Signer;
}

export interface InternalConnectOptions extends ConnectOptions {
    client?: Client;
}

export async function connect(options: ConnectOptions): Promise<Gateway> {
    const client = (options as InternalConnectOptions).client ?? new ClientImpl(options.url);
    const signingIdentity = new SigningIdentity(options.identity, options.signer);
    return new GatewayImpl(client, signingIdentity);
}

export interface Gateway {
    getIdentity(): Identity;
    getNetwork(channelName: string): Network;
}

class GatewayImpl {
    readonly #client: Client;
    readonly #signingIdentity: SigningIdentity;

    constructor(client: Client, signingIdentity: SigningIdentity) {
        this.#client = client;
        this.#signingIdentity = signingIdentity;
    }

    getIdentity(): Identity {
        return this.#signingIdentity.getIdentity();
    }
    getNetwork(channelName: string): Network {
        return new NetworkImpl(this.#client, this.#signingIdentity, channelName);
    }
}
