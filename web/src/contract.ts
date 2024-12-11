/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { Proposal } from './proposal';
import { ProposalBuilder, ProposalOptions } from './proposalbuilder';
import { SigningIdentity } from './signingidentity';

/**
 * Represents a smart contract, and allows applications to create transaction proposals using {@link newProposal}. The
 * proposal can be serialized and sent to a remote server that will interact with the Fabric network on behalf of the
 * client application.
 *
 * @example
 * ```typescript
 * const proposal = await contract.newProposal('transactionName', {
 *     arguments: ['one', 'two'],
 *     // Specify additional proposal options here
 * });
 * const serializedProposal = proposal.getBytes();
 * ```
 */
export interface Contract {
    /**
     * Get the name of the chaincode that contains this smart contract.
     * @returns The chaincode name.
     */
    getChaincodeName(): string;

    /**
     * Get the name of the smart contract within the chaincode.
     * @returns The contract name, or `undefined` for the default smart contract.
     */
    getContractName(): string | undefined;

    /**
     * Create a transaction proposal that can be evaluated or endorsed. Supports off-line signing flow.
     * @param transactionName - Name of the transaction to invoke.
     * @param options - Transaction invocation options.
     */
    newProposal(transactionName: string, options?: ProposalOptions): Promise<Proposal>;
}

export interface ContractOptions {
    signingIdentity: SigningIdentity;
    channelName: string;
    chaincodeName: string;
    contractName?: string;
}

export class ContractImpl implements Contract {
    readonly #signingIdentity: SigningIdentity;
    readonly #channelName: string;
    readonly #chaincodeName: string;
    readonly #contractName?: string;

    constructor(options: Readonly<ContractOptions>) {
        this.#signingIdentity = options.signingIdentity;
        this.#channelName = options.channelName;
        this.#chaincodeName = options.chaincodeName;
        this.#contractName = options.contractName;
    }

    getChaincodeName(): string {
        return this.#chaincodeName;
    }

    getContractName(): string | undefined {
        return this.#contractName;
    }

    async newProposal(transactionName: string, options: Readonly<ProposalOptions> = {}): Promise<Proposal> {
        const builder = await ProposalBuilder.newInstance(
            Object.assign({}, options, {
                signingIdentity: this.#signingIdentity,
                channelName: this.#channelName,
                chaincodeName: this.#chaincodeName,
                transactionName: this.#getQualifiedTransactionName(transactionName),
            }),
        );
        return builder.build();
    }

    #getQualifiedTransactionName(transactionName: string): string {
        return this.#contractName ? `${this.#contractName}:${transactionName}` : transactionName;
    }
}
