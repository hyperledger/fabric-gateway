/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CallOptions, Metadata, ServiceError, status } from '@grpc/grpc-js';
import { ChaincodeEvent } from './chaincodeevent';
import * as Checkpointers  from './checkpointers';
import { Gateway, internalConnect, InternalConnectOptions } from './gateway';
import { GatewayError } from './gatewayerror';
import { Identity } from './identity/identity';
import { Network } from './network';
import { ChaincodeEventsRequest as ChaincodeEventsRequestProto, ChaincodeEventsResponse, SignedChaincodeEventsRequest } from './protos/gateway/gateway_pb';
import { SeekPosition } from './protos/orderer/ab_pb';
import { ChaincodeEvent as ChaincodeEventProto } from './protos/peer/chaincode_event_pb';
import { getEventsIterable, MockGatewayGrpcClient, newServerStreamResponse, readElements, readEventsAndCheckpoint, } from './testutils.test';

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
    let signer: jest.Mock<Promise<Uint8Array>, Uint8Array[]>;
    let hash: jest.Mock<Uint8Array, Uint8Array[]>;
    let gateway: Gateway;
    let network: Network;


    const event1 = new ChaincodeEventProto();
    event1.setChaincodeId('CHAINCODE');
    event1.setTxId('tx-10-1');
    event1.setEventName('event1'),
    event1.setPayload(Buffer.from('payload1'));

    const event2 = new ChaincodeEventProto();
    event2.setChaincodeId('CHAINCODE');
    event2.setTxId('tx-10-2');
    event2.setEventName('event2'),
    event2.setPayload(Buffer.from('payload2'));

    const event3 = new ChaincodeEventProto();
    event3.setChaincodeId('CHAINCODE');
    event3.setTxId('tx-10-3');
    event3.setEventName('event3'),
    event3.setPayload(Buffer.from('payload3'));

    const event4 = new ChaincodeEventProto();
    event4.setChaincodeId('CHAINCODE');
    event4.setTxId('tx-10-4');
    event4.setEventName('event4'),
    event4.setPayload(Buffer.from('payload4'));

    const event5 = new ChaincodeEventProto();
    event5.setChaincodeId('CHAINCODE');
    event5.setTxId('tx-20-1');
    event5.setEventName('event5'),
    event5.setPayload(Buffer.from('payload5'));

    const event6 = new ChaincodeEventProto();
    event6.setChaincodeId('CHAINCODE');
    event6.setTxId('tx-20-2');
    event6.setEventName('event6'),
    event6.setPayload(Buffer.from('payload6'));

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
        signer = jest.fn(undefined);
        signer.mockResolvedValue(signature);
        hash = jest.fn(undefined);
        hash.mockReturnValue(Buffer.from('DIGEST'));

        const options: InternalConnectOptions = {
            identity,
            signer,
            hash,
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

            expect(request.getChannelId()).toBe(channelName);
            expect(request.getChaincodeId()).toBe('CHAINCODE');

            const startPosition = request.getStartPosition();
            expect(startPosition).toBeDefined();
            expect(startPosition?.getTypeCase()).toBe(SeekPosition.TypeCase.NEXT_COMMIT);
            expect(startPosition?.getNextCommit()).toBeDefined();
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

            expect(request.getChannelId()).toBe(channelName);
            expect(request.getChaincodeId()).toBe('CHAINCODE');

            const startPosition = request.getStartPosition();
            expect(startPosition).toBeDefined();
            expect(startPosition?.getTypeCase()).toBe(SeekPosition.TypeCase.SPECIFIED);
            expect(startPosition?.getSpecified()?.getNumber()).toBe(Number(startBlock));
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

        const expectedEvents: ChaincodeEvent[] = [
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

        it('closing iterator cancels gRPC client stream', async () => {
            const iterable = newServerStreamResponse([ response1, response2 ]);
            client.mockChaincodeEventsResponse(iterable);

            const events = await network.getChaincodeEvents('CHAINCODE');
            events.close();

            expect(iterable.cancel).toBeCalled();
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

    describe('checkpoint Events', () => {
        it('Fresh checkpointer gives all the events emitted', async () => {
            const eventsInput: ChaincodeEvent[] = [
                newChaincodeEvent(10, event1),
                newChaincodeEvent(10, event2),
                newChaincodeEvent(20, event3),
            ];

            const eventsIterator = getEventsIterable(eventsInput);
            const checkPointerInstance = Checkpointers.inMemory();
            const checkpointChaincodeEvents = Checkpointers.checkpointChaincodeEvents(eventsIterator, checkPointerInstance);
            const expectedEventsLength = 3;
            const actualEvents = await readElements(checkpointChaincodeEvents, expectedEventsLength);
            expect(actualEvents).toEqual(eventsInput);
        });

        it('Used checkpointer does not emit duplicate events if checkpointed  within a block', async () => {
            const eventsInput: ChaincodeEvent[] = [
                newChaincodeEvent(10, event1),
                newChaincodeEvent(10, event2),
                newChaincodeEvent(20, event3),
            ];
            const eventsIterator = getEventsIterable(eventsInput);
            const checkPointerInstance = Checkpointers.inMemory();
            const checkpointChaincodeEvents1 = Checkpointers.checkpointChaincodeEvents(eventsIterator, checkPointerInstance);
            const eventsToCheckpoint = 1;
            await readEventsAndCheckpoint(checkpointChaincodeEvents1, eventsToCheckpoint);

            const checkpointChaincodeEvents2 = Checkpointers.checkpointChaincodeEvents(eventsIterator, checkPointerInstance);

            const expectedEventsLength = 2;
            const eventsReceived = await readElements(checkpointChaincodeEvents2, expectedEventsLength);
            expect(eventsReceived[0].blockNumber).toEqual(BigInt(10));
            expect(eventsReceived[0].transactionId).toEqual('tx-10-2');
        });

        it('Used checkpointer does not emit duplicate events if checkpointed at the end of a block', async () => {
            const eventsInput: ChaincodeEvent[] = [
                newChaincodeEvent(10, event1),
                newChaincodeEvent(10, event2),
                newChaincodeEvent(20, event5),
            ];
            const eventsIterator = getEventsIterable(eventsInput);
            const checkPointerInstance = Checkpointers.inMemory();
            const checkpointChaincodeEvents1 = Checkpointers.checkpointChaincodeEvents(eventsIterator, checkPointerInstance);
            const eventsToCheckpoint = 2;
            await readEventsAndCheckpoint(checkpointChaincodeEvents1, eventsToCheckpoint);

            const checkpointChaincodeEvents2 = Checkpointers.checkpointChaincodeEvents(eventsIterator, checkPointerInstance);

            const expectedEventsLength = 2;
            const eventsReceived = await readElements(checkpointChaincodeEvents2, expectedEventsLength);

            expect(eventsReceived[0].blockNumber).toEqual(BigInt(20));
            expect(eventsReceived[0].transactionId).toEqual('tx-20-1');
        });

        it('Used checkpointer does not emit duplicate events if checkpointed events at the beginning of a new block', async () => {
            const eventsInput: ChaincodeEvent[] = [
                newChaincodeEvent(10, event1),
                newChaincodeEvent(20, event5),
                newChaincodeEvent(20, event6),
            ];
            const eventsIterator = getEventsIterable(eventsInput);
            const checkPointerInstance = Checkpointers.inMemory();
            const checkpointChaincodeEvents1 = Checkpointers.checkpointChaincodeEvents(eventsIterator, checkPointerInstance);
            const eventsToCheckpoint = 2;
            await readEventsAndCheckpoint(checkpointChaincodeEvents1, eventsToCheckpoint);

            const checkpointChaincodeEvents2 = Checkpointers.checkpointChaincodeEvents(eventsIterator, checkPointerInstance);

            const expectedEventsLength = 1;
            const eventsReceived = await readElements(checkpointChaincodeEvents2, expectedEventsLength);

            expect(eventsReceived[0].blockNumber).toEqual(BigInt(20));
            expect(eventsReceived[0].transactionId).toEqual('tx-20-2');
        });

        it('events close get called', async () => {
            const eventsInput: ChaincodeEvent[] = [
                newChaincodeEvent(10, event1),
                newChaincodeEvent(10, event2),
                newChaincodeEvent(20, event3),
            ];
            const eventsIterator = getEventsIterable(eventsInput);

            const checkPointerInstance = Checkpointers.inMemory();
            const checkpointChaincodeEvents = Checkpointers.checkpointChaincodeEvents(eventsIterator, checkPointerInstance);
            const eventsToCheckpoint = 1;
            await readEventsAndCheckpoint(checkpointChaincodeEvents, eventsToCheckpoint);

            checkpointChaincodeEvents.close();
            expect(eventsIterator.close).toBeCalledTimes(1);
        });

        it('Resume eventing correctly after an error thrown,', async () => {
            const response1 = new ChaincodeEventsResponse();
            response1.setBlockNumber(10);
            response1.setEventsList([ event1, event2 ]);

            const response2 = new ChaincodeEventsResponse();
            response2.setBlockNumber(20);
            response2.setEventsList([ event5 ]);

            client.mockChaincodeEventsResponse(newServerStreamResponse([ response1, serviceError ]));
            const checkpointerInstance = Checkpointers.inMemory();

            const events1 = await network.getChaincodeEvents('CHAINCODE');
            const checkpointEvents1 = Checkpointers.checkpointChaincodeEvents(events1, checkpointerInstance);

            const eventsToCheckpoint = 3;
            const checkPointedEvents = readEventsAndCheckpoint(checkpointEvents1, eventsToCheckpoint);

            await expect(checkPointedEvents).rejects.toThrow(GatewayError);
            await expect(checkPointedEvents).rejects.toMatchObject({
                code: serviceError.code,
                cause: serviceError,
            });
            events1.close();

            client.mockChaincodeEventsResponse(newServerStreamResponse([ response1, response2 ]));
            const events2 = await network.getChaincodeEvents('CHAINCODE');
            const checkpointEvents2 = Checkpointers.checkpointChaincodeEvents(events2, checkpointerInstance);
            const expectedEventsLength1 = 1;

            const eventsReceived2 = await readElements<ChaincodeEvent>(checkpointEvents2, expectedEventsLength1);

            expect(eventsReceived2[0].blockNumber).toEqual(BigInt(20));
            expect(eventsReceived2[0].transactionId).toEqual('tx-20-1');
        });
    });
});
