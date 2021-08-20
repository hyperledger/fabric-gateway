/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { GatewayClient } from './client';
import { Proposal, ProposalImpl } from './proposal';
import { common, protos } from './protos/protos';
import { SigningIdentity } from './signingidentity';
import { TransactionContext } from './transactioncontext';

/**
 * Options used when evaluating or endorsing a transaction proposal.
 */
export interface ProposalOptions {
    /**
     * Arguments passed to the transaction function.
     */
    arguments?: (string | Uint8Array)[];
    
    /**
     * Private data passed to the transaction function but not recorded on the ledger.
     */
    transientData?: Record<string, string | Uint8Array>;

    /**
     * Specifies the set of organizations that will attempt to endorse the proposal.
     * No other organizations' peers will be sent this proposal.
     * This is usually used in conjunction with transientData for private data scenarios.
     */
    endorsingOrganizations?: string[];
}

export interface ProposalBuilderOptions {
    readonly client: GatewayClient;
    readonly signingIdentity: SigningIdentity;
    readonly channelName: string;
    readonly chaincodeId: string;
    readonly transactionName: string;
    readonly options: ProposalOptions;
}

export class ProposalBuilder {
    readonly #options: ProposalBuilderOptions;
    readonly #transactionContext: TransactionContext;

    constructor(options: ProposalBuilderOptions) {
        this.#options = options;
        this.#transactionContext = new TransactionContext(options.signingIdentity);
    }

    build(): Proposal {
        return new ProposalImpl({
            client: this.#options.client,
            signingIdentity: this.#options.signingIdentity,
            channelName: this.#options.channelName,
            proposedTransaction: {
                proposal: {
                    proposal_bytes: protos.Proposal.encode(this.newProposal()).finish(),
                },
                transaction_id: this.#transactionContext.getTransactionId(),
                endorsing_organizations: this.#options.options.endorsingOrganizations,
            },
        });
    }

    private newProposal(): protos.IProposal {
        return {
            header: common.Header.encode(this.newHeader()).finish(),
            payload: protos.ChaincodeProposalPayload.encode(this.newChaincodeProposalPayload()).finish(),
        };
    }

    private newHeader(): common.IHeader {
        return {
            channel_header: common.ChannelHeader.encode(this.newChannelHeader()).finish(),
            signature_header: common.SignatureHeader.encode(this.#transactionContext.getSignatureHeader()).finish(),
        };
    }

    private newChannelHeader(): common.IChannelHeader {
        return {
            type: common.HeaderType.ENDORSER_TRANSACTION,
            tx_id: this.#transactionContext.getTransactionId(),
            timestamp: {
                seconds: Date.now() / 1000,
            },
            channel_id: this.#options.channelName,
            extension: protos.ChaincodeHeaderExtension.encode(this.newChaincodeHeaderExtension()).finish(),
            epoch: 0,
        };
    }

    private newChaincodeHeaderExtension(): protos.IChaincodeHeaderExtension {
        return {
            chaincode_id: {
                name: this.#options.chaincodeId,
            },
        };
    }

    private newChaincodeProposalPayload(): protos.IChaincodeProposalPayload {
        return {
            input: protos.ChaincodeInvocationSpec.encode(this.newChaincodeInvocationSpec()).finish(),
            TransientMap: this.getTransientData(),
        };
    }

    private newChaincodeInvocationSpec(): protos.IChaincodeInvocationSpec {
        return {
            chaincode_spec: {
                type: protos.ChaincodeSpec.Type.NODE,
                chaincode_id: {
                    name: this.#options.chaincodeId,
                },
                input: {
                    args: this.getArgsAsBytes(),
                },
            },
        };
    }

    private getArgsAsBytes(): Uint8Array[] {
        return Array.of(this.#options.transactionName, ...(this.#options.options.arguments ?? []))
            .map(asBytes);
    }

    private getTransientData(): Record<string, Uint8Array> {
        const result: Record<string, Uint8Array> = {};

        for (const [key, value] of Object.entries(this.#options.options.transientData || {})) {
            result[key] = asBytes(value);
        }

        return result;
    }

}

function asBytes(value: string | Uint8Array): Uint8Array {
    return typeof value === 'string' ? Buffer.from(value) : value;
}
