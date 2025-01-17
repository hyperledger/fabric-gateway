/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ProposedTransaction } from '@hyperledger/fabric-protos/lib/gateway/gateway_pb';
import { SignedProposal } from '@hyperledger/fabric-protos/lib/peer/proposal_pb';
import { assertDefined } from './gateway';
import { SigningIdentity } from './signingidentity';

/**
 * Proposal represents a transaction proposal that can be sent to peers for endorsement or evaluated as a query.
 */
export interface Proposal {
    /**
     * Get the serialized bytes of the object. This is used to transfer the object state to a remote service.
     */
    getBytes(): Uint8Array;

    /**
     * Get the digest of the signable object. This is used to generate a digital signature.
     */
    getDigest(): Promise<Uint8Array>;

    /**
     * Get the transaction ID for this proposal.
     */
    getTransactionId(): string;
}

export interface ProposalImplOptions {
    signingIdentity: SigningIdentity;
    proposedTransaction: ProposedTransaction;
}

export class ProposalImpl implements Proposal {
    readonly #signingIdentity: SigningIdentity;
    readonly #proposedTransaction: ProposedTransaction;
    readonly #proposal: SignedProposal;

    static async newInstance(options: Readonly<ProposalImplOptions>): Promise<ProposalImpl> {
        const result = new ProposalImpl(options);
        await result.#sign();
        return result;
    }

    private constructor(options: Readonly<ProposalImplOptions>) {
        this.#signingIdentity = options.signingIdentity;
        this.#proposedTransaction = options.proposedTransaction;
        this.#proposal = assertDefined(options.proposedTransaction.getProposal(), 'Missing signed proposal');
    }

    getBytes(): Uint8Array {
        return this.#proposedTransaction.serializeBinary();
    }

    async getDigest(): Promise<Uint8Array> {
        const bytes = this.#proposal.getProposalBytes_asU8();
        return this.#signingIdentity.hash(bytes);
    }

    getTransactionId(): string {
        return this.#proposedTransaction.getTransactionId();
    }

    async #sign(): Promise<void> {
        if (this.#isSigned()) {
            return;
        }

        const signature = await this.#signingIdentity.sign(await this.getDigest());
        this.#setSignature(signature);
    }

    #isSigned(): boolean {
        const signatureLength = this.#proposal.getSignature_asU8().length;
        return signatureLength > 0;
    }

    #setSignature(signature: Uint8Array): void {
        this.#proposal.setSignature(signature);
    }
}
