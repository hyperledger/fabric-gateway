/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { DefaultCheckPointers } from './defaultcheckpointers';

    describe('Inmemory checkpointer', () => {

        it('No inmemory checkpointer exist: expected to initialize checkpointer state ', async () => {

            const inMemoryCheckPointer = DefaultCheckPointers.inMemory();

            const actualBlockNumber = await inMemoryCheckPointer.getBlockNumber();
            expect(actualBlockNumber).toBeUndefined();

            const actualTxnIDS = await inMemoryCheckPointer.getTransactionIds();
            expect(actualTxnIDS.length).toEqual(0);
        });


        it('Checkpointing only a block number in a fresh checkpointer gives block number & no transactions', async () => {

            const expectedBlockNumber = BigInt(101);

            const inMemoryCheckPointer = DefaultCheckPointers.inMemory();
            await inMemoryCheckPointer.checkpoint(expectedBlockNumber);

            const actualBlockNumber = await inMemoryCheckPointer.getBlockNumber();
            expect(actualBlockNumber).toEqual(expectedBlockNumber);

            const actualTxnIDS = await inMemoryCheckPointer.getTransactionIds();
            expect(actualTxnIDS.length).toEqual(0);
        });

        it('Checkpointing same block number and new transaction in used checkpointer gives block number and expected transactions', async () => {
            const expectedBlockNumber = BigInt(101);
            const expectedTxnIDs = ['txn1'];

            const inMemoryCheckPointer = DefaultCheckPointers.inMemory();
            await inMemoryCheckPointer.checkpoint(expectedBlockNumber)
            const actualBlockNumber = await inMemoryCheckPointer.getBlockNumber();
            expect(actualBlockNumber).toEqual(expectedBlockNumber);

            await inMemoryCheckPointer.checkpoint(expectedBlockNumber,'txn1');

            const actualTxnIDS = await inMemoryCheckPointer.getTransactionIds();
            expect(actualTxnIDS).toStrictEqual(expectedTxnIDs);
        });

        it('Checkpointing block and transaction in a fresh checkpointer, gives block number and transaction', async () => {

            const expectedBlockNumber = BigInt(101);
            const expectedTxnIDs = ['txn1'];

            const inMemoryCheckPointer = DefaultCheckPointers.inMemory();
            await inMemoryCheckPointer.checkpoint(expectedBlockNumber,'txn1')

            const actualBlockNumber = await inMemoryCheckPointer.getBlockNumber();
            expect(actualBlockNumber).toEqual(expectedBlockNumber);

            const actualTxnIDS = await inMemoryCheckPointer.getTransactionIds();
            expect(actualTxnIDS).toStrictEqual(expectedTxnIDs);
        });

        it('Checkpointing only a new block in used checkpointer gives new block number and no transactions', async () => {

            const blockNumber1 = BigInt(101);

            const blockNumber2 = BigInt(102);

            const expectedTxnIDs: any[] = [];

            const inMemoryCheckPointer = DefaultCheckPointers.inMemory();
            await inMemoryCheckPointer.checkpoint(blockNumber1,'txn1');

            let actualBlockNumber = await inMemoryCheckPointer.getBlockNumber();
            expect(actualBlockNumber).toStrictEqual(blockNumber1);

            await inMemoryCheckPointer.checkpoint(blockNumber2);
            actualBlockNumber = await inMemoryCheckPointer.getBlockNumber();
            expect(actualBlockNumber).toStrictEqual(blockNumber2);
            const actualTxnIDS = await inMemoryCheckPointer.getTransactionIds();
            expect(actualTxnIDS).toStrictEqual(expectedTxnIDs);

        });

            it('Checkpointing new block and transaction in used checkpointer gives new block and only new transaction', async () => {

                const blockNumber1 = BigInt(101);
                const blockNumber2 = BigInt(102);

                const expectedTxnIDs: string[] = ['txn2','txn3'];

                const inMemoryCheckPointer = DefaultCheckPointers.inMemory();
                await inMemoryCheckPointer.checkpoint(blockNumber1,'txn1');

                let actualBlockNumber = await inMemoryCheckPointer.getBlockNumber();
                expect(actualBlockNumber).toStrictEqual(blockNumber1);

                await inMemoryCheckPointer.checkpoint(blockNumber2);
                actualBlockNumber = await inMemoryCheckPointer.getBlockNumber();

                await inMemoryCheckPointer.checkpoint(blockNumber2,'txn2');
                await inMemoryCheckPointer.checkpoint(blockNumber2,'txn3');

                expect(actualBlockNumber).toStrictEqual(blockNumber2);
                const actualTxnIDS = await inMemoryCheckPointer.getTransactionIds();
                expect(actualTxnIDS).toStrictEqual(expectedTxnIDs);

            });
    })
