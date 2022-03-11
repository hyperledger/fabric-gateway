/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { promises as fs } from 'fs';
import * as path from 'path';
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

    function assertState(checkpointer: Checkpointer, blockNumber: bigint | undefined, ...transactionIds: string[]): void {
        expect(checkpointer.getBlockNumber()).toBe(blockNumber);
        expect(checkpointer.getTransactionIds()).toEqual(new Set(transactionIds));
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

            it('Checkpoint only block stores block and no transactions', async () => {
                await checkpointer.checkpoint(1n);

                assertState(checkpointer, 1n);
            });

            it('Checkpoint block and transaction stores block and transaction', async () => {
                await checkpointer.checkpoint(1n, 'tx1');

                assertState(checkpointer, 1n, 'tx1');
            });

            it('Checkpoint same block and new transactions stores transactions', async () => {
                await checkpointer.checkpoint(1n);
                await checkpointer.checkpoint(1n, 'tx1');
                await checkpointer.checkpoint(1n, 'tx2');

                assertState(checkpointer, 1n, 'tx1', 'tx2');
            });

            it('Checkpoint new block clears existing transactions', async () => {
                await checkpointer.checkpoint(1n, 'tx1');
                await checkpointer.checkpoint(2n);

                assertState(checkpointer, 2n);
            });

            it('Checkpoint new block and transaction clears existing transactions', async () => {
                await checkpointer.checkpoint(1n, 'tx1');
                await checkpointer.checkpoint(2n, 'tx2');

                assertState(checkpointer, 2n, 'tx2');
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
            await expected.checkpoint(1n, 'tx1');

            const actual = await checkpointers.file(checkpointFile);

            expect(actual.getBlockNumber()).toBe(expected.getBlockNumber());
            expect(actual.getTransactionIds()).toEqual(expected.getTransactionIds());
        });

        it('block number zero is persisted correctly', async () => {
            const expected = await checkpointers.file(checkpointFile);
            await expected.checkpoint(0n);

            const actual = await checkpointers.file(checkpointFile);

            expect(actual.getBlockNumber()).toBe(0n);
        });
    });
});
