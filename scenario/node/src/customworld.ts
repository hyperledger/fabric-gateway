/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { DataTable, setWorldConstructor } from '@cucumber/cucumber';
import * as grpc from '@grpc/grpc-js';
import { ChaincodeEvent, HSMSigner, HSMSignerFactory, HSMSignerOptions, Identity, Signer, signers } from '@hyperledger/fabric-gateway';
import * as crypto from 'crypto';
import { promises as fs } from 'fs';
import * as path from 'path';
import { findSoftHSMPKCS11Lib, fixturesDir, getOrgForMsp } from './fabric';
import { getSKIFromCertificate } from './fabricski';
import { GatewayContext } from './gatewaycontext';
import { TransactionInvocation } from './transactioninvocation';
import { assertDefined, Constructor, isInstanceOf } from './utils';

let hsmSignerFactory: HSMSignerFactory;

interface ConnectionInfo {
    readonly url: string;
    readonly serverNameOverride: string;
    readonly tlsRootCertPath: string;
    running: boolean;
}

const peerConnectionInfo: Record<string, ConnectionInfo> = {
    'peer0.org1.example.com': {
        url:                'localhost:7051',
        serverNameOverride: 'peer0.org1.example.com',
        tlsRootCertPath:    fixturesDir + '/crypto-material/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt',
        running:            true,
    },
    'peer1.org1.example.com': {
        url:                'localhost:9051',
        serverNameOverride: 'peer1.org1.example.com',
        tlsRootCertPath:    fixturesDir + '/crypto-material/crypto-config/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/ca.crt',
        running:            true,
    },
    'peer0.org2.example.com': {
        url:                'localhost:8051',
        serverNameOverride: 'peer0.org2.example.com',
        tlsRootCertPath:    fixturesDir + '/crypto-material/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt',
        running:            true,
    },
    'peer1.org2.example.com': {
        url:                'localhost:10051',
        serverNameOverride: 'peer1.org2.example.com',
        tlsRootCertPath:    fixturesDir + '/crypto-material/crypto-config/peerOrganizations/org2.example.com/peers/peer1.org2.example.com/tls/ca.crt',
        running:            true,
    },
    'peer0.org3.example.com': {
        url:                'localhost:11051',
        serverNameOverride: 'peer0.org3.example.com',
        tlsRootCertPath:    fixturesDir + '/crypto-material/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt',
        running:            true,
    }
};

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
    return await fs.readFile(certPath);
}

async function newSigner(user: string, mspId: string): Promise<Signer> {
    const privateKey = await readPrivateKey(user, mspId);
    return signers.newPrivateKeySigner(privateKey);
}

async function readPrivateKey(user: string, mspId: string): Promise<crypto.KeyObject> {
    const credentialsPath = getCredentialsPath(user, mspId);
    const keyPath = path.join(credentialsPath, 'keystore', 'key.pem');
    const privateKeyPem = await fs.readFile(keyPath);
    return crypto.createPrivateKey(privateKeyPem);
}

function getCredentialsPath(user: string, mspId: string): string {
    const org = getOrgForMsp(mspId);
    return path.join(fixturesDir, 'crypto-material', 'crypto-config', 'peerOrganizations', `${org}`,
        'users', `${user}@${org}`, 'msp');
}

async function newHSMIdentity(user: string, mspId: string): Promise<Identity> {
    const certificate = await readHSMCertificate(user);
    return {
        mspId,
        credentials: certificate
    };
}

async function newHSMSigner(user: string): Promise<HSMSigner> {
    if (!hsmSignerFactory) {
        hsmSignerFactory = signers.newHSMSignerFactory(findSoftHSMPKCS11Lib());
    }

    const certificate = await readHSMCertificate(user);
    const ski = getSKIFromCertificate(certificate.toString());
    const hsmConfigOptions: HSMSignerOptions = {
        label: 'ForFabric',
        pin: '98765432',
        identifier: ski
    };
    return hsmSignerFactory.newSigner(hsmConfigOptions);
}

async function readHSMCertificate(user: string): Promise<Buffer> {
    const certPath = path.join(fixturesDir, 'crypto-material', 'hsm', user, 'signcerts', 'cert.pem');
    return await fs.readFile(certPath);
}

export class CustomWorld {
    #gateways: Record<string, GatewayContext> = {};
    #currentGateway?: GatewayContext;
    #transaction?: TransactionInvocation;
    #lastCommittedBlockNumber = BigInt(0);

    async createGateway(name: string, user: string, mspId: string): Promise<void> {
        const identity = await newIdentity(user, mspId);
        const signer = await newSigner(user, mspId);
        const gateway = new GatewayContext(identity, signer);
        this.#gateways[name] = gateway;
        this.#currentGateway = gateway;
    }

    createCheckpointer(): void {
        this.getCurrentGateway().createCheckpointer();
    }

    async createGatewayWithoutSigner(name: string, user: string, mspId: string): Promise<void> {
        const identity = await newIdentity(user, mspId);
        const gateway = new GatewayContext(identity);
        this.#gateways[name] = gateway;
        this.#currentGateway = gateway;
    }

    async createGatewayWithHSMUser(name: string, user: string, mspId: string): Promise<void> {
        const identity = await newHSMIdentity(user, mspId);
        const {signer, close} = await newHSMSigner(user);
        const gateway = new GatewayContext(identity, signer, close);
        this.#gateways[name] = gateway;
        this.#currentGateway = gateway;
    }

    useGateway(name: string): void {
        this.#currentGateway = this.#gateways[name];
    }

    useNetwork(channelName: string): void {
        this.getCurrentGateway().useNetwork(channelName);
    }

    useContract(contractName: string): void {
        this.getCurrentGateway().useContract(contractName);
    }

    async connect(address: string): Promise<void> {
        // address is the name of the peer, lookup the connection info
        const peer = peerConnectionInfo[address];
        const tlsRootCert = await fs.readFile(peer.tlsRootCertPath);
        const credentials = grpc.credentials.createSsl(tlsRootCert);
        let grpcOptions: Record<string, unknown> = {};
        if (peer.serverNameOverride) {
            grpcOptions = {
                'grpc.ssl_target_name_override': peer.serverNameOverride
            };
        }
        const client = new grpc.Client(peer.url, credentials, grpcOptions);
        this.getCurrentGateway().connect(client);
    }

    prepareTransaction(action: string, transactionName: string): void {
        this.#transaction = this.getCurrentGateway().newTransaction(action, transactionName);
    }

    setArguments(jsonArgs: string): void {
        const args = JSON.parse(jsonArgs) as string[];
        this.getTransaction().options.arguments = args;
    }

    setTransientData(dataTable: DataTable): void {
        this.getTransaction().options.transientData = dataTable.rowsHash();
    }

    setEndorsingOrgs(jsonOrgs: string): void {
        const orgs = JSON.parse(jsonOrgs) as string[];
        this.getTransaction().options.endorsingOrganizations = orgs;
    }

    async listenForChaincodeEvents(listenerName: string, chaincodeName: string): Promise<void> {
        await this.getCurrentGateway().listenForChaincodeEvents(listenerName, chaincodeName);
    }

    async listenForChaincodeEventsUsingCheckpointer(listenerName: string, chaincodeName: string): Promise<void> {
        await this.getCurrentGateway().listenForChaincodeEventsUsingCheckpointer(listenerName, chaincodeName, { checkpoint: this.getCurrentGateway().getCheckpointer() });
    }

    async replayChaincodeEvents(listenerName: string, chaincodeName: string, startBlock: bigint): Promise<void> {
        await this.getCurrentGateway().listenForChaincodeEvents(listenerName, chaincodeName, { startBlock });
    }

    async nextChaincodeEvent(listenerName: string): Promise<ChaincodeEvent> {
        return await this.getCurrentGateway().nextChaincodeEvent(listenerName);
    }


    async listenForBlockEvents(listenerName: string): Promise<void> {
        await this.getCurrentGateway().listenForBlockEvents(listenerName);
    }

    async listenForBlockEventsUsingCheckpointer(listenerName: string): Promise<void> {
        await this.getCurrentGateway().listenForBlockEventsUsingCheckpointer(listenerName,  { checkpoint: this.getCurrentGateway().getCheckpointer() });
    }

    async listenForFilteredBlockEventsUsingCheckpointer(listenerName: string): Promise<void> {
        await this.getCurrentGateway().listenForFilteredBlockEventsUsingCheckpointer(listenerName,  { checkpoint: this.getCurrentGateway().getCheckpointer() });
    }

    async listenForBlockAndPrivateDataEventsUsingCheckpointer(listenerName: string): Promise<void> {
        await this.getCurrentGateway().listenForBlockAndPrivateDataEventsUsingCheckpointer(listenerName, { checkpoint: this.getCurrentGateway().getCheckpointer() });
    }

    async replayBlockEvents(listenerName: string, startBlock: bigint): Promise<void> {
        await this.getCurrentGateway().listenForBlockEvents(listenerName, { startBlock });
    }

    async nextBlockEvent(listenerName: string): Promise<unknown> {
        return await this.getCurrentGateway().nextBlockEvent(listenerName);
    }

    async listenForFilteredBlockEvents(listenerName: string): Promise<void> {
        await this.getCurrentGateway().listenForFilteredBlockEvents(listenerName);
    }

    async replayFilteredBlockEvents(listenerName: string, startBlock: bigint): Promise<void> {
        await this.getCurrentGateway().listenForFilteredBlockEvents(listenerName, { startBlock });
    }

    async nextFilteredBlockEvent(listenerName: string): Promise<unknown> {
        return await this.getCurrentGateway().nextFilteredBlockEvent(listenerName);
    }

    async listenForBlockAndPrivateDataEvents(listenerName: string): Promise<void> {
        await this.getCurrentGateway().listenForBlockAndPrivateDataEvents(listenerName);
    }

    async replayBlockAndPrivateDataEvents(listenerName: string, startBlock: bigint): Promise<void> {
        await this.getCurrentGateway().listenForBlockAndPrivateDataEvents(listenerName, { startBlock });
    }

    async nextBlockAndPrivateDataEvent(listenerName: string): Promise<unknown> {
        return await this.getCurrentGateway().nextBlockAndPrivateDataEvent(listenerName);
    }

    async setOfflineSigner(user: string, mspId: string): Promise<void> {
        const signer = await newSigner(user, mspId);
        this.getTransaction().setOfflineSigner(signer);
    }

    async invokeSuccessfulTransaction(): Promise<void> {
        await this.invokeTransaction();
        this.getTransaction().getResult();
    }

    private async invokeTransaction(): Promise<void> {
        const transaction = this.getTransaction();
        await transaction.invokeTransaction();
        this.#lastCommittedBlockNumber = transaction.getBlockNumber();
    }

    async assertTransactionFails(): Promise<void> {
        await this.invokeTransaction();
        this.getError();
    }

    getResult(): string {
        return this.getTransaction().getResult();
    }

    getError(): Error {
        return this.getTransaction().getError();
    }

    getErrorOfType<T extends Error>(type: Constructor<T>): T {
        const err = this.getTransaction().getError();

        if (!isInstanceOf(err, type)) {
            throw new TypeError(`Error is not a ${String(type)}: ${String(err)}`);
        }

        return err;
    }

    getLastCommittedBlockNumber(): bigint {
        return this.#lastCommittedBlockNumber;
    }

    close(): void {
        for (const context of Object.values(this.#gateways)) {
            context.close();
        }

        this.#gateways = {};
        this.#currentGateway = undefined;
    }

    closeChaincodeEvents(listenerName: string): void {
        this.getCurrentGateway().closeChaincodeEvents(listenerName);
    }

    closeBlockEvents(listenerName: string): void {
        this.getCurrentGateway().closeBlockEvents(listenerName);
    }

    closeFilteredBlockEvents(listenerName: string): void {
        this.getCurrentGateway().closeFilteredBlockEvents(listenerName);
    }

    closeBlockAndPrivateDataEvents(listenerName: string): void {
        this.getCurrentGateway().closeBlockAndPrivateDataEvents(listenerName);
    }
    private getCurrentGateway(): GatewayContext {
        return assertDefined(this.#currentGateway, 'currentGateway');
    }

    private getTransaction(): TransactionInvocation {
        return assertDefined(this.#transaction, 'transaction');
    }
}

setWorldConstructor(CustomWorld);
