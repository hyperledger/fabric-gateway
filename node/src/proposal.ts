/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { GatewayClient } from './client';
import { protos } from './protos/protos';
import { SigningIdentity } from './signingidentity';
import { Transaction, TransactionImpl } from './transaction';
import * as util from 'util';

export interface Proposal {
    /**
     * Get the serialized proposal message.
     */
    getBytes(): Uint8Array;

    /**
     * Get the digest of the proposal. This is used to generate a digital signature.
     */
    getDigest(): Uint8Array;

    /**
     * Get the transaction ID for this proposal.
     */
    getTransactionId(): string;

    /**
     * Evaluate the transaction proposal and obtain its result, without updating the ledger. This runs the transaction
     * on a peer to obtain a transaction result, but does not submit the endorsed transaction to the orderer to be
     * committed to the ledger.
     */
    evaluate(): Promise<Uint8Array>;

    /**
     * Obtain endorsement for the transaction proposal from sufficient peers to allow it to be committed to the ledger.
     */
    endorse(): Promise<Transaction>;
}

export interface ProposalImplOptions {
    readonly client: GatewayClient;
    readonly signingIdentity: SigningIdentity;
    readonly proposedTransaction: protos.IProposedTransaction;
}

export class ProposalImpl implements Proposal {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #proposedTransaction: protos.IProposedTransaction;
    readonly #proposal: protos.ISignedProposal;

    constructor(options: ProposalImplOptions) {
        this.#client = options.client;
        this.#signingIdentity = options.signingIdentity;
        this.#proposedTransaction = options.proposedTransaction;

        const proposal = options.proposedTransaction.proposal;
        if (!proposal) {
            throw new Error(`Proposal not defined: ${util.inspect(options.proposedTransaction)}`);
        }
        this.#proposal = proposal;
    }

    getBytes(): Uint8Array {
        return protos.ProposedTransaction.encode(this.#proposedTransaction).finish();
    }

    getDigest(): Uint8Array {
        const bytes = this.#proposal.proposal_bytes;
        if (!bytes) {
            throw new Error(`Proposal bytes not defined: ${util.inspect(this.#proposal)}`)
        }
        return this.#signingIdentity.hash(bytes);
    }

    getTransactionId(): string {
        const transactionId = this.#proposedTransaction.txId;
        if (typeof transactionId !== 'string') {
            throw new Error(`Transaction ID not defined: ${util.inspect(this.#proposedTransaction)}`);
        }
        return transactionId;
    }

    async evaluate(): Promise<Uint8Array> {
        await this.sign();
        const result = await this.#client.evaluate(this.#proposedTransaction);
        return result.value || new Uint8Array(0);
    }

    async endorse(): Promise<Transaction> {
        await this.sign();
        const preparedTransaction = await this.#client.endorse(this.#proposedTransaction);

        return new TransactionImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            preparedTransaction,
        });
    }

    setSignature(signature: Uint8Array): void {
        this.#proposal.signature = signature;
    }

    private async sign(): Promise<void> {
        if (this.isSigned()) {
            return;
        }

        const signature = await this.#signingIdentity.sign(this.getDigest());
        this.setSignature(signature);
    }

    private isSigned(): boolean {
        const signatureLength = this.#proposal.signature?.length ?? 0;
        return signatureLength > 0;
    }
}
