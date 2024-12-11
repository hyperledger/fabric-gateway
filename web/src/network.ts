/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Contract, ContractImpl } from './contract';
import { SigningIdentity } from './signingidentity';

/**
 * Network represents a network of nodes that are members of a specific Fabric channel. The Network can be used to
 * access deployed smart contracts. Network instances are obtained from a Gateway using the {@link Gateway.getNetwork}
 * method.
 */
export interface Network {
    /**
     * Get the name of the Fabric channel this network represents.
     */
    getName(): string;

    /**
     * Get a smart contract within the named chaincode. If no contract name is supplied, this is the default smart
     * contract for the named chaincode.
     * @param chaincodeName - Chaincode name.
     * @param contractName - Smart contract name.
     */
    getContract(chaincodeName: string, contractName?: string): Contract;
}

export interface NetworkOptions {
    signingIdentity: SigningIdentity;
    channelName: string;
}

export class NetworkImpl implements Network {
    readonly #signingIdentity: SigningIdentity;
    readonly #channelName: string;

    constructor(options: Readonly<NetworkOptions>) {
        this.#signingIdentity = options.signingIdentity;
        this.#channelName = options.channelName;
    }

    getName(): string {
        return this.#channelName;
    }

    getContract(chaincodeName: string, contractName?: string): Contract {
        return new ContractImpl({
            signingIdentity: this.#signingIdentity,
            channelName: this.#channelName,
            chaincodeName: chaincodeName,
            contractName,
        });
    }
}
