/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { After, AfterAll, BeforeAll, DataTable, Given, setDefaultTimeout, Then, When } from '@cucumber/cucumber';
import expect from 'expect';
import { ErrorDetail, GatewayError } from 'fabric-gateway';
import { CustomWorld } from './customworld';
import { Fabric } from './fabric';
import { bytesAsString, toError } from './utils';

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

const DEFAULT_LISTENER_NAME = '';

function parseJson(json: string): unknown {
    try {
        return JSON.parse(json);
    } catch (err) {
        const error = toError(err);
        error.message = `${error.message}: ${json}`;
        throw error;
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
    await this.invokeSuccessfulTransaction();
});

When('I listen for chaincode events from {word}', async function(this: CustomWorld, chaincodeName: string): Promise<void> {
    await this.listenForChaincodeEvents(DEFAULT_LISTENER_NAME, chaincodeName);
});

When('I listen for chaincode events from {word} on a listener named {string}', async function(this: CustomWorld, chaincodeName: string, listenerName: string): Promise<void> {
    await this.listenForChaincodeEvents(listenerName, chaincodeName);
});

When('I replay chaincode events from {word} starting at last committed block', async function(this: CustomWorld, chaincodeName: string): Promise<void> {
    await this.replayChaincodeEvents(DEFAULT_LISTENER_NAME, chaincodeName, this.getLastCommittedBlockNumber())
});

When('I stop listening for chaincode events', function(this: CustomWorld): void {
    this.closeChaincodeEvents(DEFAULT_LISTENER_NAME);
});

When('I stop listening for chaincode events on {string}', function(this: CustomWorld, listenerName: string): void {
    this.closeChaincodeEvents(listenerName);
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

Then('the error details should be', function(this: CustomWorld, dataTable: DataTable): void {
    const err = this.getErrorOfType(GatewayError);

    const expectedDetails = new Map<string, ErrorDetail>();
    dataTable.raw().forEach(row => expectedDetails.set(row[0], {
        mspId: row[0],
        address: row[1],
        message: row[2],
    }));

    err.details.forEach(actual => {
        const expected = expectedDetails.get(actual.mspId);
        expect(expected).toBeDefined();
        expect(actual.message).toContain(expected?.message);
        expectedDetails.delete(actual.mspId);
    });
    expect(Object.keys(expectedDetails)).toHaveLength(0);
});

Then('I should receive a chaincode event named {string} with payload {string}', async function(this: CustomWorld, eventName: string, payload: string): Promise<void> {
    const event = await this.nextChaincodeEvent(DEFAULT_LISTENER_NAME);
    const actual = Object.assign({}, event, { payload: bytesAsString(event.payload)})
    expect(actual).toMatchObject({ eventName, payload });
});

Then('I should receive a chaincode event named {string} with payload {string} on {string}', async function(this: CustomWorld, eventName: string, payload: string, listenerName: string): Promise<void> {
    const event = await this.nextChaincodeEvent(listenerName);
    const actual = Object.assign({}, event, { payload: bytesAsString(event.payload)})
    expect(actual).toMatchObject({ eventName, payload });
});
