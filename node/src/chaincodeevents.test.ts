/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CallOptions, Metadata, ServiceError, status } from '@grpc/grpc-js';
import { ChaincodeEvent } from './chaincodeevent';
import { Gateway, internalConnect, InternalConnectOptions } from './gateway';
import { GatewayError } from './gatewayerror';
import { Identity } from './identity/identity';
import { Network } from './network';
import { ChaincodeEventsRequest as ChaincodeEventsRequestProto, ChaincodeEventsResponse, SignedChaincodeEventsRequest } from './protos/gateway/gateway_pb';
import { SeekPosition } from './protos/orderer/ab_pb';
import { ChaincodeEvent as ChaincodeEventProto } from './protos/peer/chaincode_event_pb';
import { MockGatewayGrpcClient, newServerStreamResponse, readElements } from './testutils.test';
import * as checkpointers from './checkpointers';

function assertDecodeChaincodeEventsRequest(signedRequest: SignedChaincodeEventsRequest): ChaincodeEventsRequestProto {
    const requestBytes = signedRequest.getRequest_asU8();
    expect(requestBytes).toBeDefined();
    return ChaincodeEventsRequestProto.deserializeBinary(requestBytes);
}

function newChaincodeEvent(blockNumber: number, event: ChaincodeEventProto): ChaincodeEvent {
    return {
        blockNumber: BigInt(blockNumber),
        chaincodeName: event.getChaincodeId() ?? '',
        eventName: event.getEventName() ?? '',
        transactionId: event.getTxId() ?? '',
        payload: event.getPayload_asU8() ?? new Uint8Array(),
    };
}
function assertChaincodeEventRequest(request: ChaincodeEventsRequestProto, channelName: string, chaincodeName: string, typeCase: SeekPosition.TypeCase, blockNumber?: bigint, transactionId?: string): void {
    expect(request.getChannelId()).toBe(channelName);
    expect(request.getChaincodeId()).toBe(chaincodeName);

    const startPosition = request.getStartPosition();

    expect(startPosition).toBeDefined();
    expect(startPosition?.getTypeCase()).toBe(typeCase);
    if (blockNumber != undefined) {
        expect(startPosition?.getSpecified()?.getNumber()).toBe(Number(blockNumber));
    }
    if (transactionId) {
        expect(request.getAfterTransactionId()).toEqual(transactionId);
    }
}

describe('Chaincode Events', () => {
    const channelName = 'CHANNEL_NAME';
    const signature = Buffer.from('SIGNATURE');
    const serviceError: ServiceError = Object.assign(new Error('ERROR_MESSAGE'), {
        code: status.UNAVAILABLE,
        details: 'DETAILS',
        metadata: new Metadata(),
    });

    let chaincodeEventsOptions: () => CallOptions;
    let client: MockGatewayGrpcClient;
    let identity: Identity;
    let gateway: Gateway;
    let network: Network;

    const event1 = new ChaincodeEventProto();
    event1.setChaincodeId('CHAINCODE');
    event1.setTxId('tx1');
    event1.setEventName('event1'),
    event1.setPayload(Buffer.from('payload1'));

    const event2 = new ChaincodeEventProto();
    event2.setChaincodeId('CHAINCODE');
    event2.setTxId('tx2');
    event2.setEventName('event2'),
    event2.setPayload(Buffer.from('payload2'));

    const event3 = new ChaincodeEventProto();
    event3.setChaincodeId('CHAINCODE');
    event3.setTxId('tx3');
    event3.setEventName('event3'),
    event3.setPayload(Buffer.from('payload3'));

    beforeEach(() => {
        const now = new Date();
        const callOptions = {
            deadline: now.setHours(now.getHours() + 1),
        };
        chaincodeEventsOptions = () => callOptions; // Return a specific object to test modification

        client = new MockGatewayGrpcClient();
        identity = {
            mspId: 'MSP_ID',
            credentials: Buffer.from('CERTIFICATE'),
        };

        const options: InternalConnectOptions = {
            identity,
            signer: () => Promise.resolve(signature),
            client,
            chaincodeEventsOptions,
        };
        gateway = internalConnect(options);
        network = gateway.getNetwork(channelName);
    });

    describe('request', () => {
        it('sends valid request with default start position', async () => {
            await network.getChaincodeEvents('CHAINCODE');

            const signedRequest = client.getChaincodeEventsRequests()[0];
            expect(signedRequest.getSignature()).toEqual(signature);

            const request = assertDecodeChaincodeEventsRequest(signedRequest);
            assertChaincodeEventRequest(request, channelName, 'CHAINCODE', SeekPosition.TypeCase.NEXT_COMMIT);
        });

        it('throws with negative specified start block number', async () => {
            const startBlock = BigInt(-1);
            return expect(network.getChaincodeEvents('CHAINCODE', { startBlock }))
                .rejects
                .toThrow();
        });

        it('sends valid request with specified start block number', async () => {
            const startBlock = BigInt(418);

            await network.getChaincodeEvents('CHAINCODE', { startBlock });

            const signedRequest = client.getChaincodeEventsRequests()[0];
            expect(signedRequest.getSignature()).toEqual(signature);

            const request = assertDecodeChaincodeEventsRequest(signedRequest);
            assertChaincodeEventRequest(request, channelName, 'CHAINCODE', SeekPosition.TypeCase.SPECIFIED, startBlock);
        });

        it('Sends valid request with specified start block number and fresh checkpointer', async () => {
            const startBlock = BigInt(418);
            const checkpointer = checkpointers.inMemory();

            await network.getChaincodeEvents('CHAINCODE', { startBlock: startBlock, checkpointer});

            const signedRequest = client.getChaincodeEventsRequests()[0];
            expect(signedRequest.getSignature()).toEqual(signature);

            const request = assertDecodeChaincodeEventsRequest(signedRequest);
            assertChaincodeEventRequest(request, channelName, 'CHAINCODE', SeekPosition.TypeCase.SPECIFIED, startBlock, '');
        });

        it('Sends valid request with specified start block and checkpointed block', async () => {
            const startBlock = BigInt(418);
            const checkpointer = checkpointers.inMemory();
            await checkpointer.checkpointBlock(1n);

            await network.getChaincodeEvents('CHAINCODE', { startBlock: startBlock, checkpointer});

            const signedRequest = client.getChaincodeEventsRequests()[0];
            expect(signedRequest.getSignature()).toEqual(signature);

            const request = assertDecodeChaincodeEventsRequest(signedRequest);
            assertChaincodeEventRequest(request, channelName, 'CHAINCODE', SeekPosition.TypeCase.SPECIFIED, 1n+ 1n, '');
        });

        it('Sends valid request with specified start block and checkpointed transaction id', async () => {
            const startBlock = BigInt(418);
            const checkpointer = checkpointers.inMemory();
            await checkpointer.checkpointTransaction(1n, 'txn1');

            await network.getChaincodeEvents('CHAINCODE', { startBlock: startBlock, checkpointer});

            const signedRequest = client.getChaincodeEventsRequests()[0];
            expect(signedRequest.getSignature()).toEqual(signature);

            const request = assertDecodeChaincodeEventsRequest(signedRequest);
            assertChaincodeEventRequest(request, channelName, 'CHAINCODE', SeekPosition.TypeCase.SPECIFIED, 1n, 'txn1');
        });

        it('Sends valid request with no start block and fresh checkpointer', async () => {
            const checkpointer = checkpointers.inMemory();

            await network.getChaincodeEvents('CHAINCODE', { checkpointer});

            const signedRequest = client.getChaincodeEventsRequests()[0];
            expect(signedRequest.getSignature()).toEqual(signature);

            const request = assertDecodeChaincodeEventsRequest(signedRequest);
            assertChaincodeEventRequest(request, channelName, 'CHAINCODE', SeekPosition.TypeCase.NEXT_COMMIT);
        });

        it('Sends valid request with no start block and checkpointer transaction id', async () => {
            const checkpointer = checkpointers.inMemory();
            await checkpointer.checkpointTransaction(1n, 'txn1');

            await network.getChaincodeEvents('CHAINCODE', { checkpointer});

            const signedRequest = client.getChaincodeEventsRequests()[0];
            expect(signedRequest.getSignature()).toEqual(signature);

            const request = assertDecodeChaincodeEventsRequest(signedRequest);
            assertChaincodeEventRequest(request, channelName, 'CHAINCODE', SeekPosition.TypeCase.SPECIFIED, 1n, 'txn1');
        });

        it('Sends valid request with with start block and checkpointer chaincode event', async () => {
            const checkpointer = checkpointers.inMemory();
            const event: ChaincodeEvent =  {
                blockNumber: BigInt(1),
                chaincodeName: 'chaincode',
                eventName: 'event1',
                transactionId: 'txn1',
                payload: new Uint8Array(),
            };

            await checkpointer.checkpointChaincodeEvent(event);
            await network.getChaincodeEvents('CHAINCODE', { checkpointer });

            const signedRequest = client.getChaincodeEventsRequests()[0];
            expect(signedRequest.getSignature()).toEqual(signature);

            const request = assertDecodeChaincodeEventsRequest(signedRequest);
            assertChaincodeEventRequest(request, channelName, 'CHAINCODE', SeekPosition.TypeCase.SPECIFIED, 1n, 'txn1');
        });

        it('uses specified call options', async () => {
            const deadline = Date.now() + 1000;

            await network.newChaincodeEventsRequest('CHAINCODE')
                .getEvents({ deadline });

            const actual = client.getChaincodeEventsOptions()[0];
            expect(actual.deadline).toBe(deadline);
        });

        it('uses default call options', async () => {
            await network.getChaincodeEvents('CHAINCODE');

            const actual = client.getChaincodeEventsOptions()[0];
            expect(actual.deadline).toBe(chaincodeEventsOptions().deadline);
        });

        it('default call options are not modified', async () => {
            const expected = chaincodeEventsOptions().deadline;
            const deadline = Date.now() + 1000;

            await network.newChaincodeEventsRequest('CHAINCODE')
                .getEvents({ deadline });

            expect(chaincodeEventsOptions().deadline).toBe(expected);
        });

        it('throws GatewayError on call ServiceError', async () => {
            client.mockChaincodeEventsError(serviceError);

            const t = network.getChaincodeEvents('CHAINCODE');

            await expect(t).rejects.toThrow(GatewayError);
            await expect(t).rejects.toThrow(serviceError.message);
            await expect(t).rejects.toMatchObject({
                code: serviceError.code,
                cause: serviceError,
            });
        });
    });

    describe('event delivery', () => {
        const response1 = new ChaincodeEventsResponse();
        response1.setBlockNumber(1);
        response1.setEventsList([ event1, event2 ]);

        const response2 = new ChaincodeEventsResponse();
        response2.setBlockNumber(2);
        response2.setEventsList([ event3 ]);

        const expectedEvents = [
            newChaincodeEvent(1, event1),
            newChaincodeEvent(1, event2),
            newChaincodeEvent(2, event3),
        ];

        it('returns events as AsyncIterable', async () => {
            client.mockChaincodeEventsResponse(newServerStreamResponse([ response1, response2 ]));

            const events = await network.getChaincodeEvents('CHAINCODE');

            const actualEvents = await readElements(events, expectedEvents.length);
            expect(actualEvents).toEqual(expectedEvents);
        });

        it('closing iterable cancels gRPC client stream', async () => {
            const responseStream = newServerStreamResponse([ response1, response2 ]);
            client.mockChaincodeEventsResponse(responseStream);

            const events = await network.getChaincodeEvents('CHAINCODE');
            events.close();

            expect(responseStream.cancel).toBeCalled();
        });

        it('throws GatewayError on call ServiceError', async () => {
            client.mockChaincodeEventsResponse(newServerStreamResponse<ChaincodeEventsResponse>([ serviceError ]));

            const events = await network.getChaincodeEvents('CHAINCODE');
            const t = readElements(events, 1);

            await expect(t).rejects.toThrow(GatewayError);
            await expect(t).rejects.toThrow(serviceError.message);
            await expect(t).rejects.toMatchObject({
                code: serviceError.code,
                cause: serviceError,
            });
        });
    });
});
