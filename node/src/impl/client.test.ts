/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import * as grpc from '@grpc/grpc-js';
import { ClientImpl } from './client';
import { protos } from '../protos/protos'

function createMockProposal(): protos.IProposedTransaction {
    return {
        proposal: {
            proposal_bytes: Buffer.from('mock payload'),
            signature: Buffer.from('mock signature')
        }
    }
}

function createMockTransaction(): protos.IPreparedTransaction {
    return {
        envelope: {
            payload: Buffer.from('mock payload'),
            signature: Buffer.from('mock signature')
        }
    }
}

function createMockServiceClient(err: Error | null, response: Uint8Array | null) {
    return {
        makeUnaryRequest: (
            method: string,
            serialize: (value: any) => Buffer,
            deserialize: (value: Buffer) => Buffer,
            argument: any, metadata: grpc.Metadata,
            options: grpc.CallOptions,
            callback: any) => {
            callback(err, response);
        },
        makeServerStreamRequest: (method: string,
            serialize: (value: any) => Buffer,
            deserialize: (value: Buffer) => Buffer,
            argument: any,
            metadata: grpc.Metadata,
            options?: grpc.CallOptions | undefined): any => {
                return {
                    on: (action: string, handler: Function) => {
                        handler(action === 'error' ? err : response);
                    }
                }
        }
    }
}

test('evaluate', async () => {
    const client = new ClientImpl('mockUrl:1234');
    const mockValue = 'mock result';
    const mockResponse = protos.Result.encode({ value: Buffer.from(mockValue) }).finish();
    (client as any).serviceClient = createMockServiceClient(null, mockResponse);
    const proposal = createMockProposal();
    const result = await client._evaluate(proposal);
    expect(result).toEqual(mockValue);
})

test('evaluate throws error', async () => {
    const client = new ClientImpl('mockUrl:1234');
    (client as any).serviceClient = createMockServiceClient(new Error('mock error'), null);
    const proposal = createMockProposal();
    await expect(client._evaluate(proposal)).rejects.toThrow('mock error');
})

test('endorse', async () => {
    const client = new ClientImpl('mockUrl:1234');
    const mockValue = {
        txId: 'mock txid'
    };
    const mockResponse = protos.PreparedTransaction.encode(mockValue).finish();
    (client as any).serviceClient = createMockServiceClient(null, mockResponse);
    const proposal = createMockProposal();
    const result = await client._endorse(proposal);
    expect(result).toEqual(mockValue);
})

test('endorse throws error', async () => {
    const client = new ClientImpl('mockUrl:1234');
    (client as any).serviceClient = createMockServiceClient(new Error('mock error'), null);
    const proposal = createMockProposal();
    await expect(client._endorse(proposal)).rejects.toThrow('mock error');
})

test('submit', async () => {
    const client = new ClientImpl('mockUrl:1234');
    const mockValue = {
        value: Buffer.from('mock value')
    };
    const mockResponse = protos.Event.encode(mockValue).finish();
    (client as any).serviceClient = createMockServiceClient(null, mockResponse);
    const txn = createMockTransaction();
    const result = await client._submit(txn);
    expect(result).toEqual(mockValue);
})

test('submit throws error', async () => {
    const client = new ClientImpl('mockUrl:1234');
    (client as any).serviceClient = createMockServiceClient(new Error('mock error'), null);
    const txn = createMockTransaction();
    await expect(client._submit(txn)).rejects.toThrow('mock error');
})

