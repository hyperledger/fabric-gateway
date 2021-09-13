/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { ChaincodeEvent } from './chaincodeevent';
import { MockGatewayClient, newMockGatewayClient } from './client.test';
import { Gateway, internalConnect, InternalConnectOptions } from './gateway';
import { Identity } from './identity/identity';
import { Network } from './network';
import { ChaincodeEventsRequest as ChaincodeEventsRequestProto, ChaincodeEventsResponse, SignedChaincodeEventsRequest } from './protos/gateway/gateway_pb';
import { SeekPosition } from './protos/orderer/ab_pb';
import { ChaincodeEvent as ChaincodeEventProto } from './protos/peer/chaincode_event_pb';

function assertDecodeChaincodeEventsRequest(signedRequest: SignedChaincodeEventsRequest): ChaincodeEventsRequestProto {
    const requestBytes = signedRequest.getRequest_asU8();
    expect(requestBytes).toBeDefined();
    return ChaincodeEventsRequestProto.deserializeBinary(requestBytes);
}

function newChaincodeEvent(blockNumber: number, event: ChaincodeEventProto): ChaincodeEvent {
    return {
        blockNumber: BigInt(blockNumber),
        chaincodeId: event.getChaincodeId() ?? '',
        eventName: event.getEventName() ?? '',
        transactionId: event.getTxId() ?? '',
        payload: event.getPayload_asU8() ?? new Uint8Array(),
    };
}

function newAsyncIterable<T>(values: T[]): AsyncIterable<T> {
    return {
        async* [Symbol.asyncIterator]() { // eslint-disable-line @typescript-eslint/require-await
            for (const value of values) {
                yield value;
            }
        }
    }
}

function mockAsyncListener<T>(expectedCallCount = 1): {
    listener: (event: T) => Promise<void>,
    mock: jest.Mock<Promise<void>, T[]>,
    complete: Promise<void>,
} {
    let resolve: () => void;
    const complete = new Promise<void>(_resolve => resolve = _resolve);

    const mock = jest.fn<Promise<void>, T[]>();
    const listener = async (event: T): Promise<void> => {
        try {
            await mock(event);
        } finally {
            expectedCallCount--;
            if (expectedCallCount === 0) {
                resolve();
            }
        }
    };
    return { listener, mock, complete }
}

async function readElements<T>(iter: AsyncIterable<T>, count: number): Promise<T[]> {
    const elements: T[] = [];
    for await (const element of iter) {
        elements.push(element);

        if (--count <= 0) {
            break;
        }
    }

    return elements;
}

describe('Chaincode Events', () => {
    const channelName = 'CHANNEL_NAME';
    const signature = Buffer.from('SIGNATURE');
    let client: MockGatewayClient;
    let identity: Identity;
    let signer: jest.Mock<Promise<Uint8Array>, Uint8Array[]>;
    let hash: jest.Mock<Uint8Array, Uint8Array[]>;
    let gateway: Gateway;
    let network: Network;

    beforeEach(() => {
        client = newMockGatewayClient();
        identity = {
            mspId: 'MSP_ID',
            credentials: Buffer.from('CERTIFICATE'),
        }
        signer = jest.fn(undefined);
        signer.mockResolvedValue(signature);
        hash = jest.fn(undefined);
        hash.mockReturnValue(Buffer.from('DIGEST'));

        const options: InternalConnectOptions = {
            identity,
            signer,
            hash,
            gatewayClient: client,
        };
        gateway = internalConnect(options);
        network = gateway.getNetwork(channelName);
    });

    describe('request', () => {
        it('throws on connection error', async () => {
            client.chaincodeEvents.mockImplementation(() => {
                throw new Error('CONNECTION_ERROR');
            });
    
            await expect(network.getChaincodeEvents('CHAINCODE'))
                .rejects
                .toThrow('CONNECTION_ERROR');
        });
    
        it('sends valid request with default start position', async () => {
            await network.getChaincodeEvents('CHAINCODE');
    
            const signedRequest = client.chaincodeEvents.mock.calls[0][0];
            expect(signedRequest.getSignature()).toEqual(signature);
    
            const request = assertDecodeChaincodeEventsRequest(signedRequest);
    
            expect(request.getChannelId()).toBe(channelName);
            expect(request.getChaincodeId()).toBe('CHAINCODE');
    
            const startPosition = request.getStartPosition();
            expect(startPosition).toBeDefined();
            expect(startPosition?.getTypeCase()).toBe(SeekPosition.TypeCase.NEXT_COMMIT);
            expect(startPosition?.getNextCommit()).toBeDefined();
        });
    
        it('sends valid request with specified start block number', async () => {
            const startBlock = BigInt(418);
            await network.getChaincodeEvents('CHAINCODE', { startBlock: startBlock });
    
            const signedRequest = client.chaincodeEvents.mock.calls[0][0];
            expect(signedRequest.getSignature()).toEqual(signature);
    
            const request = assertDecodeChaincodeEventsRequest(signedRequest);
    
            expect(request.getChannelId()).toBe(channelName);
            expect(request.getChaincodeId()).toBe('CHAINCODE');
    
            const startPosition = request.getStartPosition();
            expect(startPosition).toBeDefined();
            expect(startPosition?.getTypeCase()).toBe(SeekPosition.TypeCase.SPECIFIED);
            expect(startPosition?.getSpecified()?.getNumber()).toBe(Number(startBlock));
        });
    });

    describe('event delivery', () => {
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
            client.chaincodeEvents.mockReturnValue(newAsyncIterable([ response1, response2 ]));
    
            const events = await network.getChaincodeEvents('CHAINCODE');
    
            const actualEvents = await readElements(events, expectedEvents.length);
            expect(actualEvents).toEqual(expectedEvents);
        });

        it('delivers events to listener callback', async () => {
            client.chaincodeEvents.mockReturnValue(newAsyncIterable([ response1, response2 ]));
    
            const { listener, mock, complete } = mockAsyncListener<ChaincodeEvent | unknown>(3);
    
            void network.newChaincodeEventsRequest('CHAINCODE')
                .onEvent((err, event) => listener(err || event));
    
            await complete;
            expect(mock.mock.calls.flat()).toEqual(expectedEvents);
        });
    
        it('delivers errors to listener callback', async () => {
            const expectedError = new Error('EXPECTED_ERROR');
            const asyncIter: AsyncIterable<ChaincodeEventsResponse> = {
                async* [Symbol.asyncIterator]() { // eslint-disable-line @typescript-eslint/require-await,require-yield
                    throw expectedError;
                }
            }
            client.chaincodeEvents.mockReturnValue(asyncIter);
    
            const { listener, mock, complete } = mockAsyncListener<ChaincodeEvent | unknown>(1);
    
            void network.newChaincodeEventsRequest('CHAINCODE')
                .onEvent((err, event) => listener(err || event));
    
            await complete;
            expect(mock.mock.calls.flat()).toEqual([expectedError]);
        });
    
        it('listener callback error does not stop event delivery', async () => {
            client.chaincodeEvents.mockReturnValue(newAsyncIterable([ response1, response2 ]));
    
            const { listener, mock, complete } = mockAsyncListener<ChaincodeEvent | unknown>(3);
            mock.mockRejectedValueOnce(new Error('EXPECTED_ERROR'));
    
            void network.newChaincodeEventsRequest('CHAINCODE')
                .onEvent((err, event) => listener(err || event));
    
            await complete;
            expect(mock.mock.calls.flat()).toEqual(expectedEvents);
        });
    })
});
