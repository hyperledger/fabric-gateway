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

export interface ProposalOptions {
    arguments?: Array<string|Uint8Array>;
    transientData?: { [k: string]: Uint8Array };
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
            TransientMap: this.#options.options.transientData,
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
            .map(arg => typeof arg === 'string' ? Buffer.from(arg) : arg);
    }
}
