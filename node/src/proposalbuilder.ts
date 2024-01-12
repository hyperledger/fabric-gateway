/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { common, gateway, peer } from '@hyperledger/fabric-protos';
import { Timestamp } from 'google-protobuf/google/protobuf/timestamp_pb';
import { GatewayClient } from './client';
import { Proposal, ProposalImpl } from './proposal';
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

export interface ProposalBuilderOptions extends ProposalOptions {
    client: GatewayClient;
    signingIdentity: SigningIdentity;
    channelName: string;
    chaincodeName: string;
    transactionName: string;
}

export class ProposalBuilder {
    readonly #options: Readonly<ProposalBuilderOptions>;
    readonly #transactionContext: TransactionContext;

    constructor(options: Readonly<ProposalBuilderOptions>) {
        this.#options = options;
        this.#transactionContext = new TransactionContext(options.signingIdentity);
    }

    build(): Proposal {
        return new ProposalImpl({
            client: this.#options.client,
            signingIdentity: this.#options.signingIdentity,
            channelName: this.#options.channelName,
            proposedTransaction: this.#newProposedTransaction(),
        });
    }

    #newProposedTransaction(): gateway.ProposedTransaction {
        const result = new gateway.ProposedTransaction();
        result.setProposal(this.#newSignedProposal());
        result.setTransactionId(this.#transactionContext.getTransactionId());
        if (this.#options.endorsingOrganizations) {
            result.setEndorsingOrganizationsList(this.#options.endorsingOrganizations);
        }
        return result;
    }

    #newSignedProposal(): peer.SignedProposal {
        const result = new peer.SignedProposal();
        result.setProposalBytes(this.#newProposal().serializeBinary());
        return result;
    }

    #newProposal(): peer.Proposal {
        const result = new peer.Proposal();
        result.setHeader(this.#newHeader().serializeBinary());
        result.setPayload(this.#newChaincodeProposalPayload().serializeBinary());
        return result;
    }

    #newHeader(): common.Header {
        const result = new common.Header();
        result.setChannelHeader(this.#newChannelHeader().serializeBinary());
        result.setSignatureHeader(this.#transactionContext.getSignatureHeader().serializeBinary());
        return result;
    }

    #newChannelHeader(): common.ChannelHeader {
        const result = new common.ChannelHeader();
        result.setType(common.HeaderType.ENDORSER_TRANSACTION);
        result.setTxId(this.#transactionContext.getTransactionId());
        result.setTimestamp(Timestamp.fromDate(new Date()));
        result.setChannelId(this.#options.channelName);
        result.setExtension$(this.#newChaincodeHeaderExtension().serializeBinary());
        result.setEpoch(0);
        return result;
    }

    #newChaincodeHeaderExtension(): peer.ChaincodeHeaderExtension {
        const result = new peer.ChaincodeHeaderExtension();
        result.setChaincodeId(this.#newChaincodeID());
        return result;
    }

    #newChaincodeID(): peer.ChaincodeID {
        const result = new peer.ChaincodeID();
        result.setName(this.#options.chaincodeName);
        return result;
    }

    #newChaincodeProposalPayload(): peer.ChaincodeProposalPayload {
        const result = new peer.ChaincodeProposalPayload();
        result.setInput(this.#newChaincodeInvocationSpec().serializeBinary());
        const transientMap = result.getTransientmapMap();
        for (const [key, value] of Object.entries(this.#getTransientData())) {
            transientMap.set(key, value);
        }
        return result;
    }

    #newChaincodeInvocationSpec(): peer.ChaincodeInvocationSpec {
        const result = new peer.ChaincodeInvocationSpec();
        result.setChaincodeSpec(this.#newChaincodeSpec());
        return result;
    }

    #newChaincodeSpec(): peer.ChaincodeSpec {
        const result = new peer.ChaincodeSpec();
        result.setType(peer.ChaincodeSpec.Type.NODE);
        result.setChaincodeId(this.#newChaincodeID());
        result.setInput(this.#newChaincodeInput());
        return result;
    }

    #newChaincodeInput(): peer.ChaincodeInput {
        const result = new peer.ChaincodeInput();
        result.setArgsList(this.#getArgsAsBytes());
        return result;
    }

    #getArgsAsBytes(): Uint8Array[] {
        return Array.of(this.#options.transactionName, ...(this.#options.arguments ?? [])).map(asBytes);
    }

    #getTransientData(): Record<string, Uint8Array> {
        const result: Record<string, Uint8Array> = {};

        for (const [key, value] of Object.entries(this.#options.transientData ?? {})) {
            result[key] = asBytes(value);
        }

        return result;
    }
}

function asBytes(value: string | Uint8Array): Uint8Array {
    return typeof value === 'string' ? Buffer.from(value) : value;
}
