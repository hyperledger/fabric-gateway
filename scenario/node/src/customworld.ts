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
import * as grpc from '@grpc/grpc-js';

interface ConnectionInfo {
    readonly url: string;
    readonly serverNameOverride: string;
    readonly tlsRootCertPath: string;
    running : boolean;
}

const peerConnectionInfo: { [peer: string]: ConnectionInfo } = {
    "peer0.org1.example.com": {
        url:                "localhost:7051",
        serverNameOverride: "peer0.org1.example.com",
        tlsRootCertPath:    fixturesDir + "/crypto-material/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt",
        running:            true,
    },
    "peer1.org1.example.com": {
        url:                "localhost:9051",
        serverNameOverride: "peer1.org1.example.com",
        tlsRootCertPath:    fixturesDir + "/crypto-material/crypto-config/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/ca.crt",
        running:            true,
    },
    "peer0.org2.example.com": {
        url:                "localhost:8051",
        serverNameOverride: "peer0.org2.example.com",
        tlsRootCertPath:    fixturesDir + "/crypto-material/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt",
        running:            true,
    },
    "peer1.org2.example.com": {
        url:                "localhost:10051",
        serverNameOverride: "peer1.org2.example.com",
        tlsRootCertPath:    fixturesDir + "/crypto-material/crypto-config/peerOrganizations/org2.example.com/peers/peer1.org2.example.com/tls/ca.crt",
        running:            true,
    },
    "peer0.org3.example.com": {
        url:                "localhost:11051",
        serverNameOverride: "peer0.org3.example.com",
        tlsRootCertPath:    fixturesDir + "/crypto-material/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt",
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
    return await fs.promises.readFile(certPath);
}

async function newSigner(user: string, mspId: string): Promise<Signer> {
    const privateKey = await readPrivateKey(user, mspId);
    return signers.newPrivateKeySigner(privateKey);
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


function assertDefined<T>(value: T | null | undefined, property: string): T {
    if (null == value) {
        throw new Error(`Bad step sequence: ${property} not defined`);
    }
    return value;
}

interface GatewayContext {
    gateway?: Gateway;
    identity?: Identity;
    signer?: Signer;
    network?: Network;
    contract?: Contract;
    transaction?: TransactionInvocation;
}

export class CustomWorld {
    private gateways: { [name: string]: GatewayContext } = {};
    private currentGateway: GatewayContext = {};

    async createGateway(name: string, user: string, mspId: string): Promise<void> {
        const gateway: GatewayContext = {};
        gateway.identity = await newIdentity(user, mspId);
        gateway.signer = await newSigner(user, mspId);
        this.gateways[name] = gateway;
        this.currentGateway = gateway;
    }

    async createGatewayWithoutSigner(name: string, user: string, mspId: string): Promise<void> {
        const gateway: GatewayContext = {};
        gateway.identity = await newIdentity(user, mspId);
        this.gateways[name] = gateway;
        this.currentGateway = gateway;
    }

    useGateway(name: string): void {
        this.currentGateway = this.gateways[name];
    }

    useNetwork(channelName: string): void {
        this.currentGateway.network = this.getGateway().getNetwork(channelName);
    }

    useContract(contractName: string): void {
        this.currentGateway.contract = this.getNetwork().getContract(contractName);
    }

    async connect(address: string): Promise<void> {
        // address is the name of the peer, lookup the connection info
        const peer = peerConnectionInfo[address];
        const tlsRootCert = fs.readFileSync(peer.tlsRootCertPath)
        const GrpcClient = grpc.makeGenericClientConstructor({}, '');
        const credentials = grpc.credentials.createSsl(tlsRootCert);
        let grpcOptions: Record<string, unknown> = {};
        if (peer.serverNameOverride) {
            grpcOptions = {
                'grpc.ssl_target_name_override': peer.serverNameOverride
            };
        }
        const client = new GrpcClient(peer.url, credentials, grpcOptions);
        const options: ConnectOptions = {
            signer: this.currentGateway.signer,
            identity: this.getIdentity(),
            client,
        };
        this.currentGateway.gateway = await connect(options);
    }

    prepareTransaction(action: string, transactionName: string): void {
        this.currentGateway.transaction = new TransactionInvocation(action, this.getNetwork(), this.getContract(), transactionName);
    }

    setArguments(jsonArgs: string): void {
        const args = JSON.parse(jsonArgs);
        this.getTransaction().options.arguments = args;
    }

    setTransientData(dataTable: DataTable): void {
        const hash = dataTable.rowsHash();
        const transient: { [key: string]: Buffer } = {};
        for (const key in hash) {
            transient[key] = Buffer.from(hash[key]);
        }
        this.getTransaction().options.transientData = transient;
    }

    setEndorsingOrgs(jsonOrgs: string): void {
        const orgs = JSON.parse(jsonOrgs);
        this.getTransaction().options.endorsingOrganizations = orgs;
    }

    close(): void {
        this.currentGateway.gateway?.close();
        delete this.currentGateway.gateway;
    }

    async setOfflineSigner(user: string, mspId: string): Promise<void> {
        const signer = await newSigner(user, mspId);
        this.getTransaction().setOfflineSigner(signer);
    }

    async invokeTransaction(): Promise<void> {
        await this.getTransaction().invokeTransaction();
        this.getTransaction().getResult();
    }

    async assertTransactionFails(): Promise<void> {
        await this.getTransaction().invokeTransaction();
        this.getTransaction().getError();
    }

    getResult(): string {
        return this.getTransaction().getResult();
    }

    private getGateway(): Gateway {
        return assertDefined(this.currentGateway.gateway, 'gateway');
    }

    private getNetwork(): Network {
        return assertDefined(this.currentGateway.network, 'network');
    }

    private getContract(): Contract {
        return assertDefined(this.currentGateway.contract, 'contract');
    }

    private getTransaction(): TransactionInvocation {
        return assertDefined(this.currentGateway.transaction, 'transaction');
    }

    private getIdentity(): Identity {
        return assertDefined(this.currentGateway.identity, 'identity');
    }
}

setWorldConstructor(CustomWorld);
