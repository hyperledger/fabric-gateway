/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { DataTable, setWorldConstructor } from '@cucumber/cucumber';
import * as crypto from 'crypto';
import { connect, ConnectOptions, Contract, Gateway, Identity, Network, Signer, signers } from 'fabric-gateway';
import * as fs from 'fs';
import * as path from 'path';
import { fixturesDir, getOrgForMsp } from './fabric';
import { TransactionInvocation } from './transactioninvocation';

async function newIdentity(user: string, mspId: string): Promise<Identity> {
    const certificate = await readCertificate(user, mspId);
    return {
        mspId,
        credentials: certificate
    };
}

async function readCertificate(user: string, mspId: string): Promise<Buffer> {
    const org = getOrgForMsp(mspId);
    const credentialsPath = getCredentialsPath(user, mspId);
    const certPath = path.join(credentialsPath, 'signcerts', `${user}@${org}-cert.pem`);
    return await fs.promises.readFile(certPath);
}

async function newSigner(user: string, mspId: string): Promise<Signer> {
    const privateKey = await readPrivateKey(user, mspId);
    return signers.newECDSAPrivateKeySigner(privateKey);
}

async function readPrivateKey(user: string, mspId: string): Promise<crypto.KeyObject> {
    const credentialsPath = getCredentialsPath(user, mspId);
    const keyPath = path.join(credentialsPath, 'keystore', 'key.pem');
    const privateKeyPem = await fs.promises.readFile(keyPath);
    return crypto.createPrivateKey(privateKeyPem);
}

function getCredentialsPath(user: string, mspId: string): string {
    const org = getOrgForMsp(mspId);
    return path.join(fixturesDir, 'crypto-material', 'crypto-config', 'peerOrganizations', `${org}`,
        'users', `${user}@${org}`, 'msp');
}

export class CustomWorld {
    private identity?: Identity;
    private signer?: Signer;
    private gateway?: Gateway;
    private network?: Network;
    private contract?: Contract;
    private transaction?: TransactionInvocation;

    async createGateway(user: string, mspId: string): Promise<void> {
        this.identity = await newIdentity(user, mspId);
        this.signer = await newSigner(user, mspId);
        delete this.gateway;
    }

    async createGatewayWithoutSigner(user: string, mspId: string): Promise<void> {
        this.identity = await newIdentity(user, mspId);
        delete this.gateway;
    }

    useNetwork(channelName: string): void {
        this.network = this.gateway!.getNetwork(channelName);
    }

    useContract(contractName: string): void {
        this.contract = this.network!.getContract(contractName);
    }

    async connect(address: string): Promise<void> {
        const options: ConnectOptions = {
            url: address,
            signer: this.signer,
            identity: this.identity!
        }
        this.gateway = await connect(options);
    }

    prepareTransaction(action: string, transactionName: string): void {
        this.transaction = new TransactionInvocation(action, this.contract!, transactionName);
    }

    setArguments(jsonArgs: string): void {
        const args = JSON.parse(jsonArgs);
        this.transaction!.options.arguments = args;
    }

    setTransientData(dataTable: DataTable) {
        const hash = dataTable.rowsHash();
        const transient: { [key: string]: Buffer } = {};
        for (const key in hash) {
            transient[key] = Buffer.from(hash[key]);
        }
        this.transaction!.options.transientData = transient;
    }

    close() {
        this.gateway?.close();
        delete this.gateway;
    }

    async setOfflineSigner(user: string, mspId: string): Promise<void> {
        const signer = await newSigner(user, mspId);
        this.transaction!.setOfflineSigner(signer);
    }

    async invokeTransaction(): Promise<void> {
        await this.transaction!.invokeTransaction();
        this.transaction!.getResult();
    }

    async assertTransactionFails(): Promise<void> {
        await this.transaction!.invokeTransaction();
        this.transaction!.getError();
    }

    getResult(): string {
        return this.transaction!.getResult();
    }
}

setWorldConstructor(CustomWorld);
