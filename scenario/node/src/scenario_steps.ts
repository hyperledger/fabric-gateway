/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { After, AfterAll, BeforeAll, DataTable, Given, setDefaultTimeout, Then, When } from '@cucumber/cucumber';
import expect from 'expect';
import { CustomWorld } from './customworld';
import { Fabric } from './fabric';
import { asString } from './utils';

setDefaultTimeout(30 * 1000);

const TIMEOUTS = {
    HUGE_TIME: 20 * 60 * 1000,
    LONG_STEP: 240 * 1000,
    MED_STEP: 120 * 1000,
    SHORT_STEP: 60 * 1000,
    LONG_INC: 30 * 1000,
    MED_INC: 10 * 1000,
    SHORT_INC: 5 * 1000
};

function parseJson(json: string): unknown {
    try {
        return JSON.parse(json);
    } catch (err) {
        err.message = `${err.message}: ${json}`;
        throw err;
    }
}

let fabric: Fabric;

BeforeAll(function(): void {
    fabric = new Fabric();
});

AfterAll(function(this: CustomWorld): void {
    fabric.dockerDown();
});

After(function(this: CustomWorld): void {
    this.close();
});

Given('I have deployed a Fabric network', { timeout: TIMEOUTS.LONG_STEP }, async function(this: CustomWorld): Promise<void> {
    await fabric.deployNetwork();
});

Given('I have created and joined all channels', { timeout: TIMEOUTS.LONG_STEP }, async function(this: CustomWorld): Promise<void> {
    await fabric.createChannels();
});

Given(/^I deploy (\w+) chaincode named (\w+) at version ([^ ]+) for all organizations on channel (\w+) with endorsement policy (.+)$/,
    { timeout: TIMEOUTS.LONG_STEP },
    async function(this: CustomWorld, ccType: string, ccName: string, version: string, channelName: string, signaturePolicy: string): Promise<void> {
        await fabric.deployChaincode(ccType, ccName, version, channelName, signaturePolicy);
    });

Given('I register and enroll an HSM user {word} in MSP Org1MSP', async function(this: CustomWorld, user: string): Promise<void> {
    await fabric.generateHSMUser(user);
});

Given('I create a gateway named {word} for user {word} in MSP {word}', async function(this: CustomWorld, name: string, user: string, mspId: string): Promise<void> {
    await this.createGateway(name, user, mspId);
});

Given('I create a gateway named {word} for HSM user {word} in MSP {word}', async function(this: CustomWorld, name: string, user: string, mspId: string): Promise<void> {
    await this.createGatewayWithHSMUser(name, user, mspId);
});

Given('I create a gateway named {word} without signer for user {word} in MSP {word}', async function(this: CustomWorld, name: string, user: string, mspId: string): Promise<void> {
    await this.createGatewayWithoutSigner(name, user, mspId);
});

Given('I use the gateway named {word}', async function(this: CustomWorld, name: string): Promise<void> {
    await this.useGateway(name);
});

Given('I connect the gateway to {word}', async function(this: CustomWorld, address: string): Promise<void> {
    await this.connect(address);
});

When('I use the {word} network', function (this: CustomWorld, channelName: string): void {
    this.useNetwork(channelName);
});

When('I use the {word} contract', function(this: CustomWorld, contractName: string): void {
    this.useContract(contractName);
});

When(/I stop the peer named (.+)/, function(this: CustomWorld, peer: string): void {
    fabric.stopPeer(peer);
});

When(/I start the peer named (.+)/, async function(this: CustomWorld, peer: string): Promise<void> {
    await fabric.startPeer(peer);
});

When(/I prepare to (evaluate|submit) an? ([^ ]+) transaction/, function(this: CustomWorld, action: string, txnName: string): void {
    this.prepareTransaction(action, txnName);
});

When(/I set the transaction arguments? to (.+)/, function(this: CustomWorld, jsonArgs: string): void {
    this.setArguments(jsonArgs);
});

When('I set transient data on the transaction to', function(this: CustomWorld, dataTable: DataTable): void {
    this.setTransientData(dataTable);
});

When(/I set the endorsing organizations? to (.+)/, function(this: CustomWorld, jsonOrgs: string): void {
    this.setEndorsingOrgs(jsonOrgs);
});

When('I do off-line signing as user {word} in MSP {word}', async function(this: CustomWorld, user: string, mspId: string): Promise<void> {
    await this.setOfflineSigner(user, mspId);
})

When('I invoke the transaction', async function(this: CustomWorld): Promise<void> {
    await this.invokeTransaction();
});

When('I listen for chaincode events from {word}', async function(this: CustomWorld, chaincodeId: string): Promise<void> {
    await this.listenForChaincodeEvents(chaincodeId);
});

Then('the transaction invocation should fail', async function(this: CustomWorld): Promise<void> {
    await this.assertTransactionFails();
});

Then('the response should be JSON matching', function(this: CustomWorld, docString: string): void {
    const resultText = this.getResult();
    const actual = parseJson(resultText);
    const expected = parseJson(docString);
    expect(actual).toEqual(expected);
});

Then('the response should be {string}', function(this: CustomWorld, expected: string): void {
    const actual = this.getResult();
    expect(actual).toEqual(expected);
});

Then('the error message should contain {string}', function(this: CustomWorld, expected: string): void {
    const actual = this.getError().message;
    expect(actual).toContain(expected);
});

Then('I should receive a chaincode event named {string} with payload {string}', async function(this: CustomWorld, eventName: string, payload: string): Promise<void> {
    const event = await this.nextChaincodeEvent();
    const actual = Object.assign({}, event, { payload: asString(event.payload)})
    expect(actual).toMatchObject({ eventName, payload });
});
