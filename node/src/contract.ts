/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { GatewayClient } from './client';
import { newCommitError } from './commiterror';
import { Proposal } from './proposal';
import { ProposalBuilder, ProposalOptions } from './proposalbuilder';
import { SigningIdentity } from './signingidentity';
import { SubmittedTransaction } from './submittedtransaction';

/**
 * Represents a smart contract, and allows applications to:
 * - Evaluate transactions that query state from the ledger using {@link evaluateTransaction}.
 * - Submit transactions that store state to the ledger using {@link submitTransaction}.
 *
 * For more complex transaction invocations, such as including private data, transactions can be evaluated or
 * submitted using {@link evaluate} or {@link submit} respectively. The result of a submitted transaction can be
 * accessed prior to its commit to the ledger using {@link submitAsync}.
 *
 * A finer-grained transaction flow can be employed by using {@link newProposal}. This allows retry of individual steps
 * in the flow in response to errors.
 *
 * By default, proposal, transaction and commit status messages will be signed using the signing implementation
 * specified when connecting the Gateway. In cases where an external client holds the signing credentials, a default
 * signing implementation can be omitted and off-line signing can be carried out by:
 * 1. Returning the serialized proposal, transaction or commit status message along with its digest to the client for
 * them to generate a signature.
 * 1. With the serialized message and signature received from the client to create a signed proposal, transaction or
 * commit using the Gateway's {@link Gateway.newSignedProposal}, {@link Gateway.newSignedTransaction} or
 * {@link Gateway.newSignedCommit} methods respectively.
 *
 * @example Evaluate transaction
 * ```typescript
 * const result = await contract.evaluate('transactionName', {
 *     arguments: ['one', 'two'],
 *     // Specify additional proposal options here
 * });
 * ```
 *
 * @example Submit transaction
 * ```typescript
 * const result = await contract.submit('transactionName', {
 *     arguments: ['one', 'two'],
 *     // Specify additional proposal options here
 * });
 * ```
 *
 * @example Async submit
 * ```typescript
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
 * @example Fine-grained submit transaction
 * ```typescript
 * const proposal = contract.newProposal('transactionName');
 * const transaction = await proposal.endorse();
 * const commit = await transaction.submit();
 *
 * const result = transaction.getResult();
 * const status = await commit.getStatus();
 * ```
 *
 * @example Off-line signing
 * ```typescript
 * const unsignedProposal = contract.newProposal('transactionName');
 * const proposalBytes = unsignedProposal.getBytes();
 * const proposalDigest = unsignedProposal.getDigest();
 * const proposalSignature = // Generate signature from digest
 * const signedProposal = gateway.newSignedProposal(proposalBytes, proposalSignature);
 *
 * const unsignedTransaction = await signedProposal.endorse();
 * const transactionBytes = unsignedTransaction.getBytes();
 * const transactionDigest = unsignedTransaction.getDigest();
 * const transactionSignature = // Generate signature from digest
 * const signedTransaction = gateway.newSignedTransaction(transactionBytes, transactionSignature);
 *
 * const unsignedCommit = await signedTransaction.submit();
 * const commitBytes = unsignedCommit.getBytes();
 * const commitDigest = unsignedCommit.getDigest();
 * const commitSignature = // Generate signature from digest
 * const signedCommit = gateway.newSignedCommit(commitBytes, commitSignature);
 *
 * const result = signedTransaction.getResult();
 * const status = await signedCommit.getStatus();
 * ```
 */
export interface Contract {
    /**
     * Get the name of the chaincode that contains this smart contract.
     * @returns The chaincode name.
     */
    getChaincodeName(): string;

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
     * ```typescript
     * contract.evaluate(name, { arguments: [ arg1, arg2 ] });
     * ```
     * @param name - Name of the transaction to invoke.
     * @param args - Transaction arguments.
     * @returns The result returned by the transaction function.
     * @throws {@link GatewayError}
     * Thrown if the gRPC service invocation fails.
     */
    evaluateTransaction(name: string, ...args: Array<string | Uint8Array>): Promise<Uint8Array>;

    /**
     * Submit a transaction to the ledger and return its result only after it is committed to the ledger. The
     * transaction function will be evaluated on endorsing peers and then submitted to the ordering service to be
     * committed to the ledger.
     *
     * This method is equivalent to:
     * ```typescript
     * contract.submit(name, { arguments: [ arg1, arg2 ] });
     * ```
     * @param name - Name of the transaction to be invoked.
     * @param args - Transaction arguments.
     * @returns The result returned by the transaction function.
     * @throws {@link EndorseError}
     * Thrown if the endorse invocation fails.
     * @throws {@link SubmitError}
     * Thrown if the submit invocation fails.
     * @throws {@link CommitStatusError}
     * Thrown if the commit status invocation fails.
     * @throws {@link CommitError}
     * Thrown if the transaction commits unsuccessfully.
     */
    submitTransaction(name: string, ...args: Array<string | Uint8Array>): Promise<Uint8Array>;

    /**
     * Evaluate a transaction function and return its results. A transaction proposal will be evaluated on endorsing
     * peers but the transaction will not be sent to the ordering service and so will not be committed to the ledger.
     * This can be used for querying the world state.
     * @param transactionName - Name of the transaction to invoke.
     * @param options - Transaction invocation options.
     * @returns The result returned by the transaction function.
     * @throws {@link GatewayError}
     * Thrown if the gRPC service invocation fails.
     */
    evaluate(transactionName: string, options?: ProposalOptions): Promise<Uint8Array>;

    /**
     * Submit a transaction to the ledger and return its result only after it is committed to the ledger. The
     * transaction function will be evaluated on endorsing peers and then submitted to the ordering service to be
     * committed to the ledger.
     * @param transactionName - Name of the transaction to invoke.
     * @param options - Transaction invocation options.
     * @returns The result returned by the transaction function.
     * @throws {@link EndorseError}
     * Thrown if the endorse invocation fails.
     * @throws {@link SubmitError}
     * Thrown if the submit invocation fails.
     * @throws {@link CommitStatusError}
     * Thrown if the commit status invocation fails.
     * @throws {@link CommitError}
     * Thrown if the transaction commits unsuccessfully.
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
     * @throws {@link GatewayError}
     * Thrown if the gRPC service invocation fails.
     */
    submitAsync(transactionName: string, options?: ProposalOptions): Promise<SubmittedTransaction>;

    /**
     * Create a transaction proposal that can be evaluated or endorsed. Supports off-line signing flow.
     * @param transactionName - Name of the transaction to invoke.
     * @param options - Transaction invocation options.
     */
    newProposal(transactionName: string, options?: ProposalOptions): Proposal;
}

export interface ContractOptions {
    client: GatewayClient;
    signingIdentity: SigningIdentity;
    channelName: string;
    chaincodeName: string;
    contractName?: string;
}

export class ContractImpl implements Contract {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #channelName: string;
    readonly #chaincodeName: string;
    readonly #contractName?: string;

    constructor(options: Readonly<ContractOptions>) {
        this.#client = options.client;
        this.#signingIdentity = options.signingIdentity;
        this.#channelName = options.channelName;
        this.#chaincodeName = options.chaincodeName;
        this.#contractName = options.contractName;
    }

    getChaincodeName(): string {
        return this.#chaincodeName;
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
            throw newCommitError(status);
        }

        return submitted.getResult();
    }

    async submitAsync(transactionName: string, options?: Readonly<ProposalOptions>): Promise<SubmittedTransaction> {
        const transaction = await this.newProposal(transactionName, options).endorse();
        return await transaction.submit();
    }

    newProposal(transactionName: string, options: Readonly<ProposalOptions> = {}): Proposal {
        return new ProposalBuilder(Object.assign(
            {},
            options,
            {
                client: this.#client,
                signingIdentity: this.#signingIdentity,
                channelName: this.#channelName,
                chaincodeName: this.#chaincodeName,
                transactionName: this.#getQualifiedTransactionName(transactionName),
            },
        )).build();
    }

    #getQualifiedTransactionName(transactionName: string): string {
        return this.#contractName ? `${this.#contractName}:${transactionName}` : transactionName;
    }
}
