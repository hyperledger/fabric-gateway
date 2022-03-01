/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import path from 'path';
import fs from 'fs';
import { DefaultCheckPointers } from './defaultcheckpointers';
import { CheckPointerState } from './checkpointer';

async function writeToFile(checkpointerState:any,path:string):Promise<void>{
    const data = JSON.stringify(checkpointerState, (_key, value) =>{
    value = typeof value === "bigint" ? value.toString() + "n" : value
    return value;
    });
    await fs.promises.writeFile(path, data);
}

async function readFromFile(path:string): Promise<CheckPointerState>{

    const data = (await fs.promises.readFile(path)).toString();

    return JSON.parse(data,(key, value) => {
        if (typeof value === "string" && /^\d+n$/.test(value)) {
          return BigInt(value.substr(0, value.length - 1));
        }
        return value;
      }) as CheckPointerState;
}

    describe('File checkpointer', () => {

        const dir = './testUtils';
        const fileName = 'checkpoint.json';
        const checkPointerFilePath = path.join(dir, fileName);

        beforeEach(async () => {
            if (!fs.existsSync(dir)){
                fs.mkdirSync(dir);
            }
            await writeToFile('', checkPointerFilePath);
        });
        afterAll(async ()=>{
            if (fs.existsSync(dir)){
                fs.rmSync(dir, { recursive: true, force: true });
            }
        })


        it('No checkpointer file exist: expected to create new file with checkpointer state initialized', async () => {

            const defaultCheckPointers = await DefaultCheckPointers.file(checkPointerFilePath);
            expect(fs.existsSync(checkPointerFilePath)).toEqual(true);

            const actualBlockNumber = await defaultCheckPointers.getBlockNumber();
            expect(actualBlockNumber).toBeUndefined();
            const actualTxnIDS = await defaultCheckPointers.getTransactionIds();
            expect(actualTxnIDS.length).toEqual(0);
        });

        it('Checkpointer file exist with checkpointer state initialized', async () => {
            const checkPointerState : CheckPointerState = {
                    blockNumber : undefined,
                    transactionIDs: []
            }

            await writeToFile(checkPointerState,checkPointerFilePath);
            const actualData = await readFromFile(checkPointerFilePath);
            expect(actualData).toEqual(
                checkPointerState
            )

            const defaultCheckPointers = await DefaultCheckPointers.file(checkPointerFilePath);
            expect(fs.existsSync(checkPointerFilePath)).toEqual(true);

            const actualBlockNumber = await defaultCheckPointers.getBlockNumber();
            expect(actualBlockNumber).toBeUndefined();

            const actualTxnIDS :string[] = await defaultCheckPointers.getTransactionIds() ;
            expect(actualTxnIDS.length).toEqual(0);
        });

        it('Checkpointing only a block number in a fresh checkpointer gives block number & no transactions', async () => {

            const expectedBlockNumber = BigInt(101);

            const defaultCheckPointers = await DefaultCheckPointers.file(checkPointerFilePath);
            await defaultCheckPointers.checkpoint(expectedBlockNumber);
            const actualBlockNumber = await defaultCheckPointers.getBlockNumber();
            expect(actualBlockNumber).toEqual(expectedBlockNumber);

            const actualTxnIDS = await defaultCheckPointers.getTransactionIds();
            expect(actualTxnIDS.length).toEqual(0);
        });

        it('Checkpointing same block number and new transaction in used checkpointer gives block number and expected transactions', async () => {
            const expectedBlockNumber = BigInt(101);
            const expectedTxnIDs = ['txn1'];

            const defaultCheckPointers = await DefaultCheckPointers.file(checkPointerFilePath);
            await defaultCheckPointers.checkpoint(expectedBlockNumber)
            const actualBlockNumber = await defaultCheckPointers.getBlockNumber();
            expect(actualBlockNumber).toEqual(expectedBlockNumber);

            await defaultCheckPointers.checkpoint(expectedBlockNumber,'txn1');

            const actualTxnIDS = await defaultCheckPointers.getTransactionIds();
            expect(actualTxnIDS).toStrictEqual(expectedTxnIDs);
        });

        it('Checkpointing block and transaction in a fresh checkpointer, gives block number and transaction', async () => {

            const expectedBlockNumber = BigInt(101);
            const expectedTxnIDs = ['txn1'];

            const defaultCheckPointers = await DefaultCheckPointers.file(checkPointerFilePath);
            await defaultCheckPointers.checkpoint(expectedBlockNumber,'txn1')

            const actualBlockNumber = await defaultCheckPointers.getBlockNumber();
            expect(actualBlockNumber).toEqual(expectedBlockNumber);

            const actualTxnIDS = await defaultCheckPointers.getTransactionIds();
            expect(actualTxnIDS).toStrictEqual(expectedTxnIDs);
        });

        it('Checkpointing only a new block in used checkpointer gives new block number and no transactions', async () => {

            const blockNumber1 = BigInt(101);

            const blockNumber2 = BigInt(102);

            const expectedTxnIDs: any[] = [];

            const defaultCheckPointers = await DefaultCheckPointers.file(checkPointerFilePath);
            await defaultCheckPointers.checkpoint(blockNumber1,'txn1');

            let actualBlockNumber = await defaultCheckPointers.getBlockNumber();
            expect(actualBlockNumber).toStrictEqual(blockNumber1);

            await defaultCheckPointers.checkpoint(blockNumber2);
            actualBlockNumber = await defaultCheckPointers.getBlockNumber();
            expect(actualBlockNumber).toStrictEqual(blockNumber2);
            const actualTxnIDS = await defaultCheckPointers.getTransactionIds();
            expect(actualTxnIDS).toStrictEqual(expectedTxnIDs);

        });

            it('Checkpointing new block and transaction in used checkpointer gives new block and only new transaction', async () => {

                const blockNumber1 = BigInt(101);
                const blockNumber2 = BigInt(102);

                const expectedTxnIDs: string[] = ['txn2','txn3'];

                const defaultCheckPointers = await DefaultCheckPointers.file(checkPointerFilePath);
                await defaultCheckPointers.checkpoint(blockNumber1,'txn1');

                let actualBlockNumber = await defaultCheckPointers.getBlockNumber();
                expect(actualBlockNumber).toStrictEqual(blockNumber1);

                await defaultCheckPointers.checkpoint(blockNumber2);
                actualBlockNumber = await defaultCheckPointers.getBlockNumber();

                await defaultCheckPointers.checkpoint(blockNumber2,'txn2');
                await defaultCheckPointers.checkpoint(blockNumber2,'txn3');

                expect(actualBlockNumber).toStrictEqual(blockNumber2);
                const actualTxnIDS = await defaultCheckPointers.getTransactionIds();
                expect(actualTxnIDS).toStrictEqual(expectedTxnIDs);

            });
    })
