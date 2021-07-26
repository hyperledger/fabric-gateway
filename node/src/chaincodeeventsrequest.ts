/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ChaincodeEvent, newChaincodeEvents } from "./chaincodeevent";
import { GatewayClient } from "./client";
import { gateway } from "./protos/protos";
import { Signable } from "./signable";
import { SigningIdentity } from "./signingidentity";
import { MandatoryProperties } from "./utils";

export type ChaincodeEventCallback = (event: ChaincodeEvent) => Promise<void>;

/**
 * Delivers events emitted by transaction functions in a specific chaincode.
 */
export interface ChaincodeEventsRequest extends Signable {
    /**
     * Get chaincode events emitted by transaction functions of a specific chaincode.
     * @param callback - Event callback function.
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
    readonly request: gateway.IChaincodeEventsRequest;
}

export class ChaincodeEventsRequestImpl implements ChaincodeEventsRequest {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #signedRequest: MandatoryProperties<gateway.ISignedChaincodeEventsRequest, 'request'>;

    constructor(options: ChaincodeEventsRequestOptions) {
        this.#client = options.client;
        this.#signingIdentity = options.signingIdentity;
        this.#signedRequest = {
            request: gateway.ChaincodeEventsRequest.encode(options.request).finish(),
        }
    }

    async onEvent(listener: ChaincodeEventCallback): Promise<void> {
        const events = await this.getEvents();
        void (async () => {
            try {
                for await (const event of events) {
                    try {
                        await listener(event);
                    } catch (err) {
                        console.error('Chaincode event listener error:', err);
                    }
                }
            } catch (err) {
                // Reconnect? Surface error to listener?
                console.error('Error receiving chaincode events:', err);
            }
        })();
    }

    async getEvents(): Promise<AsyncIterable<ChaincodeEvent>> {
        await this.sign();
        const responses = this.#client.chaincodeEvents(this.#signedRequest);
        return newChaincodeEvents(responses);
    }

    getBytes(): Uint8Array {
        return this.#signedRequest.request;
    }

    getDigest(): Uint8Array {
        return this.#signingIdentity.hash(this.#signedRequest.request);
    }

    setSignature(signature: Uint8Array): void {
        this.#signedRequest.signature = signature;
    }

    private async sign(): Promise<void> {
        if (this.isSigned()) {
            return;
        }

        const signature = await this.#signingIdentity.sign(this.getDigest());
        this.setSignature(signature);
    }

    private isSigned(): boolean {
        const signatureLength = this.#signedRequest.signature?.length ?? 0;
        return signatureLength > 0;
    }
}
