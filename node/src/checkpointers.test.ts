/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { promises as fs } from 'fs';
import * as path from 'path';
import { ChaincodeEvent } from './chaincodeevent';
import { Checkpointer } from './checkpointer';
import * as checkpointers from './checkpointers';
import { createTempDir } from './testutils.test';

/* eslint-disable jest/expect-expect */

describe('Checkpointers', () => {
    let tempDir: string;
    let checkpointFile: string;

    beforeAll(async () => {
        tempDir = await createTempDir();
        checkpointFile = path.join(tempDir, 'checkpoint.json');
    });

    afterAll(async () => {
        await fs.rm(tempDir, { recursive: true, force: true });
    });

    function assertState(checkpointer: Checkpointer, blockNumber: bigint | undefined, transactionId?: string): void {
        expect(checkpointer.getBlockNumber()).toBe(blockNumber);
        expect(checkpointer.getTransactionId()).toEqual(transactionId);
    }

    const testCases = [
        {
            description: 'In-memory',
            after: () => Promise.resolve(),
            newCheckpointer: () => Promise.resolve(checkpointers.inMemory()),
        },
        {
            description: 'File',
            after: () => fs.rm(checkpointFile, { force: true }),
            newCheckpointer: () => checkpointers.file(checkpointFile),
        },
    ];

    testCases.forEach(testCase => {
        describe(`${testCase.description} common behaviour`, () => {
            let checkpointer: Checkpointer;

            beforeEach(async () => {
                checkpointer = await testCase.newCheckpointer();
            });

            afterEach(async () => {
                await testCase.after();
            });

            it('Initial state is undefined block and no transactions', () => {
                assertState(checkpointer, undefined);
            });

            it('Checkpointing a block gives next block number & empty transaction ID', async () => {
                await checkpointer.checkpointBlock(1n);

                assertState(checkpointer, 1n + 1n);
            });

            it('Checkpointing a transaction gives valid transaction ID and blocknumber', async () => {
                await checkpointer.checkpointTransaction(1n, 'tx1');

                assertState(checkpointer, 1n, 'tx1');
            });

            it('Checkpointing a chaincode event gives valid transaction ID and blocknumber', async () => {
                const event: ChaincodeEvent = {
                    blockNumber: BigInt(1),
                    chaincodeName: 'CHAINCODE',
                    eventName: 'EVENT1',
                    transactionId: 'TXN1',
                    payload: new Uint8Array(),
                };

                await checkpointer.checkpointChaincodeEvent(event);

                assertState(checkpointer, event.blockNumber, event.transactionId);
            });
        });
    });

    describe('File-specific behaviour', () => {
        afterEach(async () => {
            await fs.rm(checkpointFile, { force: true });
        });

        it('throws on unwritable file location', async () => {
            const badFile = path.join(tempDir, 'MISSING_DIRECTORY', 'checkpoint.json');
            await expect(checkpointers.file(badFile)).rejects.toThrow();
        });

        it('state is persisted', async () => {
            const expected = await checkpointers.file(checkpointFile);
            await expected.checkpointTransaction(1n, 'tx1');

            const actual = await checkpointers.file(checkpointFile);

            expect(actual.getBlockNumber()).toBe(expected.getBlockNumber());
            expect(actual.getTransactionId()).toEqual(expected.getTransactionId());
        });

        it('block number zero is persisted correctly', async () => {
            const expected = await checkpointers.file(checkpointFile);
            await expected.checkpointBlock(0n);

            const actual = await checkpointers.file(checkpointFile);

            expect(actual.getBlockNumber()).toBe(0n + 1n);
        });
    });
});
