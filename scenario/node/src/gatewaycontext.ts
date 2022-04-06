/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as grpc from '@grpc/grpc-js';
import { BlockEventsOptions, ChaincodeEvent, ChaincodeEventsOptions, Checkpointer, connect, ConnectOptions, Contract, Gateway, Identity, Network, Signer } from '@hyperledger/fabric-gateway';
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
    #blockEventListeners: Map<string, EventListener<unknown>> = new Map();
    #filteredBlockEventListeners: Map<string, EventListener<unknown>> = new Map();
    #blockAndPrivateDataEventListeners: Map<string, EventListener<unknown>> = new Map();
    #lastEventReceived?: ChaincodeEvent;

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
        const event: ChaincodeEvent = await this.getChaincodeEventListener(listenerName).next();
        this.#lastEventReceived = event;
        return event;
    }

    async checkPointBlock(listenerName: string, checkpointer: Checkpointer,): Promise<void> {
        await this.getChaincodeEventListener(listenerName).checkpointBlock(checkpointer, this.getLastEventReceived().blockNumber);
    }

    async listenForBlockEvents(listenerName: string, options?: BlockEventsOptions): Promise<void> {
        this.closeBlockEvents(listenerName);
        const events = await this.getNetwork().getBlockEvents(options);
        const listener = new EventListener(events);
        this.#blockEventListeners.set(listenerName, listener);
    }

    async nextBlockEvent(listenerName: string): Promise<unknown> {
        return await this.getBlockEventListener(listenerName).next();
    }

    async listenForFilteredBlockEvents(listenerName: string, options?: BlockEventsOptions): Promise<void> {
        this.closeFilteredBlockEvents(listenerName);
        const events = await this.getNetwork().getFilteredBlockEvents(options);
        const listener = new EventListener(events);
        this.#filteredBlockEventListeners.set(listenerName, listener);
    }

    async nextFilteredBlockEvent(listenerName: string): Promise<unknown> {
        return await this.getFilteredBlockEventListener(listenerName).next();
    }

    async listenForBlockAndPrivateDataEvents(listenerName: string, options?: BlockEventsOptions): Promise<void> {
        this.closeBlockEvents(listenerName);
        const events = await this.getNetwork().getBlockAndPrivateDataEvents(options);
        const listener = new EventListener(events);
        this.#blockAndPrivateDataEventListeners.set(listenerName, listener);
    }

    async nextBlockAndPrivateDataEvent(listenerName: string): Promise<unknown> {
        return await this.getBlockAndPrivateDataEventListener(listenerName).next();
    }

    close(): void {
        this.#chaincodeEventListeners.forEach(listener => listener.close());
        this.#blockEventListeners.forEach(listener => listener.close());
        this.#filteredBlockEventListeners.forEach(listener => listener.close());
        this.#blockAndPrivateDataEventListeners.forEach(listener => listener.close());
        this.#gateway?.close();
        this.#client?.close();
        if (this.#signerClose) {
            this.#signerClose();
        }
    }

    closeChaincodeEvents(listenerName: string): void {
        this.#chaincodeEventListeners.get(listenerName)?.close();
        this.#chaincodeEventListeners.delete(listenerName);
    }

    closeBlockEvents(listenerName: string): void {
        this.#blockEventListeners.get(listenerName)?.close();
        this.#blockEventListeners.delete(listenerName);
    }

    closeFilteredBlockEvents(listenerName: string): void {
        this.#filteredBlockEventListeners.get(listenerName)?.close();
        this.#filteredBlockEventListeners.delete(listenerName);
    }

    closeBlockAndPrivateDataEvents(listenerName: string): void {
        this.#blockAndPrivateDataEventListeners.get(listenerName)?.close();
        this.#blockAndPrivateDataEventListeners.delete(listenerName);
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

    private getBlockEventListener(listenerName: string): EventListener<unknown> {
        return assertDefined(this.#blockEventListeners.get(listenerName), `blockEventListener: ${listenerName}`);
    }

    private getFilteredBlockEventListener(listenerName: string): EventListener<unknown> {
        return assertDefined(this.#filteredBlockEventListeners.get(listenerName), `filteredBlockEventListener: ${listenerName}`);
    }

    private getBlockAndPrivateDataEventListener(listenerName: string): EventListener<unknown> {
        return assertDefined(this.#blockAndPrivateDataEventListeners.get(listenerName), `blockAndPrivateDataEventListener: ${listenerName}`);
    }

    private getLastEventReceived(): ChaincodeEvent {
        return assertDefined(this.#lastEventReceived, 'event');
    }
}
