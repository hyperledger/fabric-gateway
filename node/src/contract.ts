/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { GatewayClient } from "client";
import { SigningIdentity } from "signingidentity";
import { OldTransaction } from "./transaction";

export interface Contract {
    getChaincodeId(): string;
    getContractName(): string | undefined;
    evaluateTransaction(name: string, ...args: Array<string|Uint8Array>): Promise<Uint8Array>;
    submitTransaction(name: string, ...args: Array<string|Uint8Array>): Promise<Uint8Array>;

    // evaluate(/* ProposalOptions */): Promise<Uint8Array>;
    // submit(/* ProposalOptions */): Promise<Uint8Array>;

    // newProposal(/* ProposalOptions */): Proposal;
    // newSignedProposal(bytes: Uint8Array, signature: Uint8Array): Proposal;
    // newSignedTransaction(bytes: Uint8Array, signature: Uint8Array): Transaction;

    // TODO: Remove
    createTransaction(transactionName: string): OldTransaction;
    prepareToEvaluate(transactionName: string): EvaluateTransaction;
    prepareToSubmit(transactionName: string): SubmitTransaction;
}

export class ContractImpl implements Contract {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #channelName: string;
    readonly #chaincodeId: string;
    readonly #contractName?: string;

    constructor(client: GatewayClient, signingIdentity: SigningIdentity, channelName: string, chaincodeId: string, contractName?: string) {
        this.#client = client;
        this.#signingIdentity = signingIdentity;
        this.#channelName = channelName;
        this.#chaincodeId = chaincodeId;
        this.#contractName = contractName;
    }

    getChaincodeId(): string {
        return this.#chaincodeId;
    }

    getContractName(): string | undefined {
        return this.#contractName;
    }

    createTransaction(transactionName: string): OldTransaction {
        return new OldTransaction(this.#client, this.#signingIdentity, this.#channelName, this.#chaincodeId, this.getQualifiedTransactionName(transactionName));
    }

    async evaluateTransaction(name: string, ...args: string[]): Promise<Uint8Array> {
        return this.createTransaction(name).evaluate(...args);
    }

    async submitTransaction(name: string, ...args: string[]): Promise<Uint8Array> {
        return this.createTransaction(name).submit(...args);
    }

    prepareToEvaluate(transactionName: string): EvaluateTransaction {
        return new EvaluateTransaction(this.#client, this.#signingIdentity, this.#channelName, this.#chaincodeId, this.getQualifiedTransactionName(transactionName));
    }

    prepareToSubmit(transactionName: string): SubmitTransaction {
        return new SubmitTransaction(this.#client, this.#signingIdentity, this.#channelName, this.#chaincodeId, this.getQualifiedTransactionName(transactionName));
    }

    private getQualifiedTransactionName(transactionName: string) {
        return this.#contractName ? `${this.#contractName}:${transactionName}` : transactionName;
    }
}

class EvaluateTransaction extends OldTransaction {
    private args: string[];

    constructor(client: GatewayClient, signingIdentity: SigningIdentity, channelName: string, chaincodeId: string, transactionName: string) {
        super(client, signingIdentity, channelName, chaincodeId, transactionName);
        this.args = [];
    }

    setArgs(...args: string[]) {
        this.args = args;
        return this;
    }

    async invoke() {
        return this.evaluate(...this.args);
    }
}

class SubmitTransaction extends OldTransaction {
    private args: string[];

    constructor(client: GatewayClient, signingIdentity: SigningIdentity, channelName: string, chaincodeId: string, transactionName: string) {
        super(client, signingIdentity, channelName, chaincodeId, transactionName);
        this.args = [];
    }

    setArgs(...args: string[]) {
        this.args = args;
        return this;
    }

    async invoke() {
        return this.submit(...this.args);
    }
}
