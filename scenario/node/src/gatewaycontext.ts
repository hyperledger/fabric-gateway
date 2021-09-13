/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as grpc from '@grpc/grpc-js';
import { ChaincodeEvent, connect, ConnectOptions, Contract, Gateway, Identity, Network, Signer } from 'fabric-gateway';
import { TransactionInvocation } from './transactioninvocation';
import { assertDefined } from './utils';

export class GatewayContext {
    readonly #identity: Identity;
    readonly #signer?: Signer;
    readonly #signerClose?: () => void;
    #client?: grpc.Client;
    #gateway?: Gateway;
    #network?: Network;
    #contract?: Contract;
    #chaincodeEvents?: AsyncIterator<ChaincodeEvent>;

    constructor(identity: Identity, signer?: Signer, signerClose?: () => void) {
        this.#identity = identity;
        this.#signer = signer;
        this.#signerClose = signerClose;
    }

    async connect(client: grpc.Client): Promise<void> {
        this.#client = client;
        const options: ConnectOptions = {
            signer: this.#signer,
            identity: this.#identity,
            client,
        };
        this.#gateway = await connect(options);
    }

    useNetwork(channelName: string): void {
        this.#network = this.getGateway().getNetwork(channelName);
        this.#contract = undefined;
    }

    useContract(contractName: string): void {
        this.#contract = this.getNetwork().getContract(contractName);
    }

    newTransaction(action: string, transactionName: string): TransactionInvocation {
        return new TransactionInvocation(action, this.getNetwork(), this.getContract(), transactionName);
    }

    async listenForChaincodeEvents(chaincodeId: string): Promise<void> {
        const events = await this.getNetwork().getChaincodeEvents(chaincodeId);
        this.#chaincodeEvents = events[Symbol.asyncIterator]();
    }

    async nextChaincodeEvent(): Promise<ChaincodeEvent> {
        const result = await this.getChaincodeEvents().next();
        return result.value;
    }

    close(): void {
        this.#gateway?.close();
        this.#client?.close();
        if (this.#signerClose) {
            this.#signerClose();
        }
    }

    private getGateway(): Gateway {
        return assertDefined(this.#gateway, 'gateway');
    }

    private getNetwork(): Network {
        return assertDefined(this.#network, 'network');
    }

    private getContract(): Contract {
        return assertDefined(this.#contract, 'contract');
    }

    private getChaincodeEvents(): AsyncIterator<ChaincodeEvent> {
        return assertDefined(this.#chaincodeEvents, 'chaincodeEvents');
    }
}
