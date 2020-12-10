/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { GatewayClient } from 'client';
import * as crypto from 'crypto';
import { SigningIdentity } from 'signingidentity';
import { common, google, protos } from './protos/protos';

type TransientData = { [k: string]: Uint8Array; };

type Commit = Promise<Uint8Array>;

export interface Transaction {
    /**
     * Get the serialized transaction message.
     */
    getBytes(): Uint8Array;

    /**
     * Get the digest of the transaction. This is used to generate a digital signature.
     */
    getDigest(): Uint8Array;

    /**
     * Get the transaction result. This is obtained during the endorsement process when the transaction proposal is
     * run on endorsing peers.
     */
    getResult(): Uint8Array;

    /**
     * Submit the transaction to the orderer to be committed to the ledger.
     */
    submit(): Promise<Commit>;
}

export class OldTransaction {
    readonly #client: GatewayClient;
    readonly #signingIdentity: SigningIdentity;
    readonly #channelName: string;
    readonly #chaincodeId: string;
    readonly #transactionName: string;
    #transientData: TransientData;

    constructor(client: GatewayClient, signingIdentity: SigningIdentity, channelName: string, chaincodeId: string, transactionName: string) {
        this.#client = client;
        this.#signingIdentity = signingIdentity;
        this.#channelName = channelName;
        this.#chaincodeId = chaincodeId;
        this.#transactionName = transactionName;
        this.#transientData = {};
    }

    getName(): string {
        return this.#transactionName;
    }

    setTransient(transientMap: TransientData): this {
        this.#transientData = transientMap;
        return this;
    }

    async evaluate(...args: string[]): Promise<Uint8Array> {
        const proposal = this.createProposal(args);
        const signedProposal = this.signProposal(proposal);
        const wrapper = this.createProposedWrapper(signedProposal);
        const result = await this.#client.evaluate(wrapper);
        return result.value!; // TODO: Can this be null or undefined?
    }

    async submit(...args: string[]): Promise<Uint8Array> {
        const proposal = this.createProposal(args);
        const signedProposal = this.signProposal(proposal);
        const wrapper = this.createProposedWrapper(signedProposal);
        const preparedTxn = await this.#client.endorse(wrapper);
        const envelope = preparedTxn.envelope!;
        const digest = this.#signingIdentity.hash(envelope.payload!);
        envelope.signature = this.#signingIdentity.sign(digest);
        await this.#client.submit(preparedTxn);
        return preparedTxn.response!.value!; // TODO: Can this be null or undefined?
    }

    private createProposedWrapper(signedProposal: protos.ISignedProposal): protos.IProposedTransaction {
        return {
            proposal: signedProposal
        };
    }

    private createProposal(args: string[]): protos.IProposal {
        const creator = this.#signingIdentity.getCreator();
        const nonce = crypto.randomBytes(24);
        const txHash = this.#signingIdentity.hash(Buffer.concat([nonce, creator]));
        const txid = Buffer.from(txHash).toString('hex');

        const hdr = {
            channel_header: common.ChannelHeader.encode({
                type: common.HeaderType.ENDORSER_TRANSACTION,
                tx_id: txid,
                timestamp: google.protobuf.Timestamp.create({
                    seconds: Date.now() / 1000
                }),
                channel_id: this.#channelName,
                extension: protos.ChaincodeHeaderExtension.encode({
                    chaincode_id: protos.ChaincodeID.create({
                        name: this.#chaincodeId
                    })
                }).finish(),
                epoch: 0
            }).finish(),
            signature_header: common.SignatureHeader.encode({
                creator,
                nonce: nonce
            }).finish()
        }

        const allArgs = [Buffer.from(this.getName())];
        args.forEach(arg => allArgs.push(Buffer.from(arg)));

        const ccis = protos.ChaincodeInvocationSpec.encode({
            chaincode_spec: protos.ChaincodeSpec.create({
                type: 2,
                chaincode_id: protos.ChaincodeID.create({
                    name: this.#chaincodeId
                }),
                input: protos.ChaincodeInput.create({
                    args: allArgs
                })
            })
        }).finish();

        const proposal = {
            header: common.Header.encode(hdr).finish(),
            payload: protos.ChaincodeProposalPayload.encode({
                input: ccis,
                TransientMap: this.#transientData
            }).finish()
        }

        return proposal;
    }

    private signProposal(proposal: protos.IProposal): protos.ISignedProposal {
        const payload = protos.Proposal.encode(proposal).finish();
        const digest = this.#signingIdentity.hash(payload);
        const signature = this.#signingIdentity.sign(digest);
        return {
            proposal_bytes: payload,
            signature: signature
        };
    }
}
