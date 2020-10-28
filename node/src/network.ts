/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { Gateway } from "./gateway";
import { Contract } from "./contract"

export class Network {
    private readonly name: string;
    readonly _gateway: Gateway;

    constructor(name: string, gateway: Gateway) {
        this.name = name;
        this._gateway = gateway;
    }

    getName() {
        return this.name;
    }

    getContract(contractName: string) {
        return new Contract(contractName, this);
    }

}