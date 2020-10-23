/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

const protoBase = __dirname + '/../../node_modules/fabric-protos/protos';
console.log(protoBase);
const PROTO_PATH = [
    __dirname + '/../../../protos/gateway.proto', // TODO this is highly fragile - once the gateway sdk is published, it can be added as a dependency and pulled in like the other protos
    protoBase + '/peer/proposal.proto',
    protoBase + '/peer/proposal_response.proto',
    protoBase + '/peer/chaincode.proto',
    protoBase + '/common/common.proto',
    protoBase + '/common/policies.proto',
    protoBase + '/msp/identities.proto',
    protoBase + '/msp/msp_principal.proto',
];
import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import * as protobuf from 'protobufjs';

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
const protos: any = protoDescriptor.protos;
const protosGateway = protos.Gateway;

const root = protobuf.loadSync(PROTO_PATH)
//console.log(root)


function create(name: string, payload: {[k: string]: any;}) {
    const type = root.lookupType(name);
    const errMsg = type.verify(payload);
    if (errMsg) console.log('ERROR: ', errMsg);
    const message = type.create(payload);
    //console.log('message:', message);
    return message;
}

function marshal(name: string, payload: {[k: string]: any;}): Uint8Array {
    const type = root.lookupType(name);
    const errMsg = type.verify(payload);
    if (errMsg) console.log('ERROR: ', errMsg);
    const message = type.create(payload);
    //console.log('message:', message);
    const buffer = type.encode(message).finish();
    //console.log('buffer: ', buffer)
    return buffer;
}

export {create, marshal, protosGateway}