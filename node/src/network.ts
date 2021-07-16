/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ChaincodeEvent } from './chaincodeevent';
import { ChaincodeEventCallback, ChaincodeEventsRequest, ChaincodeEventsRequestImpl } from './chaincodeeventsrequest';
import { GatewayClient } from './client';
import { Commit, CommitImpl } from './commit';
import { Contract, ContractImpl } from './contract';
import { gateway } from './protos/protos';
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
     * @param chaincodeId - Chaincode name.
     * @param name - Smart contract name.
     */
    getContract(chaincodeId: string, name?: string): Contract;

    /**
     * Create a commit with the specified digital signature, which can be used to access information about a
     * transaction that is committed to the ledger. Supports off-line signing flow.
     * @param bytes - Serialized commit status request.
     * @param signature - Digital signature.
     * @returns A signed commit status request.
     */
    newSignedCommit(bytes: Uint8Array, signature: Uint8Array): Commit;

    /**
     * Get chaincode events emitted by transaction functions of a specific chaincode.
     * @param chaincodeId - A chaincode ID.
     * @param callback - Event callback function.
     */
    onChaincodeEvent(chaincodeId: string, callback: ChaincodeEventCallback): Promise<void>;

    /**
     * Get chaincode events emitted by transaction functions of a specific chaincode.
     * @param chaincodeId - A chaincode ID.
     * @returns Chaincode events.
     * @example
     * ```
     * const events = await network.getChaincodeEvents(chaincodeId);
     * for async (const event of events) {
     *     // Process event
     * }
     * ```
     */
    getChaincodeEvents(chaincodeId: string): Promise<AsyncIterable<ChaincodeEvent>>

    newChaincodeEventsRequest(chaincodeId: string): ChaincodeEventsRequest;
    newSignedChaincodeEventsRequest(bytes: Uint8Array, signature: Uint8Array): ChaincodeEventsRequest;
}

export interface NetworkOptions {
    readonly client: GatewayClient;
    readonly signingIdentity: SigningIdentity;
    readonly channelName: string;
}

export class NetworkImpl implements Network {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #channelName: string;

    constructor(options: NetworkOptions) {
        this.#client = options.client;
        this.#signingIdentity = options.signingIdentity;
        this.#channelName = options.channelName;
    }

    getName(): string {
        return this.#channelName;
    }

    getContract(chaincodeId: string, contractName?: string): Contract {
        return new ContractImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            channelName: this.#channelName,
            chaincodeId,
            contractName,
        });
    }

    newSignedCommit(bytes: Uint8Array, signature: Uint8Array): Commit {
        const signedRequest = gateway.SignedCommitStatusRequest.decode(bytes);
        const request = gateway.CommitStatusRequest.decode(signedRequest.request);

        const result = new CommitImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            transactionId: request.transaction_id,
            signedRequest: signedRequest,
        });
        result.setSignature(signature);

        return result;
    }

    async onChaincodeEvent(chaincodeId: string, listener: ChaincodeEventCallback): Promise<void> {
        const request = this.newChaincodeEventsRequest(chaincodeId);
        return request.onEvent(listener);
    }

    async getChaincodeEvents(chaincodeId: string): Promise<AsyncIterable<ChaincodeEvent>> {
        const request = this.newChaincodeEventsRequest(chaincodeId);
        return request.getEvents();
    }

    newChaincodeEventsRequest(chaincodeId: string): ChaincodeEventsRequest {
        const request = this.newChaincodeEventsRequestProto(chaincodeId);

        return new ChaincodeEventsRequestImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            request,
        });
    }

    newSignedChaincodeEventsRequest(bytes: Uint8Array, signature: Uint8Array): ChaincodeEventsRequest {
        const request = gateway.ChaincodeEventsRequest.decode(bytes);

        const result = new ChaincodeEventsRequestImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            request,
        });
        result.setSignature(signature);

        return result;
    }

    private newChaincodeEventsRequestProto(chaincodeId: string): gateway.IChaincodeEventsRequest {
        return {
            channel_id: this.#channelName,
            chaincode_id: chaincodeId,
            identity: this.#signingIdentity.getCreator(),
        };
    }
}
