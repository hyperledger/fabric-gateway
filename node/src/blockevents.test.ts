/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { CallOptions, Metadata, ServiceError, status } from '@grpc/grpc-js';
import { CloseableAsyncIterable } from '.';
import { BlockEventsOptions } from './blockeventsbuilder';
import { BlockEventsRequest, BlockAndPrivateDataEventsRequest, FilteredBlockEventsRequest } from './blockeventsrequest';
import { assertDefined, Gateway, internalConnect, InternalConnectOptions } from './gateway';
import { GatewayError } from './gatewayerror';
import { Identity } from './identity/identity';
import { Network } from './network';
import { Block, BlockHeader, ChannelHeader, Envelope, Payload, SignatureHeader, Status } from './protos/common/common_pb';
import { TxPvtReadWriteSet } from './protos/ledger/rwset/rwset_pb';
import { SerializedIdentity } from './protos/msp/identities_pb';
import { SeekInfo, SeekPosition } from './protos/orderer/ab_pb';
import { BlockAndPrivateData, DeliverResponse, FilteredBlock } from './protos/peer/events_pb';
import { DuplexStreamResponseStub, MockGatewayGrpcClient, newDuplexStreamResponse, readElements } from './testutils.test';

describe('Block Events', () => {
    const channelName = 'CHANNEL_NAME';
    const signature = Buffer.from('SIGNATURE');
    const serviceError: ServiceError = Object.assign(new Error('ERROR_MESSAGE'), {
        code: status.UNAVAILABLE,
        details: 'DETAILS',
        metadata: new Metadata(),
    });

    let defaultOptions: () => CallOptions;
    let client: MockGatewayGrpcClient;
    let identity: Identity;
    let signer: jest.Mock<Promise<Uint8Array>, Uint8Array[]>;
    let hash: jest.Mock<Uint8Array, Uint8Array[]>;
    let gateway: Gateway;
    let network: Network;

    beforeEach(() => {
        const now = new Date();
        const callOptions = {
            deadline: now.setHours(now.getHours() + 1),
        };
        defaultOptions = () => callOptions; // Return a specific object to test modification

        client = new MockGatewayGrpcClient();
        identity = {
            mspId: 'MSP_ID',
            credentials: new Uint8Array(Buffer.from('CERTIFICATE')),
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
            blockEventsOptions: defaultOptions,
            filteredBlockEventsOptions: defaultOptions,
            blockAndPrivateDataEventsOptions: defaultOptions,
        };
        gateway = internalConnect(options);
        network = gateway.getNetwork(channelName);
    });

    function assertValidBlockEventsRequestHeader(payload: Payload): void {
        const header = assertDefined(payload.getHeader(), 'header');
        const channelHeader = ChannelHeader.deserializeBinary(header.getChannelHeader_asU8());
        const signatureHeader = SignatureHeader.deserializeBinary(header.getSignatureHeader_asU8());
        const creator = SerializedIdentity.deserializeBinary(signatureHeader.getCreator_asU8());

        const actualCreator: Identity = {
            credentials: creator.getIdBytes_asU8(),
            mspId: creator.getMspid(),
        };

        expect(channelHeader.getChannelId()).toBe(network.getName());
        expect(actualCreator).toEqual(gateway.getIdentity());
    }

    interface TestCase {
        description: string;
        mockResponse(stream: DuplexStreamResponseStub<Envelope, DeliverResponse>): void;
        mockError(err: ServiceError): void;
        getEvents(options?: BlockEventsOptions): Promise<CloseableAsyncIterable<unknown>>;
        newEventsRequest(options?: BlockEventsOptions): BlockEventsRequest | FilteredBlockEventsRequest | BlockAndPrivateDataEventsRequest;
        getCallOptions(): CallOptions[];
        newBlockResponse(blockNumber: number): DeliverResponse;
        getBlockFromResponse(response: DeliverResponse): Block | FilteredBlock | BlockAndPrivateData | undefined;
    }

    const testCases: TestCase[] = [
        {
            description: 'Blocks',
            mockResponse(stream) {
                client.mockBlockEventsResponse(stream);
            },
            mockError(err) {
                client.mockBlockEventsError(err);
            },
            getEvents(options?) {
                return network.getBlockEvents(options);
            },
            newEventsRequest(options?) {
                return network.newBlockEventsRequest(options);
            },
            getCallOptions() {
                return client.getBlockEventsOptions();
            },
            newBlockResponse(blockNumber) {
                const header = new BlockHeader();
                header.setNumber(blockNumber);

                const block = new Block();
                block.setHeader(header);

                const response = new DeliverResponse();
                response.setBlock(block);

                return response;
            },
            getBlockFromResponse(response) {
                return response.getBlock();
            },
        },
        {
            description: 'Filtered blocks',
            mockResponse(stream) {
                client.mockFilteredBlockEventsResponse(stream);
            },
            mockError(err) {
                client.mockFilteredBlockEventsError(err);
            },
            getEvents(options?) {
                return network.getFilteredBlockEvents(options);
            },
            newEventsRequest(options?) {
                return network.newFilteredBlockEventsRequest(options);
            },
            getCallOptions() {
                return client.getFilteredBlockEventsOptions();
            },
            newBlockResponse(blockNumber) {
                const block = new FilteredBlock();
                block.setNumber(blockNumber);

                const response = new DeliverResponse();
                response.setFilteredBlock(block);

                return response;
            },
            getBlockFromResponse(response) {
                return response.getFilteredBlock();
            },
        },
        {
            description: 'Blocks and private data',
            mockResponse(stream) {
                client.mockBlockAndPrivateDataEventsResponse(stream);
            },
            mockError(err) {
                client.mockBlockAndPrivateDataEventsError(err);
            },
            getEvents(options?) {
                return network.getBlockAndPrivateDataEvents(options);
            },
            newEventsRequest(options?) {
                return network.newBlockAndPrivateDataEventsRequest(options);
            },
            getCallOptions() {
                return client.getBlockAndPrivateDataEventsOptions();
            },
            newBlockResponse(blockNumber) {
                const header = new BlockHeader();
                header.setNumber(blockNumber);

                const block = new Block();
                block.setHeader(header);

                const blockAndPrivateData = new BlockAndPrivateData();
                blockAndPrivateData.setBlock(block);
                const privateData = blockAndPrivateData.getPrivateDataMapMap();
                privateData.set(0, new TxPvtReadWriteSet());

                const response = new DeliverResponse();
                response.setBlockAndPrivateData(blockAndPrivateData);

                return response;
            },
            getBlockFromResponse(response) {
                return response.getBlockAndPrivateData();
            },
        },
    ];
    testCases.forEach(testCase => describe(`${testCase.description}`, () => {
        it('sends valid request with default start position', async () => {
            const stream = newDuplexStreamResponse<Envelope, DeliverResponse>([]);
            testCase.mockResponse(stream);

            await testCase.getEvents();

            expect(stream.write.mock.calls.length).toBe(1);
            const request = stream.write.mock.calls[0][0];

            const payload = Payload.deserializeBinary(request.getPayload_asU8());
            assertValidBlockEventsRequestHeader(payload);

            const seekInfo = SeekInfo.deserializeBinary(payload.getData_asU8());
            const start = seekInfo.getStart();
            expect(start).toBeDefined();
            expect(start?.getTypeCase()).toBe(SeekPosition.TypeCase.NEXT_COMMIT);
            expect(start?.getNextCommit()).toBeDefined();
            const stop = seekInfo.getStop();
            expect(stop).toBeDefined();
            expect(stop?.getTypeCase()).toBe(SeekPosition.TypeCase.SPECIFIED);
            expect(stop?.getSpecified()).toBeDefined();
            expect(stop?.getSpecified()?.getNumber()).toBe(Number.MAX_SAFE_INTEGER);
        });

        it('throws with negative specified start block number', async () => {
            const startBlock = BigInt(-1);
            await expect(network.getBlockEvents({ startBlock }))
                .rejects
                .toThrow();
        });

        it('sends valid request with specified start block number', async () => {
            const startBlock = BigInt(418);
            const stream = newDuplexStreamResponse<Envelope, DeliverResponse>([]);
            testCase.mockResponse(stream);

            await testCase.getEvents({ startBlock });

            expect(stream.write.mock.calls.length).toBe(1);
            const request = stream.write.mock.calls[0][0];

            const payload = Payload.deserializeBinary(request.getPayload_asU8());
            assertValidBlockEventsRequestHeader(payload);

            const seekInfo = SeekInfo.deserializeBinary(payload.getData_asU8());
            const start = seekInfo.getStart();
            expect(start).toBeDefined();
            expect(start?.getTypeCase()).toBe(SeekPosition.TypeCase.SPECIFIED);
            expect(start?.getSpecified()?.getNumber()).toBe(Number(startBlock));
            const stop = seekInfo.getStop();
            expect(stop).toBeDefined();
            expect(stop?.getTypeCase()).toBe(SeekPosition.TypeCase.SPECIFIED);
            expect(stop?.getSpecified()).toBeDefined();
            expect(stop?.getSpecified()?.getNumber()).toBe(Number.MAX_SAFE_INTEGER);
        });

        it('uses specified call options', async () => {
            const deadline = Date.now() + 1000;

            await testCase.newEventsRequest()
                .getEvents({ deadline });

            const actual = testCase.getCallOptions()[0];
            expect(actual.deadline).toBe(deadline);
        });

        it('uses default call options', async () => {
            await testCase.getEvents();

            const actual = testCase.getCallOptions()[0];
            expect(actual.deadline).toBe(defaultOptions().deadline);
        });

        it('default call options are not modified', async () => {
            const expected = defaultOptions().deadline;
            const deadline = Date.now() + 1000;

            await network.newBlockEventsRequest()
                .getEvents({ deadline });

            expect(defaultOptions().deadline).toBe(expected);
        });

        it('throws GatewayError on call ServiceError', async () => {
            testCase.mockError(serviceError);

            const t = testCase.getEvents();

            await expect(t).rejects.toThrow(GatewayError);
            await expect(t).rejects.toThrow(serviceError.message);
            await expect(t).rejects.toMatchObject({
                code: serviceError.code,
                cause: serviceError,
            });
        });

        it('throws GatewayError on receive ServiceError', async () => {
            const stream = newDuplexStreamResponse<Envelope, DeliverResponse>([serviceError]);
            testCase.mockResponse(stream);

            const events = await testCase.getEvents();
            const t = readElements(events, 1);

            await expect(t).rejects.toThrow(GatewayError);
            await expect(t).rejects.toThrow(serviceError.message);
            await expect(t).rejects.toMatchObject({
                code: serviceError.code,
                cause: serviceError,
            });
        });

        it('throws on receive of status message', async () => {
            const response = new DeliverResponse();
            response.setStatus(Status.SERVICE_UNAVAILABLE);
            const stream = newDuplexStreamResponse<Envelope, DeliverResponse>([response]);
            testCase.mockResponse(stream);

            const events = await testCase.getEvents();
            const t = readElements(events, 1);

            await expect(t).rejects.toThrow(String(Status.SERVICE_UNAVAILABLE));
        });

        it('throws on receive of unexpected message type', async () => {
            const response = new DeliverResponse();
            const stream = newDuplexStreamResponse<Envelope, DeliverResponse>([response]);
            testCase.mockResponse(stream);

            const events = await testCase.getEvents();
            const t = readElements(events, 1);

            await expect(t).rejects.toThrow(String(response.getTypeCase()));
        });

        it('receives events', async () => {
            const responses = [testCase.newBlockResponse(1), testCase.newBlockResponse(2)];
            const stream = newDuplexStreamResponse<Envelope, DeliverResponse>(responses);
            testCase.mockResponse(stream);

            const events = await testCase.getEvents();
            const actual = await readElements(events, responses.length);

            const expected = responses.map(response => testCase.getBlockFromResponse(response));

            expect(actual).toEqual(expected);
        });

        it('closing iterator cancels gRPC stream', async () => {
            const responses = [testCase.newBlockResponse(1), testCase.newBlockResponse(2)];
            const stream = newDuplexStreamResponse<Envelope, DeliverResponse>(responses);
            testCase.mockResponse(stream);

            const events = await testCase.getEvents();
            events.close();

            expect(stream.cancel).toBeCalled();
        });
    }));
});
