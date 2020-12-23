/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { GatewayClient } from "./client";
import { SigningIdentity } from "./signingidentity";
import { Contract, ContractImpl } from "./contract";

export interface Network {
    getName(): string;
    getContract(chaincodeId: string, name?: string): Contract;
}

export interface NetworkOptions {
    readonly client: GatewayClient;
    readonly signingIdentity: SigningIdentity;
    readonly channelName: string;
}

export class NetworkImpl implements Network {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #channelName: string;

    constructor(options: NetworkOptions) {
        this.#client = options.client;
        this.#signingIdentity = options.signingIdentity;
        this.#channelName = options.channelName;
    }

    getName(): string {
        return this.#channelName;
    }

    getContract(chaincodeId: string, contractName?: string): Contract {
        return new ContractImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            channelName: this.#channelName,
            chaincodeId,
            contractName,
        });
    }

}