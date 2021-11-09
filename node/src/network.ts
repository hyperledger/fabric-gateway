/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ChaincodeEvent } from './chaincodeevent';
import { ChaincodeEventsBuilder, ChaincodeEventsOptions } from './chaincodeeventsbuilder';
import { ChaincodeEventsRequest } from './chaincodeeventsrequest';
import { CloseableAsyncIterable, GatewayClient } from './client';
import { Contract, ContractImpl } from './contract';
import { SigningIdentity } from './signingidentity';

/**
 * Network represents a blockchain network, or Fabric channel. The Network can be used to access deployed smart
 * contracts, and to listen for events emitted when blocks are committed to the ledger.
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

    /**
     * Get chaincode events emitted by transaction functions of a specific chaincode.
     * @param chaincodeName - A chaincode name.
     * @returns Chaincode events.
     * @throws {@link GatewayError}
     * Thrown by the iterator if the gRPC service invocation fails.
     * @example
     * ```
     * const events = await network.getChaincodeEvents(chaincodeName, { startBlock: BigInt(101) });
     * try {
     *     for async (const event of events) {
     *         // Process event
     *     }
     * } finally {
     *     events.close();
     * }
     * ```
     */
    getChaincodeEvents(chaincodeName: string, options?: ChaincodeEventsOptions): Promise<CloseableAsyncIterable<ChaincodeEvent>>

    /**
     * Create a request to receive chaincode events emitted by transaction functions of a specific chaincode. Supports
     * off-line signing flow.
     * @param chaincodeName - Chaincode name.
     * @param options - Event listening options.
     * @returns A chaincode events request.
     */
    newChaincodeEventsRequest(chaincodeName: string, options?: ChaincodeEventsOptions): ChaincodeEventsRequest;
}

export interface NetworkOptions {
    client: GatewayClient;
    signingIdentity: SigningIdentity;
    channelName: string;
}

export class NetworkImpl implements Network {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #channelName: string;

    constructor(options: Readonly<NetworkOptions>) {
        this.#client = options.client;
        this.#signingIdentity = options.signingIdentity;
        this.#channelName = options.channelName;
    }

    getName(): string {
        return this.#channelName;
    }

    getContract(chaincodeName: string, contractName?: string): Contract {
        return new ContractImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            channelName: this.#channelName,
            chaincodeName: chaincodeName,
            contractName,
        });
    }

    async getChaincodeEvents(chaincodeName: string, options?: Readonly<ChaincodeEventsOptions>): Promise<CloseableAsyncIterable<ChaincodeEvent>> {
        return this.newChaincodeEventsRequest(chaincodeName, options).getEvents();
    }

    newChaincodeEventsRequest(chaincodeName: string, options: Readonly<ChaincodeEventsOptions> = {}): ChaincodeEventsRequest {
        return new ChaincodeEventsBuilder(Object.assign(
            {},
            options,
            {
                chaincodeName: chaincodeName,
                channelName: this.#channelName,
                client: this.#client,
                signingIdentity: this.#signingIdentity,
            },
        )).build();
    }
}
