/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { Contract, ContractImpl } from "./contract"
import { Client } from "impl/client";
import { SigningIdentity } from "signingidentity";

export interface Network {
    getName(): string;
    getContract(chaincodeId: string, name?: string): Contract;
}

export class NetworkImpl implements Network {
    readonly #client: Client;
    readonly #signingIdentity: SigningIdentity;
    readonly #channelName: string;

    constructor(client: Client, signingIdentity: SigningIdentity, channelName: string) {
        this.#client = client;
        this.#signingIdentity = signingIdentity;
        this.#channelName = channelName;
    }

    getName(): string {
        return this.#channelName;
    }

    getContract(chaincodeId: string, contractName?: string): Contract {
        return new ContractImpl(this.#client, this.#signingIdentity, this.#channelName, chaincodeId, contractName);
    }

}