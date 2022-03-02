/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import * as checkpointers from './checkpointers';
import fs from 'fs';
import { createTempDir,rmdir } from './testutils.test';
import path from 'path';
import { Checkpointer } from './checkpointer';

describe('Checkpointers', () => {
    let filePath:string;
    function getInmemoryInstance(): Promise<Checkpointer> {
        return Promise.resolve(checkpointers.inMemory());
    }
    async function getFileCheckpointerInstance(): Promise<Checkpointer> {
        return await checkpointers.file(filePath);
    }
    async function createCheckpointerFile(): Promise<string> {
        const dir = await createTempDir();
        filePath = path.join(dir, 'checkpoint.json');
        return dir;
    }
    async  function cleanup(dir:string):Promise<void>{
        await rmdir(dir);
    }

    function noOperation():Promise<void> {
        return Promise.resolve();
    }
    const checkpointerTypes = [
        { getInstance: getInmemoryInstance, description: 'In-memory checkpointer',createFile: noOperation ,cleanup: noOperation},
        { getInstance: getFileCheckpointerInstance, description: 'File checkpointer',createFile: createCheckpointerFile ,cleanup: cleanup},
    ];
    checkpointerTypes.forEach(checkpointer => {
        describe(`${checkpointer.description}`,() => {
            let dir:string|void;

            beforeEach(async() => {
                dir = await checkpointer.createFile();
            });
            afterEach(async() => {
                if(dir){
                    await checkpointer.cleanup(dir);

                }
            });

            it('Initializes default checkpointer state when no checkpointer already exist', async () => {
                const checkPointerInstance = await checkpointer.getInstance();
                expect(checkPointerInstance.getBlockNumber()).toBeUndefined();
                expect(checkPointerInstance.getTransactionIds().size).toEqual(0);
            });

            it('Checkpointing only a block number in a fresh checkpointer gives block number & no transactions', async () => {
                const blockNumber = BigInt(101);
                const checkPointerInstance = await checkpointer.getInstance();
                await checkPointerInstance.checkpoint(blockNumber);
                expect(checkPointerInstance.getBlockNumber()).toEqual(blockNumber);
                expect(checkPointerInstance.getTransactionIds().size).toEqual(0);
            });
            it('Checkpointing same block number and new transaction in used checkpointer gives block number and expected transactions', async () => {
                const blockNumber = BigInt(101);
                const checkPointerInstance = await checkpointer.getInstance();
                await checkPointerInstance.checkpoint(blockNumber);
                await checkPointerInstance.checkpoint(blockNumber, 'txn1');
                expect(checkPointerInstance.getTransactionIds()).toStrictEqual(
                    new Set(['txn1'])
                );
            });
            it('Checkpointing block and transaction in a fresh checkpointer, gives block number and transaction', async () => {
                const blockNumber = BigInt(101);
                const checkPointerInstance = await checkpointer.getInstance();
                await checkPointerInstance.checkpoint(blockNumber, 'txn1');
                expect(checkPointerInstance.getBlockNumber()).toEqual(blockNumber);
                return expect(checkPointerInstance.getTransactionIds()).toStrictEqual(
                    new Set(['txn1'])
                );
            });

            it('Checkpointing only a new block in used checkpointer gives new block number and no transactions', async () => {
                const blockNumber1 = BigInt(101);
                const blockNumber2 = BigInt(102);
                const checkPointerInstance = await checkpointer.getInstance();
                await checkPointerInstance.checkpoint(blockNumber1, 'txn1');
                await checkPointerInstance.checkpoint(blockNumber2);
                expect(checkPointerInstance.getBlockNumber()).toStrictEqual(blockNumber2);
                expect(checkPointerInstance.getTransactionIds()).toStrictEqual(new Set());
            });

            it('Checkpointing new block and transaction in used checkpointer gives new block and only new transaction', async () => {
                const blockNumber1 = BigInt(101);
                const blockNumber2 = BigInt(102);
                const checkPointerInstance = await checkpointer.getInstance();

                await checkPointerInstance.checkpoint(blockNumber1, 'txn1');
                await checkPointerInstance.checkpoint(blockNumber2, 'txn2');
                await checkPointerInstance.checkpoint(blockNumber2, 'txn3');
                expect(checkPointerInstance.getBlockNumber()).toStrictEqual(blockNumber2);
                expect(checkPointerInstance.getTransactionIds()).toStrictEqual(
                    new Set(['txn2', 'txn3'])
                );
            });
        });
    })
    describe('File Checkpointer: Test file creation and initialization', () => {
        let dir:string;
        let checkpointerPath:string;
        beforeEach(async() => {
            dir = await createTempDir();
            checkpointerPath = path.join(dir, 'checkpoint.json')
        });
        afterEach(async() => {
            await rmdir(dir);
        });

        it('In the absence of a checkpointer file , a new one gets generated', async () => {
            expect(fs.existsSync(checkpointerPath)).toEqual(false);
            await checkpointers.file(checkpointerPath);
            expect(fs.existsSync(checkpointerPath)).toEqual(true);
        });

        it('checkpointer loads the already existing state', async () => {
            //load file checkpointer with checkpointer state
            const blockNumber = BigInt('101');
            const checkPointerInstance1 = await checkpointers.file(checkpointerPath)
            await checkPointerInstance1.checkpoint(blockNumber);

            const checkPointerInstance2 = await checkpointers.file(checkpointerPath)
            expect(checkPointerInstance2.getBlockNumber()).toEqual(blockNumber);
            expect(checkPointerInstance2.getTransactionIds().size).toEqual(0);
        });

        it('Checkpointing block number zero in a fresh checkpointer sets the state as zero', async () => {
            const blockNumber = BigInt(0);
            const checkPointerInstance1 = await checkpointers.file(checkpointerPath);
            await checkPointerInstance1.checkpoint(blockNumber);
            const checkPointerInstance2 = await checkpointers.file(checkpointerPath);
            expect(checkPointerInstance1.getBlockNumber()).toEqual(checkPointerInstance2.getBlockNumber());
        });
    });
})