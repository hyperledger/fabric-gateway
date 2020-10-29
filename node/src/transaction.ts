/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { Contract } from './contract';
import { create, marshal } from './impl/protoutils'
import * as crypto from 'crypto';
import { Signer } from './signer';

export class Transaction {
    private readonly name: string;
    private readonly contract: Contract;
    private transientMap: object;
    private readonly stub: any;

    constructor(name: string, contract: Contract) {
        this.name = name;
        this.contract = contract;
        this.transientMap = {};
        const gw: any = this.contract._network._gateway;
        this.stub = gw._stub;
    }

    getName() {
        return this.name;
    }

    setTransient(transientMap: object) {
        this.transientMap = transientMap;
        return this;
    }

    async evaluate(...args: string[]) {
        const gw: any = this.contract._network._gateway;
        const proposal = this.createProposal(args, gw._signer);
        const signedProposal = this.signProposal(proposal, gw._signer);
        const wrapper = this.createProposedWrapper(signedProposal);
        return this._evaluate(wrapper);
    }

    async submit(...args: string[]) {
        const gw: any = this.contract._network._gateway;
        const proposal = this.createProposal(args, gw._signer);
        const signedProposal = this.signProposal(proposal, gw._signer);
        const wrapper = this.createProposedWrapper(signedProposal);
        const preparedTxn = await this._endorse(wrapper);
        preparedTxn.envelope.signature = gw._signer.sign(preparedTxn.envelope.payload);
        await this._submit(preparedTxn);
        return preparedTxn.response.value.toString();
    }

    private createProposedWrapper(signedProposal: any) {
        return {
            proposal: signedProposal
        };
    }

    private createProposal(args: string[], signer: Signer) {
        const creator = signer.serialize();
        const nonce = crypto.randomBytes(24);
        const hash = crypto.createHash('sha256');
        hash.update(nonce);
        hash.update(creator);
        const txid = hash.digest('hex');
    
        const hdr = {
            channelHeader: marshal('common.ChannelHeader', {
                type: 3, // ENDORSER_TRANSACTION - TODO lookup enum
                txId: txid,
                timestamp: create('google.protobuf.Timestamp', {
                    timestamp: Date.now()
                }),
                channelId: this.contract._network.getName(),
                extension: marshal('protos.ChaincodeHeaderExtension', {
                    chaincodeId: create('protos.ChaincodeID', {
                        name: this.contract.getName()
                    })
                }),
                epoch: 0
            }),
            signatureHeader: marshal('common.SignatureHeader', {
                creator: signer.serialize(),
                nonce: nonce
            })
        }
    
        const allArgs = [Buffer.from(this.getName())];
        args.forEach(arg => allArgs.push(Buffer.from(arg)));
    
        const ccis = marshal('protos.ChaincodeInvocationSpec', {
            chaincodeSpec: create('protos.ChaincodeSpec', {
                type: 2,
                chaincodeId: create('protos.ChaincodeID', {
                    name: this.contract.getName()
                }),
                input: create('protos.ChaincodeInput', {
                    args: allArgs
                })
            })
        })
    
        const proposal = {
            header: marshal('common.Header', hdr),
            payload: marshal('protos.ChaincodeProposalPayload', {
                input: ccis,
                TransientMap: this.transientMap
            })
        }
    
        return proposal;
    }
    
    private signProposal(proposal: any, signer: Signer) {
        const payload = marshal('protos.Proposal', proposal);
        const signature = signer.sign(payload);
        return {
            proposal_bytes: payload,
            signature: signature
        };
    }

    private async _endorse(signedProposal: any): Promise<any> {
        return new Promise((resolve, reject) => {
            this.stub.endorse(signedProposal, function (err: any, result: any) {
                if (err) reject(err);
                resolve(result);
            });
        })
    };

    private async _submit(preparedTransaction: any): Promise<any> {
        return new Promise((resolve, reject) => {
            const call = this.stub.submit(preparedTransaction);
            call.on('data', function (event: any) {
                console.log('Event received: ', event.value.toString());
            });
            call.on('end', function () {
                resolve()
            });
            call.on('error', function (e: any) {
                // An error has occurred and the stream has been closed.
                reject(e);
            });
            call.on('status', function (status: any) {
                // process status
            });
        })
    }

    private async _evaluate(signedProposal: any): Promise<any>  {
        return new Promise((resolve, reject) => {
            this.stub.evaluate(signedProposal, function (err: any, result: any) {
                if (err) reject(err);
                resolve(result.value.toString());
            });
        })
    }
}

