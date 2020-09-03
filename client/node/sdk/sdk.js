/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

'use strict';

const PROTO_PATH = [
    __dirname + '/../../../protos/gateway.proto',
    __dirname + '/../../../../fabric-protos/peer/proposal.proto',
    __dirname + '/../../../../fabric-protos/peer/proposal_response.proto',
    __dirname + '/../../../../fabric-protos/peer/chaincode.proto',
    __dirname + '/../../../../fabric-protos/common/common.proto',
    __dirname + '/../../../../fabric-protos/common/policies.proto',
    __dirname + '/../../../../fabric-protos/msp/identities.proto',
    __dirname + '/../../../../fabric-protos/msp/msp_principal.proto',
];
const grpc = require('@grpc/grpc-js');
const protoLoader = require('@grpc/proto-loader');
const protobuf = require('protobufjs');
const crypto = require('crypto');
const elliptic = require('elliptic');
const EC = elliptic.ec;
const ecdsaCurve = elliptic.curves['p256'];
const ecdsa = new EC(ecdsaCurve);
const { KEYUTIL } = require('jsrsasign');

const packageDefinition = protoLoader.loadSync(
    PROTO_PATH,
    {
        keepCase: true,
        longs: String,
        enums: String,
        defaults: true,
        oneofs: true
    });

const protoDescriptor = grpc.loadPackageDefinition(packageDefinition);
//console.log(protoDescriptor);
// The protoDescriptor object has the full package hierarchy
const protos = protoDescriptor.protos;

const root = protobuf.loadSync(PROTO_PATH)
//console.log(root)

class Gateway {
    constructor() { }

    async connect(url, signer) {
        this.url = url;
        this.signer = signer;
        this.stub = new protos.Gateway(url, grpc.credentials.createInsecure());
        this.evaluate = signedProposal => {
            return new Promise((resolve, reject) => {
                this.stub.evaluate(signedProposal, function (err, result) {
                    if (err) reject(err);
                    resolve(result.value.toString());
                });
            })
        };
        this.prepare = signedProposal => {
            return new Promise((resolve, reject) => {
                this.stub.prepare(signedProposal, function (err, result) {
                    if (err) reject(err);
                    resolve(result);
                });
            })
        };
        this.commit = preparedTransaction => {
            return new Promise((resolve, reject) => {
                const call = this.stub.commit(preparedTransaction);
                call.on('data', function (event) {
                    console.log('Event received: ', event.value.toString());
                });
                call.on('end', function () {
                    resolve()
                });
                call.on('error', function (e) {
                    // An error has occurred and the stream has been closed.
                    reject(e);
                });
                call.on('status', function (status) {
                    // process status
                });
            })
        };
    }

    getNetwork(networkName) {
        return new Network(networkName, this);
    }
}

class Network {
    constructor(name, gateway) {
        this.name = name;
        this.gateway = gateway;
    }

    getContract(contractName) {
        return new Contract(contractName, this);
    }
}

class Contract {
    constructor(name, network) {
        this.name = name;
        this.network = network;
    }

    createTransaction(transactionName) {
        return new Transaction(transactionName, this);
    }

    async evaluateTransaction(name, ...args) {
        return this.createTransaction(name).evaluate(...args);
    }

    async submitTransaction(name, ...args) {
        return this.createTransaction(name).submit(...args);
    }

    prepareToEvaluate(transactionName) {
        return new EvaluateTransaction(transactionName, this);
    }

    prepareToSubmit(transactionName) {
        return new SubmitTransaction(transactionName, this);
    }
}

class Transaction {
    constructor(name, contract) {
        this.name = name;
        this.contract = contract;
    }

    setTransient(transientMap) {
        this.transientMap = transientMap;
    }

    async evaluate(...args) {
        const gw = this.contract.network.gateway;
        const proposal = createProposal(this, args, gw.signer);
        const signedProposal = signProposal(proposal, gw.signer);
        return gw.evaluate(signedProposal);
    }

    async submit(...args) {
        const gw = this.contract.network.gateway;
        const proposal = createProposal(this, args, gw.signer);
        const signedProposal = signProposal(proposal, gw.signer);
        const preparedTxn = await gw.prepare(signedProposal);
        preparedTxn.envelope.signature = gw.signer.sign(preparedTxn.envelope.payload);
        await gw.commit(preparedTxn);
        return preparedTxn.response.value.toString();
    }
}

class EvaluateTransaction extends Transaction {
    setArgs(...args) {
        this.args = args;
    }

    async invoke() {
        return this.evaluate(...this.args);
    }
}

class SubmitTransaction extends Transaction {
    setArgs(...args) {
        this.args = args;
    }

    async invoke() {
        return this.submit(...this.args);
    }
}

class Signer {
    constructor(mspid, certPem, keyPem) {
        this.mspid = mspid;
        this.cert = certPem;
        const { prvKeyHex } = KEYUTIL.getKey(keyPem.toString()); // convert the pem encoded key to hex encoded private key
        this.signKey = ecdsa.keyFromPrivate(prvKeyHex, 'hex');
        this.serialized = marshal('msp.SerializedIdentity', {
            mspid: mspid,
            idBytes: certPem
        })
    }

    sign(msg) {
        const hash = crypto.createHash('sha256');
        hash.update(msg);
        const digest = hash.digest();
        const sig = ecdsa.sign(Buffer.from(digest, 'hex'), this.signKey);
        _preventMalleability(sig, ecdsaCurve);
        return sig.toDER();
    }

    serialize() {
        return this.serialized;
    }
}

function _preventMalleability(sig, curve) {

    const halfOrder = curve.n.shrn(1);
    if (!halfOrder) {
        throw new Error('Can not find the half order needed to calculate "s" value for immalleable signatures. Unsupported curve name: ' + curve);
    }

    if (sig.s.cmp(halfOrder) === 1) {
        const bigNum = curve.n;
        sig.s = bigNum.sub(sig.s);
    }

    return sig;
}

function createProposal(txn, args, signer) {
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
            channelId: txn.contract.network.name,
            extension: marshal('protos.ChaincodeHeaderExtension', {
                chaincodeId: create('protos.ChaincodeID', {
                    name: txn.contract.name
                })
            }),
            epoch: 0
        }),
        signatureHeader: marshal('common.SignatureHeader', {
            creator: signer.serialize(),
            nonce: nonce
        })
    }

    const allArgs = [Buffer.from(txn.name)];
    args.forEach(arg => allArgs.push(Buffer.from(arg)));

    const ccis = marshal('protos.ChaincodeInvocationSpec', {
        chaincodeSpec: create('protos.ChaincodeSpec', {
            type: 2,
            chaincodeId: create('protos.ChaincodeID', {
                name: txn.contract.name
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
            TransientMap: txn.transientMap
        })
    }

    return proposal;
}

function signProposal(proposal, signer) {
    const payload = marshal('protos.Proposal', proposal);
    const signature = signer.sign(payload);
    return {
        proposal_bytes: payload,
        signature: signature
    };
}

function create(name, payload) {
    const type = root.lookupType(name);
    const errMsg = type.verify(payload);
    if (errMsg) console.log('ERROR: ', errMsg);
    const message = type.create(payload);
    //console.log('message:', message);
    return message;
}

function marshal(name, payload) {
    const type = root.lookupType(name);
    const errMsg = type.verify(payload);
    if (errMsg) console.log('ERROR: ', errMsg);
    const message = type.create(payload);
    //console.log('message:', message);
    const buffer = type.encode(message).finish();
    //console.log('buffer: ', buffer)
    return buffer;
}

module.exports = { Gateway, Network, Contract, Transaction, Signer }