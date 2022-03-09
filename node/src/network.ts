/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { BlockEventsBuilder, BlockEventsOptions, BlockEventsWithPrivateDataBuilder, FilteredBlockEventsBuilder } from './blockeventsbuilder';
import { BlockEventsRequest, BlockEventsWithPrivateDataRequest, FilteredBlockEventsRequest } from './blockeventsrequest';
import { ChaincodeEvent } from './chaincodeevent';
import { ChaincodeEventsBuilder, ChaincodeEventsOptions } from './chaincodeeventsbuilder';
import { ChaincodeEventsRequest } from './chaincodeeventsrequest';
import { CloseableAsyncIterable, GatewayClient } from './client';
import { Contract, ContractImpl } from './contract';
import { Block } from './protos/common/common_pb';
import { BlockAndPrivateData, FilteredBlock } from './protos/peer/events_pb';
import { SigningIdentity } from './signingidentity';

/**
 * Network represents a network of nodes that are members of a specific Fabric channel. The Network can be used to
 * access deployed smart contracts, and to listen for events emitted when blocks are committed to the ledger. Network
 * instances are obtained from a Gateway using the {@link Gateway.getNetwork} method.
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
     * @param options - Event listening options.
     * @returns The iterator should be closed after use to complete the eventing session.
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
    getChaincodeEvents(chaincodeName: string, options?: ChaincodeEventsOptions): Promise<CloseableAsyncIterable<ChaincodeEvent>>;

    /**
     * Create a request to receive chaincode events emitted by transaction functions of a specific chaincode. Supports
     * off-line signing flow.
     * @param chaincodeName - Chaincode name.
     * @param options - Event listening options.
     */
    newChaincodeEventsRequest(chaincodeName: string, options?: ChaincodeEventsOptions): ChaincodeEventsRequest;

    /**
     * Get block events.
     * @param options - Event listening options.
     * @returns Block protocol buffer messages. The iterator should be closed after use to complete the eventing session.
     * @throws {@link GatewayError}
     * Thrown by the iterator if the gRPC service invocation fails.
     * @example
     * ```
     * const blocks = await network.getBlockEvents();
     * try {
     *     for async (const block of blocks) {
     *         // Process block
     *     }
     * } finally {
     *     blocks.close();
     * }
     * ```
     */
    getBlockEvents(options?: BlockEventsOptions): Promise<CloseableAsyncIterable<Block>>;

    /**
      * Create a request to receive block events. Supports off-line signing flow.
      * @param options - Event listening options.
      */
    newBlockEventsRequest(options?: BlockEventsOptions): BlockEventsRequest;

    /**
     * Get filtered block events.
     * @param options - Event listening options.
     * @returns Filtered block protocol buffer messages. The iterator should be closed after use to complete the
     * eventing session.
     * @throws {@link GatewayError}
     * Thrown by the iterator if the gRPC service invocation fails.
     * @example
     * ```
     * const blocks = await network.getFilteredBlockEvents();
     * try {
     *     for async (const block of blocks) {
     *         // Process block
     *     }
     * } finally {
     *     blocks.close();
     * }
     * ```
     */
    getFilteredBlockEvents(options?: BlockEventsOptions): Promise<CloseableAsyncIterable<FilteredBlock>>;

    /**
      * Create a request to receive filtered block events. Supports off-line signing flow.
      * @param options - Event listening options.
      */
    newFilteredBlockEventsRequest(options?: BlockEventsOptions): FilteredBlockEventsRequest;

    /**
     * Get block events with private data.
     * @param options - Event listening options.
     * @returns Blocks with private data protocol buffer messages. The iterator should be closed after use to complete
     * the eventing session.
     * @throws {@link GatewayError}
     * Thrown by the iterator if the gRPC service invocation fails.
     * @example
     * ```
     * const blocks = await network.getBlockEventsWithPrivateData();
     * try {
     *     for async (const block of blocks) {
     *         // Process block
     *     }
     * } finally {
     *     blocks.close();
     * }
     * ```
     */
    getBlockEventsWithPrivateData(options?: BlockEventsOptions): Promise<CloseableAsyncIterable<BlockAndPrivateData>>;

    /**
      * Create a request to receive block events with private data. Supports off-line signing flow.
      * @param options - Event listening options.
      */
    newBlockEventsWithPrivateDataRequest(options?: BlockEventsOptions): BlockEventsWithPrivateDataRequest;
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
        return  this.newChaincodeEventsRequest(chaincodeName, options).getEvents();
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

    async getBlockEvents(options?: Readonly<BlockEventsOptions>): Promise<CloseableAsyncIterable<Block>> {
        return this.newBlockEventsRequest(options).getEvents();
    }

    newBlockEventsRequest(options: Readonly<BlockEventsOptions> = {}): BlockEventsRequest {
        return new BlockEventsBuilder(Object.assign(
            {},
            options,
            {
                channelName: this.#channelName,
                client: this.#client,
                signingIdentity: this.#signingIdentity,
            },
        )).build();
    }

    async getFilteredBlockEvents(options?: Readonly<BlockEventsOptions>): Promise<CloseableAsyncIterable<FilteredBlock>> {
        return this.newFilteredBlockEventsRequest(options).getEvents();
    }

    newFilteredBlockEventsRequest(options: Readonly<BlockEventsOptions> = {}): FilteredBlockEventsRequest {
        return new FilteredBlockEventsBuilder(Object.assign(
            {},
            options,
            {
                channelName: this.#channelName,
                client: this.#client,
                signingIdentity: this.#signingIdentity,
            },
        )).build();
    }

    async getBlockEventsWithPrivateData(options?: Readonly<BlockEventsOptions>): Promise<CloseableAsyncIterable<BlockAndPrivateData>> {
        return this.newBlockEventsWithPrivateDataRequest(options).getEvents();
    }

    newBlockEventsWithPrivateDataRequest(options: Readonly<BlockEventsOptions> = {}): BlockEventsWithPrivateDataRequest {
        return new BlockEventsWithPrivateDataBuilder(Object.assign(
            {},
            options,
            {
                channelName: this.#channelName,
                client: this.#client,
                signingIdentity: this.#signingIdentity,
            },
        )).build();
    }
}
