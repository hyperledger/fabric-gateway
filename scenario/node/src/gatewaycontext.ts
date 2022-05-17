/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as grpc from '@grpc/grpc-js';
import { BlockEventsOptions, ChaincodeEvent, ChaincodeEventsOptions, Checkpointer, checkpointers, connect, ConnectOptions, Contract, Gateway, Identity, Network, Signer } from '@hyperledger/fabric-gateway';
import { CheckpointEventListener } from './checkpointeventlistener';
import { BaseEventListener } from './baseeventlistener';
import { TransactionInvocation } from './transactioninvocation';
import { assertDefined } from './utils';
import { common, peer } from '@hyperledger/fabric-protos';
import { EventListener } from './eventlistener';

export class GatewayContext {
    readonly #identity: Identity;
    readonly #signer?: Signer;
    readonly #signerClose?: () => void;
    #client?: grpc.Client;
    #gateway?: Gateway;
    #network?: Network;
    #contract?: Contract;
    #checkpointer?: Checkpointer;
    readonly #chaincodeEventListeners: Map<string, EventListener<ChaincodeEvent>> = new Map();
    readonly #blockEventListeners: Map<string, EventListener<common.Block>> = new Map();
    readonly #filteredBlockEventListeners: Map<string, EventListener<peer.FilteredBlock>> = new Map();
    readonly #blockAndPrivateDataEventListeners: Map<string, EventListener<peer.BlockAndPrivateData>> = new Map();

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

    createCheckpointer(): void {
        this.#checkpointer = checkpointers.inMemory();
    }

    getCheckpointer(): Checkpointer {
        return assertDefined(this.#checkpointer, 'checkpointer');
    }

    async listenForChaincodeEvents(listenerName: string, chaincodeName: string, options?: ChaincodeEventsOptions): Promise<void> {
        this.closeChaincodeEvents(listenerName);
        const events = await this.getNetwork().getChaincodeEvents(chaincodeName, options);
        const listener = new BaseEventListener(events);
        this.#chaincodeEventListeners.set(listenerName, listener);
    }

    async listenForChaincodeEventsUsingCheckpointer(listenerName: string, chaincodeName: string, options?: ChaincodeEventsOptions): Promise<void> {
        this.closeChaincodeEvents(listenerName);
        const events = await this.getNetwork().getChaincodeEvents(chaincodeName, options);
        const listener = new CheckpointEventListener<ChaincodeEvent>(events, async (event: ChaincodeEvent): Promise <void> => {
            await this.getCheckpointer().checkpointChaincodeEvent(event);});
        this.#chaincodeEventListeners.set(listenerName, listener);
    }

    async nextChaincodeEvent(listenerName: string): Promise<ChaincodeEvent> {
        const event: ChaincodeEvent = await this.getChaincodeEventListener(listenerName).next();
        return event;
    }

    async listenForBlockEvents(listenerName: string, options?: BlockEventsOptions): Promise<void> {
        this.closeBlockEvents(listenerName);
        const events = await this.getNetwork().getBlockEvents(options);
        const listener = new BaseEventListener<common.Block>(events);
        this.#blockEventListeners.set(listenerName, listener);
    }

    async listenForBlockEventsUsingCheckpointer(listenerName: string, options?: BlockEventsOptions): Promise<void> {
        this.closeBlockEvents(listenerName);
        const events = await this.getNetwork().getBlockEvents(options);
        const listener = new CheckpointEventListener<common.Block>(events, async (event: common.Block): Promise<void> => {
            const header = assertDefined(event.getHeader(), 'block header');
            const blockNumber = header.getNumber();
            await this.getCheckpointer().checkpointBlock(BigInt(blockNumber));
        });
        this.#blockEventListeners.set(listenerName, listener);
    }

    async nextBlockEvent(listenerName: string): Promise<unknown> {
        return await this.getBlockEventListener(listenerName).next();
    }

    async listenForFilteredBlockEvents(listenerName: string, options?: BlockEventsOptions): Promise<void> {
        this.closeFilteredBlockEvents(listenerName);
        const events = await this.getNetwork().getFilteredBlockEvents(options);
        const listener = new BaseEventListener<peer.FilteredBlock>(events);
        this.#filteredBlockEventListeners.set(listenerName, listener);
    }

    async listenForFilteredBlockEventsUsingCheckpointer(listenerName: string, options?: BlockEventsOptions): Promise<void> {
        this.closeFilteredBlockEvents(listenerName);
        const events = await this.getNetwork().getFilteredBlockEvents(options);
        const listener = new CheckpointEventListener<peer.FilteredBlock>(events, async (event: peer.FilteredBlock): Promise<void> => {
            const blockNumber = event.getNumber();
            await this.getCheckpointer().checkpointBlock(BigInt(blockNumber));
        });
        this.#filteredBlockEventListeners.set(listenerName, listener);
    }

    async nextFilteredBlockEvent(listenerName: string): Promise<unknown> {
        const event = await this.getFilteredBlockEventListener(listenerName).next();
        return event;
    }
    async listenForBlockAndPrivateDataEvents(listenerName: string, options?: BlockEventsOptions): Promise<void> {
        this.closeBlockAndPrivateDataEvents(listenerName);
        const events = await this.getNetwork().getBlockAndPrivateDataEvents(options);
        const listener = new BaseEventListener<peer.BlockAndPrivateData>(events);
        this.#blockAndPrivateDataEventListeners.set(listenerName, listener);
    }

    async listenForBlockAndPrivateDataEventsUsingCheckpointer(listenerName: string, options?: BlockEventsOptions): Promise<void> {
        this.closeBlockAndPrivateDataEvents(listenerName);
        const events = await this.getNetwork().getBlockAndPrivateDataEvents(options);
        const listener = new CheckpointEventListener<peer.BlockAndPrivateData>(events, async (event: peer.BlockAndPrivateData): Promise<void>  => {
            const block = assertDefined(event.getBlock(), 'block');
            const header = assertDefined(block.getHeader(), 'block header');
            const blockNumber = header.getNumber();
            await this.getCheckpointer().checkpointBlock(BigInt(blockNumber));
        });
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
}
