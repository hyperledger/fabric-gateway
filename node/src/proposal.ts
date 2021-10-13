/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { inspect } from 'util';
import { GatewayClient } from './client';
import { Envelope } from './protos/common/common_pb';
import { EndorseRequest, EvaluateRequest, PreparedTransaction, ProposedTransaction } from './protos/gateway/gateway_pb';
import { SignedProposal } from './protos/peer/proposal_pb';
import { Response } from './protos/peer/proposal_response_pb';
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
     * @returns The result returned by the transaction function.
     * @throws {@link GatewayError}
     * Thrown if the gRPC service invocation fails.
     */
    evaluate(): Promise<Uint8Array>;

    /**
     * Obtain endorsement for the transaction proposal from sufficient peers to allow it to be committed to the ledger.
     * @returns An endorsed transaction that can be submitted to the ledger.
     * @throws {@link GatewayError}
     * Thrown if the gRPC service invocation fails.
     */
    endorse(): Promise<Transaction>;
}

export interface ProposalImplOptions {
    readonly client: GatewayClient;
    readonly signingIdentity: SigningIdentity;
    readonly channelName: string;
    readonly proposedTransaction: ProposedTransaction;
}

export class ProposalImpl implements Proposal {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #channelName: string;
    readonly #proposedTransaction: ProposedTransaction;
    readonly #proposal: SignedProposal;

    constructor(options: ProposalImplOptions) {
        this.#client = options.client;
        this.#signingIdentity = options.signingIdentity;
        this.#channelName = options.channelName;
        this.#proposedTransaction = options.proposedTransaction;

        const proposal = options.proposedTransaction.getProposal();
        if (!proposal) {
            throw new Error(`Proposal not defined: ${inspect(options.proposedTransaction)}`);
        }
        this.#proposal = proposal;
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

    async evaluate(): Promise<Uint8Array> {
        await this.sign();
        const evaluateResponse = await this.#client.evaluate(this.newEvaluateRequest());
        const result = evaluateResponse.getResult();

        return result?.getPayload_asU8() || new Uint8Array(0);
    }

    async endorse(): Promise<Transaction> {
        await this.sign();
        const endorseResponse = await this.#client.endorse(this.newEndorseRequest());

        const preparedTx = endorseResponse.getPreparedTransaction();
        const response = endorseResponse.getResult();
        if (!preparedTx || !response) {
            throw new Error(`Invalid endorsement response: ${inspect(endorseResponse)}`)
        }

        return new TransactionImpl({
            client: this.#client,
            signingIdentity: this.#signingIdentity,
            channelName: this.#channelName,
            preparedTransaction: this.newPreparedTransaction(preparedTx, response)
        });
    }

    setSignature(signature: Uint8Array): void {
        this.#proposal.setSignature(signature);
    }

    private async sign(): Promise<void> {
        if (this.isSigned()) {
            return;
        }

        const signature = await this.#signingIdentity.sign(this.getDigest());
        this.setSignature(signature);
    }

    private isSigned(): boolean {
        const signatureLength = this.#proposal.getSignature_asU8().length;
        return signatureLength > 0;
    }

    private newEvaluateRequest(): EvaluateRequest {
        const result = new EvaluateRequest();
        result.setTransactionId(this.#proposedTransaction.getTransactionId());
        result.setChannelId(this.#channelName);
        result.setProposedTransaction(this.#proposal);
        result.setTargetOrganizationsList(this.#proposedTransaction.getEndorsingOrganizationsList());
        return result;
    }

    private newEndorseRequest(): EndorseRequest {
        const result = new EndorseRequest();
        result.setTransactionId(this.#proposedTransaction.getTransactionId())
        result.setChannelId(this.#channelName);
        result.setProposedTransaction(this.#proposal);
        result.setEndorsingOrganizationsList(this.#proposedTransaction.getEndorsingOrganizationsList());
        return result;
    }

    private newPreparedTransaction(envelope: Envelope, response: Response): PreparedTransaction {
        const result = new PreparedTransaction();
        result.setEnvelope(envelope);
        result.setResult(response);
        result.setTransactionId(this.#proposedTransaction.getTransactionId());
        return result;
    }
}
