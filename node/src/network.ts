/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ChaincodeEvent } from './chaincodeevent';
import { ChaincodeEventsBuilder, ChaincodeEventsOptions } from './chaincodeeventsbuilder';
import { ChaincodeEventsRequest, ChaincodeEventsRequestImpl } from './chaincodeeventsrequest';
import { CloseableAsyncIterable, GatewayClient } from './client';
import { Commit, CommitImpl } from './commit';
import { Contract, ContractImpl } from './contract';
import { ChaincodeEventsRequest as ChaincodeEventsRequestProto, CommitStatusRequest, SignedCommitStatusRequest } from './protos/gateway/gateway_pb';
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
     * @returns Chaincode events.
     * @example
     * ```
     * const events = await network.getChaincodeEvents(chaincodeId, { startBlock: BigInt(101) });
     * try {
     *     for async (const event of events) {
     *         // Process event
     *     }
     * } finally {
     *     events.close();
     * }
     * ```
     */
    getChaincodeEvents(chaincodeId: string, options?: ChaincodeEventsOptions): Promise<CloseableAsyncIterable<ChaincodeEvent>>

    /**
     * Create a request to receive chaincode events emitted by transaction functions of a specific chaincode. Supports
     * off-line signing flow.
     * @param chaincodeId - Chaincode name.
     * @param options - Event listening options.
     * @returns A chaincode events request.
     */
    newChaincodeEventsRequest(chaincodeId: string, options?: ChaincodeEventsOptions): ChaincodeEventsRequest;

    /**
     * Create a chaincode events request with the specified digital signature. Supports off-line signing flow.
     * @param bytes - Serialized chaincode events request.
     * @param signature - Digital signature.
     * @returns A signed chaincode events request.
     */
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
        const signedRequest = SignedCommitStatusRequest.deserializeBinary(bytes);
        const request = CommitStatusRequest.deserializeBinary(signedRequest.getRequest_asU8());

        const result = new CommitImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            transactionId: request.getTransactionId(),
            signedRequest: signedRequest,
        });
        result.setSignature(signature);

        return result;
    }

    async getChaincodeEvents(chaincodeId: string, options?: ChaincodeEventsOptions): Promise<CloseableAsyncIterable<ChaincodeEvent>> {
        return this.newChaincodeEventsRequest(chaincodeId, options).getEvents();
    }

    newChaincodeEventsRequest(chaincodeId: string, options: ChaincodeEventsOptions = {}): ChaincodeEventsRequest {
        const builderOptions = Object.assign(
            {
                chaincodeId,
                channelName: this.#channelName,
                client: this.#client,
                signingIdentity: this.#signingIdentity,
                options,
            },
            options,
        );
        
        return new ChaincodeEventsBuilder(builderOptions).build();
    }

    newSignedChaincodeEventsRequest(bytes: Uint8Array, signature: Uint8Array): ChaincodeEventsRequest {
        const request = ChaincodeEventsRequestProto.deserializeBinary(bytes);

        const result = new ChaincodeEventsRequestImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            request,
        });
        result.setSignature(signature);

        return result;
    }
}
