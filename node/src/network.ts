/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { common, peer } from '@hyperledger/fabric-protos';
import {
    BlockAndPrivateDataEventsBuilder,
    BlockEventsBuilder,
    BlockEventsBuilderOptions,
    BlockEventsOptions,
    FilteredBlockEventsBuilder,
} from './blockeventsbuilder';
import { BlockAndPrivateDataEventsRequest, BlockEventsRequest, FilteredBlockEventsRequest } from './blockeventsrequest';
import { ChaincodeEvent } from './chaincodeevent';
import { ChaincodeEventsBuilder, ChaincodeEventsOptions } from './chaincodeeventsbuilder';
import { ChaincodeEventsRequest } from './chaincodeeventsrequest';
import { CloseableAsyncIterable, GatewayClient } from './client';
import { Contract, ContractImpl } from './contract';
import { SigningIdentity } from './signingidentity';

/**
 * Network represents a network of nodes that are members of a specific Fabric channel. The Network can be used to
 * access deployed smart contracts, and to listen for events emitted when blocks are committed to the ledger. Network
 * instances are obtained from a Gateway using the {@link Gateway.getNetwork} method.
 *
 * To safely handle connection errors during eventing, it is recommended to use a checkpointer to track eventing
 * progress. This allows eventing to be resumed with no loss or duplication of events.
 *
 * @example Chaincode events
 * ```typescript
 * const checkpointer = checkpointers.inMemory();
 *
 * while (true) {
 *     const events = await network.getChaincodeEvents(chaincodeName, {
 *         checkpoint: checkpointer,
 *         startBlock: BigInt(101), // Ignored if the checkpointer has checkpoint state
 *     });
 *     try {
 *         for await (const event of events) {
 *             // Process then checkpoint event
 *             await checkpointer.checkpointChaincodeEvent(event)
 *         }
 *     } catch (err: unknown) {
 *         // Connection error
 *     } finally {
 *         events.close();
 *     }
 * }
 * ```
 *
 * @example Block events
 * ```typescript
 * const checkpointer = checkpointers.inMemory();
 *
 * while (true) {
 *     const events = await network.getBlockEvents({
 *         checkpoint: checkpointer,
 *         startBlock: BigInt(101), // Ignored if the checkpointer has checkpoint state
 *     });
 *     try {
 *         for await (const event of events) {
 *             // Process then checkpoint block
 *             await checkpointer.checkpointBlock(event.getHeader().getNumber())
 *         }
 *     } catch (err: unknown) {
 *         // Connection error
 *     } finally {
 *         events.close();
 *     }
 * }
 * ```
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
     * ```typescript
     * const events = await network.getChaincodeEvents(chaincodeName, { startBlock: BigInt(101) });
     * try {
     *     for await (const event of events) {
     *         // Process event
     *     }
     * } finally {
     *     events.close();
     * }
     * ```
     */
    getChaincodeEvents(
        chaincodeName: string,
        options?: ChaincodeEventsOptions,
    ): Promise<CloseableAsyncIterable<ChaincodeEvent>>;

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
     * ```typescript
     * const blocks = await network.getBlockEvents({ startBlock: BigInt(101) });
     * try {
     *     for async (const block of blocks) {
     *         // Process block
     *     }
     * } finally {
     *     blocks.close();
     * }
     * ```
     */
    getBlockEvents(options?: BlockEventsOptions): Promise<CloseableAsyncIterable<common.Block>>;

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
     * ```typescript
     * const blocks = await network.getFilteredBlockEvents({ startBlock: BigInt(101) });
     * try {
     *     for async (const block of blocks) {
     *         // Process block
     *     }
     * } finally {
     *     blocks.close();
     * }
     * ```
     */
    getFilteredBlockEvents(options?: BlockEventsOptions): Promise<CloseableAsyncIterable<peer.FilteredBlock>>;

    /**
     * Create a request to receive filtered block events. Supports off-line signing flow.
     * @param options - Event listening options.
     */
    newFilteredBlockEventsRequest(options?: BlockEventsOptions): FilteredBlockEventsRequest;

    /**
     * Get block and private data events.
     * @param options - Event listening options.
     * @returns Blocks and private data protocol buffer messages. The iterator should be closed after use to complete
     * the eventing session.
     * @throws {@link GatewayError}
     * Thrown by the iterator if the gRPC service invocation fails.
     * @example
     * ```typescript
     * const events = await network.getBlockAndPrivateEventsData({ startBlock: BigInt(101) });
     * try {
     *     for await (const event of events) {
     *         // Process block and private data event
     *     }
     * } finally {
     *     events.close();
     * }
     * ```
     */
    getBlockAndPrivateDataEvents(
        options?: BlockEventsOptions,
    ): Promise<CloseableAsyncIterable<peer.BlockAndPrivateData>>;

    /**
     * Create a request to receive block and private data events. Supports off-line signing flow.
     * @param options - Event listening options.
     */
    newBlockAndPrivateDataEventsRequest(options?: BlockEventsOptions): BlockAndPrivateDataEventsRequest;
}

export interface NetworkOptions {
    client: GatewayClient;
    signingIdentity: SigningIdentity;
    channelName: string;
    tlsCertificateHash?: Uint8Array;
}

export class NetworkImpl implements Network {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #channelName: string;
    readonly #tlsCertificateHash?: Uint8Array;

    constructor(options: Readonly<NetworkOptions>) {
        this.#client = options.client;
        this.#signingIdentity = options.signingIdentity;
        this.#channelName = options.channelName;
        this.#tlsCertificateHash = options.tlsCertificateHash;
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

    async getChaincodeEvents(
        chaincodeName: string,
        options?: Readonly<ChaincodeEventsOptions>,
    ): Promise<CloseableAsyncIterable<ChaincodeEvent>> {
        return this.newChaincodeEventsRequest(chaincodeName, options).getEvents();
    }

    newChaincodeEventsRequest(
        chaincodeName: string,
        options: Readonly<ChaincodeEventsOptions> = {},
    ): ChaincodeEventsRequest {
        return new ChaincodeEventsBuilder(
            Object.assign({}, options, {
                chaincodeName: chaincodeName,
                channelName: this.#channelName,
                client: this.#client,
                signingIdentity: this.#signingIdentity,
            }),
        ).build();
    }

    async getBlockEvents(options?: Readonly<BlockEventsOptions>): Promise<CloseableAsyncIterable<common.Block>> {
        return this.newBlockEventsRequest(options).getEvents();
    }

    newBlockEventsRequest(options: Readonly<BlockEventsOptions> = {}): BlockEventsRequest {
        const builderOptions = this.#newBlockEventsBuilderOptions(options);
        return new BlockEventsBuilder(builderOptions).build();
    }

    async getFilteredBlockEvents(
        options?: Readonly<BlockEventsOptions>,
    ): Promise<CloseableAsyncIterable<peer.FilteredBlock>> {
        return this.newFilteredBlockEventsRequest(options).getEvents();
    }

    newFilteredBlockEventsRequest(options: Readonly<BlockEventsOptions> = {}): FilteredBlockEventsRequest {
        const builderOptions = this.#newBlockEventsBuilderOptions(options);
        return new FilteredBlockEventsBuilder(builderOptions).build();
    }

    async getBlockAndPrivateDataEvents(
        options?: Readonly<BlockEventsOptions>,
    ): Promise<CloseableAsyncIterable<peer.BlockAndPrivateData>> {
        return this.newBlockAndPrivateDataEventsRequest(options).getEvents();
    }

    newBlockAndPrivateDataEventsRequest(options: Readonly<BlockEventsOptions> = {}): BlockAndPrivateDataEventsRequest {
        const builderOptions = this.#newBlockEventsBuilderOptions(options);
        return new BlockAndPrivateDataEventsBuilder(builderOptions).build();
    }

    #newBlockEventsBuilderOptions(options: Readonly<BlockEventsOptions>): BlockEventsBuilderOptions {
        return Object.assign({}, options, {
            channelName: this.#channelName,
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            tlsCertificateHash: this.#tlsCertificateHash,
        });
    }
}
