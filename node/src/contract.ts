/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { Network } from "./network";
import { Transaction } from "./transaction";

export class Contract {
    private readonly name: string;
    readonly _network: Network;

    constructor(name: string, network: Network) {
        this.name = name;
        this._network = network;
    }

    getName() {
        return this.name;
    }

    createTransaction(transactionName: string) {
        return new Transaction(transactionName, this);
    }

    async evaluateTransaction(name: string, ...args: string[]) {
        return this.createTransaction(name).evaluate(...args);
    }

    async submitTransaction(name: string, ...args: string[]) {
        return this.createTransaction(name).submit(...args);
    }

    prepareToEvaluate(transactionName: string) {
        return new EvaluateTransaction(transactionName, this);
    }

    prepareToSubmit(transactionName: string) {
        return new SubmitTransaction(transactionName, this);
    }
}

class EvaluateTransaction extends Transaction {
    private args: string[];

    constructor(name: string, contract: Contract) {
        super(name, contract);
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

    constructor(name: string, contract: Contract) {
        super(name, contract);
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
