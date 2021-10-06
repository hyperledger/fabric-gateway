/*
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Contract } = require('fabric-contract-api');

class ErrorsContract extends Contract {

    async exists(ctx, id) {
        const buffer = await ctx.stub.getState(id);
        return (!!buffer && buffer.length > 0);
    }

    async crash(ctx) {
        console.log('CRASHING........');
        process.exit(99);
    }

    async nondet(ctx) {
        const rd = Math.random() * 100000;
        console.log('Running nondet', rd);

        await ctx.stub.putState('random', Buffer.from(rd.toString()));
    }

    async orgsFail(ctx, orgsJSON) {
        let orgs = JSON.parse(orgsJSON);
        const peerOrg = await ctx.stub.getMspID();
        if (orgs.includes(peerOrg)) {
            throw new Error(peerOrg + ' refuses to endorse this');
        }
        await ctx.stub.putState('mykey', 'myvalue');
    }

    async longRunning(ctx, repeat) {
        for (let i = 0; i < repeat; i++) {
            const res = await ctx.stub.getStateByRange(null, null);
            console.log(res);
        }
    }
}

module.exports = ErrorsContract;
