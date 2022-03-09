/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as grpc from '@grpc/grpc-js';
import { ChaincodeEvent, ChaincodeEventsOptions, connect, ConnectOptions, Contract, Gateway, Identity, Network, Signer } from '@hyperledger/fabric-gateway';
import { EventListener } from './eventlistener';
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
    #chaincodeEventListeners: Map<string, EventListener<ChaincodeEvent>> = new Map();

    constructor(identity: Identity, signer?: Signer, signerClose?: () => void) {
        this.#identity = identity;
        this.#signer = signer;
        this.#signerClose = signerClose;
    }

    connect(client: grpc.Client): void {
        this.#client = client;
        const options: ConnectOptions = {
            signer: this.#signer,
            identity: this.#identity,
            client,
        };
        this.#gateway = connect(options);
    }

    useNetwork(channelName: string): void {
        this.#network = this.getGateway().getNetwork(channelName);
        this.#contract = undefined;
    }

    useContract(contractName: string): void {
        this.#contract = this.getNetwork().getContract(contractName);
    }

    newTransaction(action: string, transactionName: string): TransactionInvocation {
        return new TransactionInvocation(action, this.getGateway(), this.getContract(), transactionName);
    }

    async listenForChaincodeEvents(listenerName: string, chaincodeName: string, options?: ChaincodeEventsOptions): Promise<void> {
        this.closeChaincodeEvents(listenerName);
        const events = await this.getNetwork().getChaincodeEvents(chaincodeName, options);
        const listener = new EventListener(events);
        this.#chaincodeEventListeners.set(listenerName, listener);
    }

    async nextChaincodeEvent(listenerName: string): Promise<ChaincodeEvent> {
        return await this.getChaincodeEventListener(listenerName).next();
    }

    close(): void {
        this.#chaincodeEventListeners.forEach(listener => listener.close());
        this.#gateway?.close();
        this.#client?.close();
        if (this.#signerClose) {
            this.#signerClose();
        }
    }

    closeChaincodeEvents(listenerName: string): void {
        this.#chaincodeEventListeners.get(listenerName)?.close();
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

    private getChaincodeEventListener(listenerName: string): EventListener<ChaincodeEvent> {
        return assertDefined(this.#chaincodeEventListeners.get(listenerName), `chaincodeEventListener: ${listenerName}`);
    }
}
