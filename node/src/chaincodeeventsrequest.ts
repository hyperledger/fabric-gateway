/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CallOptions } from '@grpc/grpc-js';
import { ChaincodeEvent, newChaincodeEvents } from './chaincodeevent';
import { CloseableAsyncIterable, GatewayClient } from './client';
import { ChaincodeEventsRequest as ChaincodeEventsRequestProto, SignedChaincodeEventsRequest as SignedChaincodeEventsRequestProto } from './protos/gateway/gateway_pb';
import { Signable } from './signable';
import { SigningIdentity } from './signingidentity';

/**
 * Delivers events emitted by transaction functions in a specific chaincode.
 */
export interface ChaincodeEventsRequest extends Signable {
    /**
     * Get chaincode events emitted by transaction functions of a specific chaincode.
     * @param options - gRPC call options.
     * @returns Chaincode events.
     * @example
     * ```
     * const events = await request.getEvents();
     * try {
     *     for async (const event of events) {
     *         // Process event
     *     }
     * } finally {
     *     events.close();
     * }
     * ```
     */
     getEvents(options?: CallOptions): Promise<CloseableAsyncIterable<ChaincodeEvent>>;
}

export interface ChaincodeEventsRequestOptions {
    client: GatewayClient;
    signingIdentity: SigningIdentity;
    request: ChaincodeEventsRequestProto;
}

export class ChaincodeEventsRequestImpl implements ChaincodeEventsRequest {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #signedRequest: SignedChaincodeEventsRequestProto;

    constructor(options: Readonly<ChaincodeEventsRequestOptions>) {
        this.#client = options.client;
        this.#signingIdentity = options.signingIdentity;
        this.#signedRequest = new SignedChaincodeEventsRequestProto();
        this.#signedRequest.setRequest(options.request.serializeBinary())
    }

    async getEvents(options?: Readonly<CallOptions>): Promise<CloseableAsyncIterable<ChaincodeEvent>> {
        await this.#sign();
        const responses = this.#client.chaincodeEvents(this.#signedRequest, options);
        return newChaincodeEvents(responses);
    }

    getBytes(): Uint8Array {
        return this.#signedRequest.getRequest_asU8();
    }

    getDigest(): Uint8Array {
        return this.#signingIdentity.hash(this.#signedRequest.getRequest_asU8());
    }

    setSignature(signature: Uint8Array): void {
        this.#signedRequest.setSignature(signature);
    }

    async #sign(): Promise<void> {
        if (this.#isSigned()) {
            return;
        }

        const signature = await this.#signingIdentity.sign(this.getDigest());
        this.setSignature(signature);
    }

    #isSigned(): boolean {
        const signatureLength = this.#signedRequest.getSignature()?.length || 0;
        return signatureLength > 0;
    }
}
