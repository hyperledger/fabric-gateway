/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { GatewayClient } from './client';
import { Proposal, ProposalImpl } from './proposal';
import { ProposalBuilder, ProposalOptions } from './proposalbuilder';
import { PreparedTransaction, ProposedTransaction } from './protos/gateway/gateway_pb';
import { SigningIdentity } from './signingidentity';
import { StatusNames } from './status';
import { SubmittedTransaction } from './submittedtransaction';
import { Transaction, TransactionImpl } from './transaction';

/**
 * Represents a smart contract, and allows applications to:
 * - Evaluate transactions that query state from the ledger using {@link evaluateTransaction}.
 * - Submit transactions that store state to the ledger using {@link submitTransaction}.
 * 
 * For more complex transaction invocations, such as including private data, transactions can be evaluated or
 * submitted using {@link evaluate} or {@link submit} respectively. The result of a submitted transaction can be
 * accessed prior to its commit to the ledger using {@link submitAsync}.
 * 
 * By default, proposal, transaction and commit status messages will be signed using the signing implementation
 * specified when connecting the Gateway. In cases where an external client holds the signing credentials, a default
 * signing implementation can be omitted and off-line signing can be carried out by:
 * 1. Returning the serialized proposal, transaction or commit status message along with its digest to the client for
 * them to generate a signature.
 * 1. With the serialized message and signature received from the client to create a signed proposal or transaction
 * using the Contract's {@link newSignedProposal} or {@link newSignedTransaction} methods respectively, or create a
 * signed commit using the Network's {@link Network.newSignedCommit} method.
 * 
 * @example Evaluate transaction
 * ```
 * const result = await contract.evaluate('transactionName', {
 *     arguments: ['one', 'two'],
 *     // Specify additional proposal options here
 * });
 * ```
 * 
 * @example Submit transaction
 * ```
 * const result = await contract.submit('transactionName', {
 *     arguments: ['one', 'two'],
 *     // Specify additional proposal options here
 * });
 * ```
 * 
 * @example Async submit
 * ```
 * const commit = await contract.submitAsync('transactionName', {
 *     arguments: ['one', 'two']
 * });
 * const result = commit.getResult();
 * 
 * // Update UI or reply to REST request before waiting for commit status
 * 
 * const status = await commit.getStatus();
 * if (!status.successful) {
 *     throw new Error(`transaction ${status.transactionId} failed with status code ${status.code}`);
 * }
 * ```
 *
 * @example Off-line signing
 * ```
 * const unsignedProposal = contract.newProposal('transactionName');
 * const proposalBytes = unsignedProposal.getBytes();
 * const proposalDigest = unsignedProposal.getDigest();
 * // Generate signature from digest
 * const signedProposal = contract.newSignedProposal(proposalBytes, proposalSignature);
 * 
 * const unsignedTransaction = await signedProposal.endorse();
 * const transactionBytes = unsignedTransaction.getBytes();
 * const transactionDigest = unsignedTransaction.getDigest();
 * // Generate signature from digest
 * const signedTransaction = contract.newSignedTransaction(transactionBytes, transactionDigest);
 * 
 * const unsignedCommit = signedTransaction.submit();
 * const commitBytes = unsignedCommit.getBytes();
 * const commitDigest = unsignedCommit.getDigest();
 * // Generate signature from digest
 * const signedCommit = network.newSignedCommit(commitBytes, commitDigest);
 * 
 * const result = signedTransaction.getResult();
 * const status = await signedCommit.getStatus();
 * ```
 */
export interface Contract {
    /**
     * Get the ID of the chaincode that contains this smart contract.
     */
    getChaincodeId(): string;
    
    /**
     * Get the name of the smart contract within the chaincode.
     * @returns The contract name, or `undefined` for the default smart contract.
     */
    getContractName(): string | undefined;

    /**
     * Evaluate a transaction function and return its results. A transaction proposal will be evaluated on endorsing
     * peers but the transaction will not be sent to the ordering service and so will not be committed to the ledger.
     * This can be used for querying the world state.
     * 
     * This method is equivalent to:
     * ```
     * contract.evaluate(name, { arguments: [ arg1, arg2 ] });
     * ```
     * @param name - Name of the transaction to invoke.
     * @param args - Transaction arguments.
     * @returns The result returned by the transaction function.
     */
    evaluateTransaction(name: string, ...args: Array<string | Uint8Array>): Promise<Uint8Array>;

    /**
     * Submit a transaction to the ledger and return its result only after it is committed to the ledger. The
     * transaction function will be evaluated on endorsing peers and then submitted to the ordering service to be
     * committed to the ledger.
     * 
     * This method is equivalent to:
     * ```
     * contract.submit(name, { arguments: [ arg1, arg2 ] });
     * ```
     * @param name - Name of the transaction to be invoked.
     * @param args - Transaction arguments.
     * @returns The result returned by the transaction function.
     */
    submitTransaction(name: string, ...args: Array<string | Uint8Array>): Promise<Uint8Array>;

    /**
     * Evaluate a transaction function and return its results. A transaction proposal will be evaluated on endorsing
     * peers but the transaction will not be sent to the ordering service and so will not be committed to the ledger.
     * This can be used for querying the world state.
     * @param transactionName - Name of the transaction to invoke.
     * @param options - Transaction invocation options.
     * @returns The result returned by the transaction function.
     */
    evaluate(transactionName: string, options?: ProposalOptions): Promise<Uint8Array>;

    /**
     * Submit a transaction to the ledger and return its result only after it is committed to the ledger. The
     * transaction function will be evaluated on endorsing peers and then submitted to the ordering service to be
     * committed to the ledger.
     * @param transactionName - Name of the transaction to invoke.
     * @param options - Transaction invocation options.
     * @returns The result returned by the transaction function.
     */
    submit(transactionName: string, options?: ProposalOptions): Promise<Uint8Array>;

    /**
     * Submit a transaction to the ledger and return immediately after successfully sending to the orderer. The
     * transaction function will be evaluated on endorsing peers and then submitted to the ordering service to be
     * committed to the ledger. The submitted transaction that is returned can be used to obtain to the transaction
     * result, and to wait for it to be committed to the ledger.
     * @param transactionName - Name of the transaction to invoke.
     * @param options - Transaction invocation options.
     * @returns A submitted transaction, providing access to the transaction result and commit status.
     */
    submitAsync(transactionName: string, options?: ProposalOptions): Promise<SubmittedTransaction>;

    /**
     * Create a transaction proposal that can be evaluated or endorsed. Supports off-line signing flow.
     * @param transactionName - Name of the transaction to invoke.
     * @param options - Transaction invocation options.
     */
    newProposal(transactionName: string, options?: ProposalOptions): Proposal;

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
}

export interface ContractOptions {
    readonly client: GatewayClient;
    readonly signingIdentity: SigningIdentity;
    readonly channelName: string;
    readonly chaincodeId: string;
    readonly contractName?: string;
}

export class ContractImpl implements Contract {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #channelName: string;
    readonly #chaincodeId: string;
    readonly #contractName?: string;

    constructor(options: ContractOptions) {
        this.#client = options.client;
        this.#signingIdentity = options.signingIdentity;
        this.#channelName = options.channelName;
        this.#chaincodeId = options.chaincodeId;
        this.#contractName = options.contractName;
    }

    getChaincodeId(): string {
        return this.#chaincodeId;
    }

    getContractName(): string | undefined {
        return this.#contractName;
    }

    async evaluateTransaction(name: string, ...args: Array<string|Uint8Array>): Promise<Uint8Array> {
        return this.evaluate(name, { arguments: args });
    }

    async submitTransaction(name: string, ...args: Array<string|Uint8Array>): Promise<Uint8Array> {
        return this.submit(name, { arguments: args });
    }

    async evaluate(transactionName: string, options?: ProposalOptions): Promise<Uint8Array> {
        return this.newProposal(transactionName, options).evaluate();
    }

    async submit(transactionName: string, options?: ProposalOptions): Promise<Uint8Array> {
        const submitted = await this.submitAsync(transactionName, options);

        const status = await submitted.getStatus();
        if (!status.successful) {
            throw new Error(`Transaction ${status.transactionId} failed to commit with status code ${status.code} (${StatusNames[status.code]})`);
        }

        return submitted.getResult();
    }

    async submitAsync(transactionName: string, options?: ProposalOptions): Promise<SubmittedTransaction> {
        const transaction = await this.newProposal(transactionName, options).endorse();
        return await transaction.submit();
    }

    newProposal(transactionName: string, options: ProposalOptions = {}): Proposal {
        const builderOptions = Object.assign(
            {
                client: this.#client,
                signingIdentity: this.#signingIdentity,
                channelName: this.#channelName,
                chaincodeId: this.#chaincodeId,
                transactionName: this.getQualifiedTransactionName(transactionName),
                options,
            },
            options,
        );
        
        return new ProposalBuilder(builderOptions).build();
    }

    newSignedProposal(bytes: Uint8Array, signature: Uint8Array): Proposal {
        const proposedTransaction = ProposedTransaction.deserializeBinary(bytes);

        const result = new ProposalImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            channelName: this.#channelName,
            proposedTransaction,
        });
        result.setSignature(signature);

        return result;
    }

    newSignedTransaction(bytes: Uint8Array, signature: Uint8Array): Transaction {
        const preparedTransaction = PreparedTransaction.deserializeBinary(bytes);

        const result = new TransactionImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            channelName: this.#channelName,
            preparedTransaction,
        });
        result.setSignature(signature);

        return result;
    }

    private getQualifiedTransactionName(transactionName: string) {
        return this.#contractName ? `${this.#contractName}:${transactionName}` : transactionName;
    }
}
