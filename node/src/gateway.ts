/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CallOptions, Client } from '@grpc/grpc-js';
import { ChaincodeEventsRequest, Commit, Proposal, Transaction } from '.';
import { ChaincodeEventsRequestImpl } from './chaincodeeventsrequest';
import { GatewayClient, GatewayGrpcClient, newGatewayClient } from './client';
import { CommitImpl } from './commit';
import { Hash } from './hash/hash';
import { Identity } from './identity/identity';
import { Signer } from './identity/signer';
import { Network, NetworkImpl } from './network';
import { ProposalImpl } from './proposal';
import { ChannelHeader, Header, Payload } from './protos/common/common_pb';
import { ChaincodeEventsRequest as ChaincodeEventsRequestProto, CommitStatusRequest, PreparedTransaction, ProposedTransaction, SignedCommitStatusRequest } from './protos/gateway/gateway_pb';
import { Proposal as ProposalProto } from './protos/peer/proposal_pb';
import { SigningIdentity } from './signingidentity';
import { TransactionImpl } from './transaction';

/**
 * Options used when connecting to a Fabric Gateway.
 * @example
 * ```
 * function defaultTimeout(): grpc.CallOptions {
 *     return {
 *         deadline: Date.now() + 5000, // 5 second timeout
 *     };
 * };
 * const options: ConnectOptions {
 *     identity: {
 *         mspId: 'myorg',
 *         credentials,
 *     },
 *     signer: signers.newPrivateKeySigner(privateKey),
 *     client: new grpc.Client('gateway.example.org:1337', grpc.credentials.createInsecure()),
 *     evaluateOptions: defaultTimeout,
 *     endorseOptions: defaultTimeout,
 *     submitOptions: defaultTimeout,
 *     commitStatusOptions: defaultTimeout,
 * };
 * ```
 */
export interface ConnectOptions {
    /**
     * A gRPC client connection to a Fabric Gateway. This should be shared by all gateway instances connecting to the
     * same Fabric Gateway. The client connection will not be closed when the gateway is closed.
     */
    client: Client;

    /**
     * Client identity used by the gateway.
     */
    identity: Identity;

    /**
     * Signing implementation used to sign messages sent by the gateway.
     */
    signer?: Signer;

    /**
     * Hash implementation used by the gateway to generate digital signatures.
     */
    hash?: Hash;

    /**
     * Supplier of default call options for endorsements.
     */
    endorseOptions?: () => CallOptions;

    /**
     * Supplier of default call options for evaluating transactions.
     */
    evaluateOptions?: () => CallOptions;
 
    /**
     * Supplier of default call options for submit of transactions to the orderer.
     */
    submitOptions?: () => CallOptions;
 
    /**
     * Supplier of default call options for retrieving transaction commit status.
     */
    commitStatusOptions?: () => CallOptions;
 
    /**
     * Supplier of default call options for chaincode events.
     */
    chaincodeEventsOptions?: () => CallOptions;
}

/**
 * Connect to a Fabric Gateway using a client identity, gRPC connection and signing implementation.
 * @param options - Connection options.
 * @returns A connected gateway.
 */
export function connect(options: Readonly<ConnectOptions>): Gateway {
    return internalConnect(options);
}

export interface InternalConnectOptions extends Omit<ConnectOptions, 'client'> {
    client: GatewayGrpcClient;
}

export function internalConnect(options: Readonly<InternalConnectOptions>): Gateway {
    if (!options.client) {
        throw new Error('No client connection supplied');
    }
    if (!options.identity) {
        throw new Error('No identity supplied');
    }

    const signingIdentity = new SigningIdentity(options);
    const gatewayClient = newGatewayClient(options.client, options);

    return new GatewayImpl(gatewayClient, signingIdentity);
}

/**
 * Gateway represents the connection of a specific client identity to a Fabric Gateway. A Gateway is obtained using the
 * {@link connect} function.
 */
export interface Gateway {
    /**
     * Get the identity used by this gateway.
     */
    getIdentity(): Identity;

    /**
     * Get a network representing the named Fabric channel.
     * @param channelName - Fabric channel name.
     */
    getNetwork(channelName: string): Network;

    /**
     * Create a proposal with the specified digital signature. Supports off-line signing flow.
     * @param bytes - Serialized proposal.
     * @param signature - Digital signature.
     * @returns A signed proposal.
     */
    newSignedProposal(bytes: Uint8Array, signature: Uint8Array): Proposal;

    /**
      * Create a transaction with the specified digital signature. Supports off-line signing flow.
      * @param bytes - Serialized proposal.
      * @param signature - Digital signature.
      * @returns A signed transaction.
      */
    newSignedTransaction(bytes: Uint8Array, signature: Uint8Array): Transaction;

    /**
     * Create a commit with the specified digital signature, which can be used to access information about a
     * transaction that is committed to the ledger. Supports off-line signing flow.
     * @param bytes - Serialized commit status request.
     * @param signature - Digital signature.
     * @returns A signed commit status request.
     */
    newSignedCommit(bytes: Uint8Array, signature: Uint8Array): Commit;

    /**
     * Create a chaincode events request with the specified digital signature. Supports off-line signing flow.
     * @param bytes - Serialized chaincode events request.
     * @param signature - Digital signature.
     * @returns A signed chaincode events request.
     */
    newSignedChaincodeEventsRequest(bytes: Uint8Array, signature: Uint8Array): ChaincodeEventsRequest;

    /**
     * Close the gateway when it is no longer required. This releases all resources associated with networks and
     * contracts obtained using the Gateway, including removing event listeners.
     */
    close(): void;
}

class GatewayImpl {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;

    constructor(client: GatewayClient, signingIdentity: SigningIdentity) {
        this.#client = client;
        this.#signingIdentity = signingIdentity;
    }

    getIdentity(): Identity {
        return this.#signingIdentity.getIdentity();
    }

    getNetwork(channelName: string): Network {
        return new NetworkImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            channelName
        });
    }

    newSignedProposal(bytes: Uint8Array, signature: Uint8Array): Proposal {
        const proposedTransaction = ProposedTransaction.deserializeBinary(bytes);
        const signedProposal = assertDefined(proposedTransaction.getProposal(), 'Missing proposal');
        const proposal = ProposalProto.deserializeBinary(signedProposal.getProposalBytes_asU8());
        const header = Header.deserializeBinary(proposal.getHeader_asU8());
        const channelHeader = ChannelHeader.deserializeBinary(header.getChannelHeader_asU8());

        const result = new ProposalImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            channelName: channelHeader.getChannelId(),
            proposedTransaction,
        });
        result.setSignature(signature);

        return result;
    }

    newSignedTransaction(bytes: Uint8Array, signature: Uint8Array): Transaction {
        const preparedTransaction = PreparedTransaction.deserializeBinary(bytes);
        const envelope = assertDefined(preparedTransaction.getEnvelope(), 'Missing transaction envelope');
        const payload = Payload.deserializeBinary(envelope.getPayload_asU8());
        const header = assertDefined(payload.getHeader(), 'Missing header');
        const channelHeader = ChannelHeader.deserializeBinary(header.getChannelHeader_asU8());

        const result = new TransactionImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            channelName: channelHeader.getChannelId(),
            preparedTransaction,
        });
        result.setSignature(signature);

        return result;
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

    close(): void {
        // Nothing for now
    }
}

function assertDefined<T>(value: T | null | undefined, message: string): T {
    if (value == undefined) {
        throw new Error(message)
    }

    return value;
}
