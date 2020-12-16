/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Proposal } from "proposal";
import { GatewayClient } from "./client";
import { ProposalBuilder, ProposalOptions } from "./proposalbuilder";
import { SigningIdentity } from "./signingidentity";

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
     * Create a proposal that can be sent to peers for endorsement. Supports off-line signing transaction flow.
     * @param transactionName Name of the transaction to invoke.
     * @param options Transaction invocation options.
     * @returns A transaction proposal.
     */
    newProposal(transactionName: string, options?: ProposalOptions): Proposal;

    // newSignedProposal(bytes: Uint8Array, signature: Uint8Array): Proposal;
    // newSignedTransaction(bytes: Uint8Array, signature: Uint8Array): Transaction;
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

    async evaluateTransaction(name: string, ...args: string[]): Promise<Uint8Array> {
        return this.evaluate(name, { arguments: args });
    }

    async submitTransaction(name: string, ...args: string[]): Promise<Uint8Array> {
        return this.submitSync(name, { arguments: args });
    }

    async evaluate(transactionName: string, options?: ProposalOptions): Promise<Uint8Array> {
        return this.newProposal(transactionName, options).evaluate();
    }

    async submitSync(transactionName: string, options?: ProposalOptions): Promise<Uint8Array> {
        const transaction = await this.newProposal(transactionName, options).endorse();
        return await transaction.submit();
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

    private getQualifiedTransactionName(transactionName: string) {
        return this.#contractName ? `${this.#contractName}:${transactionName}` : transactionName;
    }
}
