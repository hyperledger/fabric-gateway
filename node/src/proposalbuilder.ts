/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Timestamp } from 'google-protobuf/google/protobuf/timestamp_pb';
import { GatewayClient } from './client';
import { Proposal, ProposalImpl } from './proposal';
import { ChannelHeader, Header, HeaderType } from './protos/common/common_pb';
import { ProposedTransaction } from './protos/gateway/gateway_pb';
import { ChaincodeID, ChaincodeInput, ChaincodeInvocationSpec, ChaincodeSpec } from './protos/peer/chaincode_pb';
import { ChaincodeHeaderExtension, ChaincodeProposalPayload, Proposal as ProposalProto, SignedProposal } from './protos/peer/proposal_pb';
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

export interface ProposalBuilderOptions extends Readonly<ProposalOptions> {
    readonly client: GatewayClient;
    readonly signingIdentity: SigningIdentity;
    readonly channelName: string;
    readonly chaincodeId: string;
    readonly transactionName: string;
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
            proposedTransaction: this.newProposedTransaction(),
        });
    }

    private newProposedTransaction(): ProposedTransaction {
        const result = new ProposedTransaction();
        result.setProposal(this.newSignedProposal());
        result.setTransactionId(this.#transactionContext.getTransactionId());
        if (this.#options.endorsingOrganizations) {
            result.setEndorsingOrganizationsList(this.#options.endorsingOrganizations);
        }
        return result;
    }

    private newSignedProposal(): SignedProposal {
        const result = new SignedProposal();
        result.setProposalBytes(this.newProposal().serializeBinary());
        return result;
    }

    private newProposal(): ProposalProto {
        const result = new ProposalProto();
        result.setHeader(this.newHeader().serializeBinary());
        result.setPayload(this.newChaincodeProposalPayload().serializeBinary());
        return result;
    }

    private newHeader(): Header {
        const result = new Header();
        result.setChannelHeader(this.newChannelHeader().serializeBinary());
        result.setSignatureHeader(this.#transactionContext.getSignatureHeader().serializeBinary());
        return result;
    }

    private newChannelHeader(): ChannelHeader {
        const result = new ChannelHeader();
        result.setType(HeaderType.ENDORSER_TRANSACTION);
        result.setTxId(this.#transactionContext.getTransactionId());
        result.setTimestamp(Timestamp.fromDate(new Date()));
        result.setChannelId(this.#options.channelName);
        result.setExtension$(this.newChaincodeHeaderExtension().serializeBinary());
        result.setEpoch(0);
        return result;
    }

    private newChaincodeHeaderExtension(): ChaincodeHeaderExtension {
        const result = new ChaincodeHeaderExtension();
        result.setChaincodeId(this.newChaincodeID());
        return result;
    }

    private newChaincodeID(): ChaincodeID {
        const result = new ChaincodeID();
        result.setName(this.#options.chaincodeId);
        return result;
    }

    private newChaincodeProposalPayload(): ChaincodeProposalPayload {
        const result = new ChaincodeProposalPayload();
        result.setInput(this.newChaincodeInvocationSpec().serializeBinary());
        const transientMap = result.getTransientmapMap();
        for (const [key, value] of Object.entries(this.getTransientData())) {
            transientMap.set(key, value);
        }
        return result;
    }

    private newChaincodeInvocationSpec(): ChaincodeInvocationSpec {
        const result = new ChaincodeInvocationSpec();
        result.setChaincodeSpec(this.newChaincodeSpec());
        return result;
    }

    private newChaincodeSpec(): ChaincodeSpec {
        const result = new ChaincodeSpec();
        result.setType(ChaincodeSpec.Type.NODE);
        result.setChaincodeId(this.newChaincodeID());
        result.setInput(this.newChaincodeInput());
        return result;
    }

    private newChaincodeInput(): ChaincodeInput {
        const result = new ChaincodeInput();
        result.setArgsList(this.getArgsAsBytes());
        return result;
    }

    private getArgsAsBytes(): Uint8Array[] {
        return Array.of(this.#options.transactionName, ...(this.#options.arguments ?? []))
            .map(asBytes);
    }

    private getTransientData(): Record<string, Uint8Array> {
        const result: Record<string, Uint8Array> = {};

        for (const [key, value] of Object.entries(this.#options.transientData || {})) {
            result[key] = asBytes(value);
        }

        return result;
    }

}

function asBytes(value: string | Uint8Array): Uint8Array {
    return typeof value === 'string' ? Buffer.from(value) : value;
}
