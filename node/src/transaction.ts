/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { Contract } from './contract';
import * as crypto from 'crypto';
import { Signer } from './signer';
import { protos, common, google } from './protos/protos'
import * as grpc from '@grpc/grpc-js';

type Map = { [k: string]: Uint8Array; };

export class Transaction {
    private readonly name: string;
    private readonly contract: Contract;
    private transientMap: Map;

    constructor(name: string, contract: Contract) {
        this.name = name;
        this.contract = contract;
        this.transientMap = {};
    }

    getName() {
        return this.name;
    }

    setTransient(transientMap: Map) {
        this.transientMap = transientMap;
        return this;
    }

    async evaluate(...args: string[]) {
        const gw = this.contract._network._gateway;
        const proposal = this.createProposal(args, gw._signer);
        const signedProposal = this.signProposal(proposal, gw._signer);
        const wrapper = this.createProposedWrapper(signedProposal);
        return this._evaluate(wrapper);
    }

    async submit(...args: string[]) {
        const gw = this.contract._network._gateway;
        const proposal = this.createProposal(args, gw._signer);
        const signedProposal = this.signProposal(proposal, gw._signer);
        const wrapper = this.createProposedWrapper(signedProposal);
        const preparedTxn = await this._endorse(wrapper);
        preparedTxn.envelope!.signature = gw._signer.sign(preparedTxn.envelope!.payload ?? Buffer.from(""));
        await this._submit(preparedTxn);
        return preparedTxn.response?.value?.toString() ?? "";
    }

    private createProposedWrapper(signedProposal: protos.ISignedProposal): protos.IProposedTransaction {
        return {
            proposal: signedProposal
        };
    }

    private createProposal(args: string[], signer: Signer): protos.IProposal {
        const creator = signer.serialize();
        const nonce = crypto.randomBytes(24);
        const hash = crypto.createHash('sha256');
        hash.update(nonce);
        hash.update(creator);
        const txid = hash.digest('hex');

        const hdr = {
            channel_header: common.ChannelHeader.encode({
                type: common.HeaderType.ENDORSER_TRANSACTION,
                tx_id: txid,
                timestamp: google.protobuf.Timestamp.create({
                    seconds: Date.now() / 1000
                }),
                channel_id: this.contract._network.getName(),
                extension: protos.ChaincodeHeaderExtension.encode({
                    chaincode_id: protos.ChaincodeID.create({
                        name: this.contract.getName()
                    })
                }).finish(),
                epoch: 0
            }).finish(),
            signature_header: common.SignatureHeader.encode({
                creator: signer.serialize(),
                nonce: nonce
            }).finish()
        }

        const allArgs = [Buffer.from(this.getName())];
        args.forEach(arg => allArgs.push(Buffer.from(arg)));

        const ccis = protos.ChaincodeInvocationSpec.encode({
            chaincode_spec: protos.ChaincodeSpec.create({
                type: 2,
                chaincode_id: protos.ChaincodeID.create({
                    name: this.contract.getName()
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
                TransientMap: this.transientMap
            }).finish()
        }

        return proposal;
    }

    private signProposal(proposal: protos.IProposal, signer: Signer) {
        const payload = protos.Proposal.encode(proposal).finish();
        const signature = signer.sign(payload);
        return {
            proposal_bytes: payload,
            signature: signature
        };
    }

    private rpcSimple(method: any, requestData: any, callback: any) {
        this.contract._network._gateway._client.makeUnaryRequest(
            'protos.Gateway/' + method.name,
            arg => arg,
            arg => arg,
            requestData,
            new grpc.Metadata(),
            {},
            callback
        )
    }

    private rpcStream(method: any, requestData: any, callback: any) {
        let data: any = undefined;
        const call = this.contract._network._gateway._client.makeServerStreamRequest(
            'protos.Gateway/' + method.name,
            arg => arg,
            arg => arg,
            requestData,
            new grpc.Metadata(),
            {}
        );
        call.on('data', function (event: any) {
            data = event;
        });
        call.on('end', function () {
            callback(null, data);
        });
        call.on('error', function (e: any) {
            // An error has occurred and the stream has been closed.
            callback(e, null);
        });
        call.on('status', function (status: any) {
            // process status
        });
    }

    private async _evaluate(signedProposal: protos.IProposedTransaction): Promise<string> {
        const service = protos.Gateway.create(this.rpcSimple.bind(this), false, false);
        return new Promise((resolve, reject) => {
            service.evaluate(signedProposal, function (err: Error|null, result: protos.Result|undefined) {
                if (err) reject(err);
                resolve(result?.value?.toString());
            });
        })
    }

    private async _endorse(signedProposal: protos.IProposedTransaction): Promise<protos.PreparedTransaction> {
        const service = protos.Gateway.create(this.rpcSimple.bind(this), false, false);
        return new Promise((resolve, reject) => {
            service.endorse(signedProposal, function (err: Error|null, result: protos.PreparedTransaction|undefined) {
                if (err) reject(err);
                resolve(result);
            });
        })
    }

    private async _submit(preparedTransaction: protos.PreparedTransaction): Promise<protos.Event> {
        const service = protos.Gateway.create(this.rpcStream.bind(this), false, false);
        return new Promise((resolve, reject) => {
            service.submit(preparedTransaction, function (err: Error|null, result: protos.Event|undefined) {
                if (err) reject(err);
                resolve(result);
            });
        })
    }


}

