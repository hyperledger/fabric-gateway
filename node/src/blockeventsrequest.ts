/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CallOptions } from '@grpc/grpc-js';
import { common, peer } from '@hyperledger/fabric-protos';
import { CloseableAsyncIterable, GatewayClient } from './client';
import { assertDefined } from './gateway';
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
     * ```typescript
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
    getEvents(options?: CallOptions): Promise<CloseableAsyncIterable<common.Block>>;
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
     * ```typescript
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
    getEvents(options?: CallOptions): Promise<CloseableAsyncIterable<peer.FilteredBlock>>;
}

/**
 * Delivers block and private data events.
 */
export interface BlockAndPrivateDataEventsRequest extends Signable {
    /**
     * Get block and private data events.
     * @param options - gRPC call
     * @returns Block and private data protocol buffer messages. The iterator should be closed after use to complete
     * the eventing session.
     * @throws {@link GatewayError}
     * Thrown by the iterator if the gRPC service invocation fails.
     * @example
     * ```typescript
     * const events = await network.getBlockAndPrivateEventsData();
     * try {
     *     for await (const event of events) {
     *         // Process block and private data event
     *     }
     * } finally {
     *     events.close();
     * }
     * ```
     */
    getEvents(options?: CallOptions): Promise<CloseableAsyncIterable<peer.BlockAndPrivateData>>;
}

export interface BlockEventsRequestOptions {
    client: GatewayClient;
    signingIdentity: SigningIdentity;
    request: common.Envelope;
}

type SignableBlockEventsRequestOptions = Pick<BlockEventsRequestOptions, 'request' | 'signingIdentity'>;

class SignableBlockEventsRequest implements Signable {
    readonly #signingIdentity: SigningIdentity;
    readonly #request: common.Envelope;

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

    protected async getSignedRequest(): Promise<common.Envelope> {
        if (!this.#isSigned()) {
            const signature = await this.#signingIdentity.sign(this.getDigest());
            this.setSignature(signature);
        }

        return this.#request;
    }

    #isSigned(): boolean {
        const signatureLength = this.#request.getSignature().length || 0;
        return signatureLength > 0;
    }
}

// @ts-expect-error Polyfill for Symbol.dispose if not present
Symbol.dispose ??= Symbol('Symbol.dispose');

export class BlockEventsRequestImpl extends SignableBlockEventsRequest implements BlockEventsRequest {
    readonly #client: GatewayClient;

    constructor(options: Readonly<BlockEventsRequestOptions>) {
        super(options);
        this.#client = options.client;
    }

    async getEvents(options?: Readonly<CallOptions>): Promise<CloseableAsyncIterable<common.Block>> {
        const signedRequest = await this.getSignedRequest();
        const responses = this.#client.blockEvents(signedRequest, options);
        return {
            [Symbol.asyncIterator]: () =>
                mapAsyncIterator(responses[Symbol.asyncIterator](), (response) =>
                    getBlock(response, () => response.getBlock()),
                ),
            close: () => {
                responses.close();
            },
            [Symbol.dispose]: () => {
                responses.close();
            },
        };
    }
}

export class FilteredBlockEventsRequestImpl extends SignableBlockEventsRequest implements FilteredBlockEventsRequest {
    readonly #client: GatewayClient;

    constructor(options: Readonly<BlockEventsRequestOptions>) {
        super(options);
        this.#client = options.client;
    }

    async getEvents(options?: Readonly<CallOptions>): Promise<CloseableAsyncIterable<peer.FilteredBlock>> {
        const signedRequest = await this.getSignedRequest();
        const responses = this.#client.filteredBlockEvents(signedRequest, options);
        return {
            [Symbol.asyncIterator]: () =>
                mapAsyncIterator(responses[Symbol.asyncIterator](), (response) =>
                    getBlock(response, () => response.getFilteredBlock()),
                ),
            close: () => {
                responses.close();
            },
            [Symbol.dispose]: () => {
                responses.close();
            },
        };
    }
}

export class BlockAndPrivateDataEventsRequestImpl
    extends SignableBlockEventsRequest
    implements BlockAndPrivateDataEventsRequest
{
    readonly #client: GatewayClient;

    constructor(options: Readonly<BlockEventsRequestOptions>) {
        super(options);
        this.#client = options.client;
    }

    async getEvents(options?: Readonly<CallOptions>): Promise<CloseableAsyncIterable<peer.BlockAndPrivateData>> {
        const signedRequest = await this.getSignedRequest();
        const responses = this.#client.blockAndPrivateDataEvents(signedRequest, options);
        return {
            [Symbol.asyncIterator]: () =>
                mapAsyncIterator(responses[Symbol.asyncIterator](), (response) =>
                    getBlock(response, () => response.getBlockAndPrivateData()),
                ),
            close: () => {
                responses.close();
            },
            [Symbol.dispose]: () => {
                responses.close();
            },
        };
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
        },
    };
}

function getBlock<T>(response: peer.DeliverResponse, getter: () => T | null | undefined): T {
    if (response.getTypeCase() === peer.DeliverResponse.TypeCase.STATUS) {
        throw new Error(`Unexpected status response: ${String(response.getStatus())}`);
    }
    const block = getter();
    return assertDefined(block, `Unexpected deliver response type: ${String(response.getTypeCase())}`);
}
