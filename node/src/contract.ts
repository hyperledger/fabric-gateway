/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 *SPDX-License-Identifier: Apache-2.0
 */

import { Client } from "impl/client";
import { SigningIdentity } from "signingidentity";
import { Transaction } from "./transaction";

export interface Contract {
    getChaincodeId(): string;
    getContractName(): string | undefined;
    evaluateTransaction(name: string, ...args: Array<string|Uint8Array>): Promise<string>; // TODO: wrong return type
    submitTransaction(name: string, ...args: Array<string|Uint8Array>): Promise<string>; // TODO: wrong return type
    createTransaction(transactionName: string): Transaction; // TODO: refactor to align with Go and Java
    prepareToEvaluate(transactionName: string): EvaluateTransaction; // TODO: refactor to align with Go and Java
    prepareToSubmit(transactionName: string): SubmitTransaction; // TODO: refactor to align with Go and Java
}

export class ContractImpl implements Contract {
    readonly #client: Client;
    readonly #signingIdentity: SigningIdentity;
    readonly #channelName: string;
    readonly #chaincodeId: string;
    readonly #contractName?: string;

    constructor(client: Client, signingIdentity: SigningIdentity, channelName: string, chaincodeId: string, contractName?: string) {
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

    createTransaction(transactionName: string): Transaction {
        return new Transaction(this.#client, this.#signingIdentity, this.#channelName, this.#chaincodeId, this.getQualifiedTransactionName(transactionName));
    }

    async evaluateTransaction(name: string, ...args: string[]): Promise<string> {
        return this.createTransaction(name).evaluate(...args);
    }

    async submitTransaction(name: string, ...args: string[]): Promise<string> {
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

class EvaluateTransaction extends Transaction {
    private args: string[];

    constructor(client: Client, signingIdentity: SigningIdentity, channelName: string, chaincodeId: string, transactionName: string) {
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

class SubmitTransaction extends Transaction {
    private args: string[];

    constructor(client: Client, signingIdentity: SigningIdentity, channelName: string, chaincodeId: string, transactionName: string) {
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
