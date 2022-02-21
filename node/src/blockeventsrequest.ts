/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CallOptions } from '@grpc/grpc-js';
import { CloseableAsyncIterable, GatewayClient } from './client';
import { assertDefined } from './gateway';
import { Block, Envelope } from './protos/common/common_pb';
import { BlockAndPrivateData, FilteredBlock } from './protos/peer/events_pb';
import { Signable } from './signable';
import { SigningIdentity } from './signingidentity';

/**
 * Delivers block events.
 */
export interface BlockEventsRequest extends Signable {
    /**
     * Get block events.
     * @param options - gRPC call 
     * @returns Block protocol buffer messages. The iterator should be closed after use to complete the eventing
     * session.
     * @throws {@link GatewayError}
     * Thrown by the iterator if the gRPC service invocation fails.
     * @example
     * ```
     * const blocks = await request.getEvents();
     * try {
     *     for async (const block of blocks) {
     *         // Process block
     *     }
     * } finally {
     *     blocks.close();
     * }
     * ```
     */
     getEvents(options?: CallOptions): Promise<CloseableAsyncIterable<Block>>;
}

/**
 * Delivers filtered block events.
 */
export interface FilteredBlockEventsRequest extends Signable {
    /**
     * Get filtered block events.
     * @param options - gRPC call 
     * @returns Filtered block protocol buffer messages. The iterator should be closed after use to complete the
     * eventing session.
     * @throws {@link GatewayError}
     * Thrown by the iterator if the gRPC service invocation fails.
     * @example
     * ```
     * const blocks = await request.getEvents();
     * try {
     *     for async (const block of blocks) {
     *         // Process block
     *     }
     * } finally {
     *     blocks.close();
     * }
     * ```
     */
     getEvents(options?: CallOptions): Promise<CloseableAsyncIterable<FilteredBlock>>;
}

/**
 * Delivers block events with private data.
 */
export interface BlockEventsWithPrivateDataRequest extends Signable {
    /**
     * Get block events with private data.
     * @param options - gRPC call 
     * @returns Blocks with private data protocol buffer messages. The iterator should be closed after use to complete
     * the eventing session.
     * @throws {@link GatewayError}
     * Thrown by the iterator if the gRPC service invocation fails.
     * @example
     * ```
     * const blocks = await request.getEvents();
     * try {
     *     for async (const block of blocks) {
     *         // Process block
     *     }
     * } finally {
     *     blocks.close();
     * }
     * ```
     */
     getEvents(options?: CallOptions): Promise<CloseableAsyncIterable<BlockAndPrivateData>>;
}

export interface BlockEventsRequestOptions {
    client: GatewayClient;
    signingIdentity: SigningIdentity;
    request: Envelope;
}

type SignableBlockEventsRequestOptions = Pick<BlockEventsRequestOptions, 'request' | 'signingIdentity'>;

class SignableBlockEventsRequest implements Signable {
    readonly #signingIdentity: SigningIdentity;
    readonly #request: Envelope;

    constructor(options: Readonly<SignableBlockEventsRequestOptions>) {
        this.#signingIdentity = options.signingIdentity;
        this.#request = options.request;
    }

    getBytes(): Uint8Array {
        return this.#request.serializeBinary();
    }

    getDigest(): Uint8Array {
        return this.#signingIdentity.hash(this.#request.getPayload_asU8());
    }

    setSignature(signature: Uint8Array): void {
        this.#request.setSignature(signature);
    }

    protected async getSignedRequest(): Promise<Envelope> {
        if (!this.#isSigned()) {
            const signature = await this.#signingIdentity.sign(this.getDigest());
            this.setSignature(signature);
        }

        return this.#request;
    }

    #isSigned(): boolean {
        const signatureLength = this.#request.getSignature()?.length || 0;
        return signatureLength > 0;
    }
}

export class BlockEventsRequestImpl extends SignableBlockEventsRequest implements BlockEventsRequest {
    readonly #client: GatewayClient;

    constructor(options: Readonly<BlockEventsRequestOptions>) {
        super(options);
        this.#client = options.client;
    }

    async getEvents(options?: Readonly<CallOptions>): Promise<CloseableAsyncIterable<Block>> {
        const signedRequest = await this.getSignedRequest();
        const responses = this.#client.blockEvents(signedRequest, options);
        return {
            [Symbol.asyncIterator]: () => mapAsyncIterator(responses[Symbol.asyncIterator](), (response) => {
                const block = response.getBlock();
                return assertDefined(block, `Unexpected deliver response type: ${response.getTypeCase()}`);
            }),
            close: () => responses.close(),
        }
    }
}

export class FilteredBlockEventsRequestImpl extends SignableBlockEventsRequest implements FilteredBlockEventsRequest {
    readonly #client: GatewayClient;

    constructor(options: Readonly<BlockEventsRequestOptions>) {
        super(options);
        this.#client = options.client;
    }

    async getEvents(options?: Readonly<CallOptions>): Promise<CloseableAsyncIterable<FilteredBlock>> {
        const signedRequest = await this.getSignedRequest();
        const responses = this.#client.filteredBlockEvents(signedRequest, options);
        return {
            [Symbol.asyncIterator]: () => mapAsyncIterator(responses[Symbol.asyncIterator](), (response) => {
                const block = response.getFilteredBlock();
                return assertDefined(block, `Unexpected deliver response type: ${response.getTypeCase()}`);
            }),
            close: () => responses.close(),
        }
    }
}

export class BlockEventsWithPrivateDataRequestImpl extends SignableBlockEventsRequest implements BlockEventsWithPrivateDataRequest {
    readonly #client: GatewayClient;

    constructor(options: Readonly<BlockEventsRequestOptions>) {
        super(options);
        this.#client = options.client;
    }

    async getEvents(options?: Readonly<CallOptions>): Promise<CloseableAsyncIterable<BlockAndPrivateData>> {
        const signedRequest = await this.getSignedRequest();
        const responses = this.#client.blockEventsWithPrivateData(signedRequest, options);
        return {
            [Symbol.asyncIterator]: () => mapAsyncIterator(responses[Symbol.asyncIterator](), (response) => {
                const block = response.getBlockAndPrivateData();
                return assertDefined(block, `Unexpected deliver response type: ${response.getTypeCase()}`);
            }),
            close: () => responses.close(),
        }
    }
}

function mapAsyncIterator<T, R>(iterator: AsyncIterator<T>, map: (element: T) => R): AsyncIterator<R> {
    return {
        next: async (...args) => {
            const result = await iterator.next(...args);
            return {
                done: result.done,
                value: map(result.value as T),
            };
        }
    };
}
