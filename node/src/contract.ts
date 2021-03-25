/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { GatewayClient } from "./client";
import { Proposal, ProposalImpl } from "./proposal";
import { ProposalBuilder, ProposalOptions } from "./proposalbuilder";
import { gateway, protos } from "./protos/protos";
import { SigningIdentity } from "./signingidentity";
import { Transaction, TransactionImpl } from "./transaction";

/**
 * Represents a smart contract, and allows applications to:
 * - Evaluate transactions that query state from the ledger using `evaluateTransaction()`.
 * - Submit transactions that store state to the ledger using `submitTransaction()`.
 * 
 * For more complex transaction invocations, such as including transient data, transactions can be evaluated or
 * submitted using `evaluate()` or `submitSync()` respectively.
 * 
 * By default, proposal and transaction messages will be signed using the signing implementation specified when
 * connecting the Gateway. In cases where an external client holds the signing credentials, a signing implementation
 * can be omitted when connecting the Gateway and off-line signing can be carried out by:
 * 1. Returning the serialized proposal or transaction message along with its digest to the client for them to
 * generate a signature.
 * 1. On receipt of the serialized message and signature from the client, creating a signed proposal or transaction
 * using the Contract's `newSignedProposal()` or `newSignedTransaction()` methods respectively.
 * 
 * @example Evaluate transaction
 * 
 * const result = await contract.evaluate('transactionName', {
 *     arguments: ['one', 'two'],
 *     // Specify additional proposal options, such as transient data
 * });
 * 
 * @example Submit transaction
 * 
 * const result = await contract.submitSync('transactionName', {
 *     arguments: ['one', 'two'],
 *     // Specify additional proposal options, such as transient data
 * });
 *
 * @example Off-line signing of proposal
 *
 * const unsignedProposal = contract.newProposal('transactionName');
 * const proposalBytes = unsignedProposal.getBytes();
 * const proposalDigest = unsignedProposal.getDigest();
 * // Generate signature from digest
 * const signedProposal = contract.newSignedProposal(proposalBytes, proposalSignature);
 *
 * @example Off-line signing of transaction
 *
 * const unsignedTransaction = await proposal.endorse();
 * const transactionBytes = unsignedTransaction.getBytes();
 * const transactionDigest = unsignedTransaction.getDigest();
 * // Generate signature from digest
 * const signedTransaction = contract.newSignedTransaction(transactionBytes, transactionSignature);
 */
export interface Contract {
    /**
     * Get the ID of the chaincode that contains this contract.
     */
    getChaincodeId(): string;
    
    /**
     * Get the name of the contract within the chaincode.
     * @returns the contract name, or `undefined` for the default contract.
     */
    getContractName(): string | undefined;

    /**
     * Evaluate a transaction function and return its results. A transaction proposal will be evaluated on endorsing
     * peers but the transaction will not be sent to the ordering service and so will not be committed to the ledger.
     * This can be used for querying the world state.
     * @param name Name of the transaction to invoke.
     * @param args Transaction arguments.
     * @returns the result returned by the transaction function.
     */
    evaluateTransaction(name: string, ...args: Array<string|Uint8Array>): Promise<Uint8Array>;

    /**
     * Submit a transaction to the ledger and return its result only after it is committed to the ledger. The
     * transaction function will be evaluated on endorsing peers and then submitted to the ordering service to be
     * committed to the ledger.
     * @param name Name of the transaction to be invoked.
     * @param args Transaction arguments.
     * @returns the result returned by the transaction function.
     */
    submitTransaction(name: string, ...args: Array<string|Uint8Array>): Promise<Uint8Array>;

    /**
     * Evaluate a transaction function and return its results. A transaction proposal will be evaluated on endorsing
     * peers but the transaction will not be sent to the ordering service and so will not be committed to the ledger.
     * This can be used for querying the world state.
     * @param transactionName Name of the transaction to invoke.
     * @param options Transaction invocation options.
     * @returns the result returned by the transaction function.
     */
    evaluate(transactionName: string, options?: ProposalOptions): Promise<Uint8Array>;

    /**
     * Submit a transaction to the ledger and return its result only after it is committed to the ledger. The
     * transaction function will be evaluated on endorsing peers and then submitted to the ordering service to be
     * committed to the ledger.
     * @param transactionName Name of the transaction to invoke.
     * @param options Transaction invocation options.
     * @returns the result returned by the transaction function.
     */
    submitSync(transactionName: string, options?: ProposalOptions): Promise<Uint8Array>;

    /**
     * Create a proposal that can be sent to peers for endorsement. Supports off-line signing flow.
     * @param transactionName Name of the transaction to invoke.
     * @param options Transaction invocation options.
     * @returns A transaction proposal.
     */
    newProposal(transactionName: string, options?: ProposalOptions): Proposal;

    /**
     * Create a proposal with the specified digital signature. Supports off-line signing flow.
     * @param bytes Serialized proposal.
     * @param signature Digital signature.
     * @returns A signed proposal.
     */
    newSignedProposal(bytes: Uint8Array, signature: Uint8Array): Proposal;

    /**
     * Create a transaction with the specified digital signature. Supports off-line signing flow.
     * @param bytes Serialized proposal.
     * @param signature Digital signature.
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
        return this.submitSync(name, { arguments: args });
    }

    async evaluate(transactionName: string, options?: ProposalOptions): Promise<Uint8Array> {
        return this.newProposal(transactionName, options).evaluate();
    }

    async submitSync(transactionName: string, options?: ProposalOptions): Promise<Uint8Array> {
        const transaction = await this.newProposal(transactionName, options).endorse();
        const commit = await transaction.submit();
        const status = await commit.getStatus();
        if (status !== protos.TxValidationCode.VALID) {
            throw new Error(`Transaction ${transaction.getTransactionId()} failed to commit with status code ${status} (${protos.TxValidationCode[status]})`)
        }
        return transaction.getResult();
    }

    newProposal(transactionName: string, options: ProposalOptions = {}): Proposal {
        return new ProposalBuilder({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            channelName: this.#channelName,
            chaincodeId: this.#chaincodeId,
            transactionName: this.getQualifiedTransactionName(transactionName),
            options,
        }).build();
    }

    newSignedProposal(bytes: Uint8Array, signature: Uint8Array): Proposal {
        const proposedTransaction = gateway.ProposedTransaction.decode(bytes);

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
        const preparedTransaction = gateway.PreparedTransaction.decode(bytes);

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
