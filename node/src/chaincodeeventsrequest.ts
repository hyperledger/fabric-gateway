/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ChaincodeEvent, newChaincodeEvents } from './chaincodeevent';
import { GatewayClient } from './client';
import { ChaincodeEventsRequest as ChaincodeEventsRequestProto, SignedChaincodeEventsRequest as SignedChaincodeEventsRequestProto } from './protos/gateway/gateway_pb';
import { Signable } from './signable';
import { SigningIdentity } from './signingidentity';
import { ErrorFirstCallback } from './utils';

export type ChaincodeEventCallback = ErrorFirstCallback<ChaincodeEvent>

/**
 * Delivers events emitted by transaction functions in a specific chaincode.
 */
export interface ChaincodeEventsRequest extends Signable {
    /**
     * Get chaincode events emitted by transaction functions of a specific chaincode.
     * @param callback - Event callback function.
     * @example
     * ```
     * await request.onEvent((err, event) => {
     *     if (err) {
     *         // Handle connection error
     *     } else {
     *         // Process event
     *     }
     * });
     * ```
     */
    onEvent(callback: ChaincodeEventCallback): Promise<void>;

    /**
     * Get chaincode events emitted by transaction functions of a specific chaincode.
     * @returns Chaincode events.
     * @example
     * ```
     * const events = await request.getEvents();
     * for async (const event of events) {
     *     // Process event
     * }
     * ```
     */
     getEvents(): Promise<AsyncIterable<ChaincodeEvent>>;
}

export interface ChaincodeEventsRequestOptions {
    readonly client: GatewayClient;
    readonly signingIdentity: SigningIdentity;
    readonly request: ChaincodeEventsRequestProto;
}

export class ChaincodeEventsRequestImpl implements ChaincodeEventsRequest {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #signedRequest: SignedChaincodeEventsRequestProto;

    constructor(options: ChaincodeEventsRequestOptions) {
        this.#client = options.client;
        this.#signingIdentity = options.signingIdentity;
        this.#signedRequest = new SignedChaincodeEventsRequestProto();
        this.#signedRequest.setRequest(options.request.serializeBinary())
    }

    async onEvent(listener: ChaincodeEventCallback): Promise<void> {
        const events = await this.getEvents();
        void (async () => {
            try {
                for await (const event of events) {
                    try {
                        await listener(undefined, event);
                    } catch (err) {
                        console.error('Chaincode event listener error:', err);
                    }
                }
            } catch (err) {
                await listener(err, undefined);
            }
        })();
    }

    async getEvents(): Promise<AsyncIterable<ChaincodeEvent>> {
        await this.sign();
        const responses = this.#client.chaincodeEvents(this.#signedRequest);
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

    private async sign(): Promise<void> {
        if (this.isSigned()) {
            return;
        }

        const signature = await this.#signingIdentity.sign(this.getDigest());
        this.setSignature(signature);
    }

    private isSigned(): boolean {
        const signatureLength = this.#signedRequest.getSignature()?.length || 0;
        return signatureLength > 0;
    }
}
