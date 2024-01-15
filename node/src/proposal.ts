/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CallOptions } from '@grpc/grpc-js';
import { common, gateway, peer } from '@hyperledger/fabric-protos';
import { GatewayClient } from './client';
import { assertDefined } from './gateway';
import { Signable } from './signable';
import { SigningIdentity } from './signingidentity';
import { Transaction, TransactionImpl } from './transaction';

/**
 * Proposal represents a transaction proposal that can be sent to peers for endorsement or evaluated as a query.
 */
export interface Proposal extends Signable {
    /**
     * Get the transaction ID for this proposal.
     */
    getTransactionId(): string;

    /**
     * Evaluate the transaction proposal and obtain its result, without updating the ledger. This runs the transaction
     * on a peer to obtain a transaction result, but does not submit the endorsed transaction to the orderer to be
     * committed to the ledger.
     * @param options - gRPC call options.
     * @returns The result returned by the transaction function.
     * @throws {@link GatewayError}
     * Thrown if the gRPC service invocation fails.
     */
    evaluate(options?: CallOptions): Promise<Uint8Array>;

    /**
     * Obtain endorsement for the transaction proposal from sufficient peers to allow it to be committed to the ledger.
     * @param options - gRPC call options.
     * @returns An endorsed transaction that can be submitted to the ledger.
     * @throws {@link EndorseError}
     * Thrown if the gRPC service invocation fails.
     */
    endorse(options?: CallOptions): Promise<Transaction>;
}

export interface ProposalImplOptions {
    client: GatewayClient;
    signingIdentity: SigningIdentity;
    channelName: string;
    proposedTransaction: gateway.ProposedTransaction;
}

export class ProposalImpl implements Proposal {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #channelName: string;
    readonly #proposedTransaction: gateway.ProposedTransaction;
    readonly #proposal: peer.SignedProposal;

    constructor(options: Readonly<ProposalImplOptions>) {
        this.#client = options.client;
        this.#signingIdentity = options.signingIdentity;
        this.#channelName = options.channelName;
        this.#proposedTransaction = options.proposedTransaction;
        this.#proposal = assertDefined(options.proposedTransaction.getProposal(), 'Missing signed proposal');
    }

    getBytes(): Uint8Array {
        return this.#proposedTransaction.serializeBinary();
    }

    getDigest(): Uint8Array {
        const bytes = this.#proposal.getProposalBytes_asU8();
        return this.#signingIdentity.hash(bytes);
    }

    getTransactionId(): string {
        return this.#proposedTransaction.getTransactionId();
    }

    async evaluate(options?: Readonly<CallOptions>): Promise<Uint8Array> {
        await this.#sign();
        const evaluateResponse = await this.#client.evaluate(this.#newEvaluateRequest(), options);
        const result = evaluateResponse.getResult();

        return result?.getPayload_asU8() ?? new Uint8Array(0);
    }

    async endorse(options?: Readonly<CallOptions>): Promise<Transaction> {
        await this.#sign();
        const endorseResponse = await this.#client.endorse(this.#newEndorseRequest(), options);

        const txEnvelope = assertDefined(endorseResponse.getPreparedTransaction(), 'Missing transaction envelope');

        return new TransactionImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            preparedTransaction: this.#newPreparedTransaction(txEnvelope),
        });
    }

    setSignature(signature: Uint8Array): void {
        this.#proposal.setSignature(signature);
    }

    async #sign(): Promise<void> {
        if (this.#isSigned()) {
            return;
        }

        const signature = await this.#signingIdentity.sign(this.getDigest());
        this.setSignature(signature);
    }

    #isSigned(): boolean {
        const signatureLength = this.#proposal.getSignature_asU8().length;
        return signatureLength > 0;
    }

    #newEvaluateRequest(): gateway.EvaluateRequest {
        const result = new gateway.EvaluateRequest();
        result.setTransactionId(this.#proposedTransaction.getTransactionId());
        result.setChannelId(this.#channelName);
        result.setProposedTransaction(this.#proposal);
        result.setTargetOrganizationsList(this.#proposedTransaction.getEndorsingOrganizationsList());
        return result;
    }

    #newEndorseRequest(): gateway.EndorseRequest {
        const result = new gateway.EndorseRequest();
        result.setTransactionId(this.#proposedTransaction.getTransactionId());
        result.setChannelId(this.#channelName);
        result.setProposedTransaction(this.#proposal);
        result.setEndorsingOrganizationsList(this.#proposedTransaction.getEndorsingOrganizationsList());
        return result;
    }

    #newPreparedTransaction(envelope: common.Envelope): gateway.PreparedTransaction {
        const result = new gateway.PreparedTransaction();
        result.setEnvelope(envelope);
        result.setTransactionId(this.#proposedTransaction.getTransactionId());
        return result;
    }
}
