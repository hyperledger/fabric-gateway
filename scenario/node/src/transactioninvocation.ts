/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Contract, Network, ProposalOptions, Signable, Signer } from 'fabric-gateway';
import { bytesAsString, toError, toString } from './utils';

export class TransactionInvocation {
    readonly options: ProposalOptions = {};

    private readonly name: string;
    private readonly network: Network;
    private readonly contract: Contract;
    private readonly invoke: () => Promise<Uint8Array>;
    private offlineSigner?: Signer;
    private result?: Uint8Array;
    private error?: Error;

    constructor(type: string, network: Network, contract: Contract, name: string) {
        this.network = network;
        this.contract = contract;
        this.name = name;
        this.invoke = this.getInvoke(type).bind(this);
    }

    async invokeTransaction(): Promise<void> {
        try {
            this.result = await this.invoke();
        } catch (error) {
            this.error = toError(error);
        }
    }

    setOfflineSigner(signer: Signer): void {
        this.offlineSigner = signer;
    }

    getResult(): string {
        if (!this.result) {
            throw new Error(`No transaction result. Error is: ${toString(this.error)}`);
        }

        return bytesAsString(this.result);
    }

    getError(): Error {
        if (!this.error) {
            throw new Error(`No transaction error. Result is: ${bytesAsString(this.result)}`)
        }

        return this.error;
    }

    private getInvoke(type: string): () => Promise<Uint8Array> {
        if (type === 'evaluate') {
            return this.evaluate;
        }
        if (type === 'submit') {
            return this.submit;
        }
        throw new Error(`Unknown invocation type: ${type}`);
    }

    private async evaluate(): Promise<Uint8Array> {
        let proposal = this.contract.newProposal(this.name, this.options);
        proposal = await this.sign(proposal, this.contract.newSignedProposal.bind(this.contract));
 
        return await proposal.evaluate();
    }
    
    private async submit(): Promise<Uint8Array> {
        const unsignedProposal = this.contract.newProposal(this.name, this.options);
        const signedProposal = await this.sign(unsignedProposal, this.contract.newSignedProposal.bind(this.contract));
        
        const unsignedTransaction = await signedProposal.endorse();
        const signedTransaction = await this.sign(unsignedTransaction, this.contract.newSignedTransaction.bind(this.contract));
    
        const submitted = await signedTransaction.submit();
        const signedCommit = await this.sign(submitted, this.network.newSignedCommit.bind(this.network));

        const successful = await signedCommit.isSuccessful();
        if (!successful) {
            const status = await signedCommit.getStatus();
            throw new Error(`Transaction commit failed with status: ${status}`)
        }

        return submitted.getResult();
    }

    private async sign<T extends Signable>(signable: T, newInstance: (bytes: Uint8Array, signature: Uint8Array) => T): Promise<T> {
        if (!this.offlineSigner) {
            return signable;
        }

        const signature = await this.offlineSigner(signable.getDigest());
        return newInstance(signable.getBytes(), signature);
    }
}
