/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { GatewayClient } from './client';
import { SigningIdentity } from './signingidentity';
import { Contract, ContractImpl } from './contract';
import { Commit, CommitImpl } from './commit';
import { gateway } from './protos/protos';

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
}