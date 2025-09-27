/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

import { execFileSync, spawn, spawnSync } from 'node:child_process';
import * as path from 'node:path';
import * as fs from 'node:fs';
import { assertDefined } from './utils';

export const fixturesDir = path.resolve(__dirname, '..', '..', 'fixtures');

const dockerComposeDir = path.join(fixturesDir, 'docker-compose');
const dockerComposeFile = 'docker-compose-tls.yaml';

const tlsOptions = [
    '--tls',
    'true',
    '--cafile',
    '/etc/hyperledger/configtx/crypto-config/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem',
];

interface OrgInfo {
    readonly orgName: string;
    readonly cli: string;
    readonly peers: string[];
}

interface OrdererInfo {
    readonly address: string;
    readonly port: string;
}

const orgs: Record<string, OrgInfo> = {
    Org1MSP: {
        orgName: 'org1.example.com',
        cli: 'org1_cli',
        peers: ['peer0.org1.example.com:7051', 'peer1.org1.example.com:9051'],
    },
    Org2MSP: {
        orgName: 'org2.example.com',
        cli: 'org2_cli',
        peers: ['peer0.org2.example.com:8051', 'peer1.org2.example.com:10051'],
    },
    Org3MSP: {
        orgName: 'org2.example.com',
        cli: 'org3_cli',
        peers: ['peer0.org3.example.com:11051'],
    },
};

const orderers: Array<OrdererInfo> = [
    { address: 'orderer1.example.com', port: '7053' },
    { address: 'orderer2.example.com', port: '8053' },
    { address: 'orderer3.example.com', port: '9053' },
];

export function getOrgForMsp(mspId: string): string {
    const org = assertDefined(orgs[mspId], `no org defined for MSP ID: ${mspId}`);
    return org.orgName;
}

export function findSoftHSMPKCS11Lib(): string {
    const commonSoftHSMPathNames = [
        '/usr/lib/softhsm/libsofthsm2.so',
        '/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so',
        '/usr/local/lib/softhsm/libsofthsm2.so',
        '/usr/lib/libacsp-pkcs11.so',
        '/opt/homebrew/lib/softhsm/libsofthsm2.so',
    ];

    for (const pathnameToTry of commonSoftHSMPathNames) {
        if (fs.existsSync(pathnameToTry)) {
            return pathnameToTry;
        }
    }

    throw new Error('Unable to find PKCS11 library');
}

async function dockerCommand(...args: string[]): Promise<string> {
    const output = await runCommand('docker', ...args);
    console.log(output);
    return output;
}

async function runCommand(command: string, ...args: string[]): Promise<string> {
    const child = spawn(command, args);
    let result = '';

    for await (const chunk of child.stdout) {
        result += String(chunk);
    }

    for await (const chunk of child.stderr) {
        result += String(chunk);
    }

    return new Promise((resolve, reject) => {
        child.on('error', (err) => {
            reject(err);
        });

        child.on('close', (code, signal) => {
            if (code === 0) {
                resolve(result);
                return;
            }

            const commandLine = args.length > 0 ? `${command} ${args.join(' ')}` : command;

            if (signal) {
                reject(new Error(`Command exited with signal ${signal}: ${commandLine}`));
            } else {
                reject(new Error(`Command exited with code ${String(code)}.`));
            }
        });
    });
}

function dockerCommandWithTLS(...args: string[]): Promise<string> {
    const allArgs = args.concat(tlsOptions);
    return dockerCommand(...allArgs);
}

async function sleep(ms: number): Promise<void> {
    await new Promise<void>((resolve) => setTimeout(resolve, ms));
}

interface ChaincodeDefinition {
    package: string;
    type: string;
    label: string;
    path: string;
}

export class Fabric {
    private fabricRunning = false;
    private channelsJoined = false;
    private readonly runningChaincodes: Record<string, string> = {};
    private readonly runningPeers: Record<string, boolean>;

    constructor() {
        this.runningPeers = Object.fromEntries(
            Object.values(orgs)
                .flatMap((org) => org.peers)
                .map((hostPort) => assertDefined(hostPort.split(':')[0], 'hostPort'))
                .map((host) => [host, true]),
        );
    }

    dockerDown(): void {
        const out = spawnSync('docker', ['compose', '-f', dockerComposeFile, '-p', 'node', 'down'], {
            cwd: dockerComposeDir,
        });
        console.log(out.output.toString());
    }

    async deployNetwork(): Promise<void> {
        if (this.fabricRunning) {
            return;
        }

        const generateOut = execFileSync('./generate.sh', { cwd: fixturesDir });
        console.log(generateOut.toString());

        const dockerComposeOut = spawnSync('docker', ['compose', '-f', dockerComposeFile, '-p', 'node', 'up', '-d'], {
            cwd: dockerComposeDir,
        });
        console.log(dockerComposeOut.output.toString());

        this.fabricRunning = true;
        await sleep(20000);
    }

    generateHSMUser(hsmuserid: string): void {
        const generateOut = execFileSync('./generate-hsm-user.sh', [hsmuserid], { cwd: fixturesDir });
        console.log(generateOut.toString());
    }

    async createChannels(): Promise<void> {
        await this.startAllPeers();

        if (this.channelsJoined) {
            return;
        }

        await this.#joinOrderers();
        await this.#joinPeers();
        this.channelsJoined = true;
        await sleep(10000);
    }

    async #joinOrderers(): Promise<string[]> {
        const promises = orderers.map((ord) => {
            const orddir =
                '/etc/hyperledger/configtx/crypto-config/ordererOrganizations/example.com/orderers/' + ord.address;
            return dockerCommand(
                'exec',
                'org1_cli',
                'osnadmin',
                'channel',
                'join',
                '--channelID',
                'mychannel',
                '--config-block',
                '/etc/hyperledger/configtx/mychannel.block',
                '-o',
                ord.address + ':' + ord.port,
                '--ca-file',
                '/etc/hyperledger/configtx/crypto-config/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem',
                '--client-cert',
                orddir + '/tls/server.crt',
                '--client-key',
                orddir + '/tls/server.key',
            );
        });
        return Promise.all(promises);
    }

    async #joinPeers(): Promise<string[]> {
        const peerPromises = Object.values(orgs).flatMap((org) => {
            return org.peers.map((peer) => {
                const env = 'CORE_PEER_ADDRESS=' + peer;
                return dockerCommandWithTLS(
                    'exec',
                    '-e',
                    env,
                    org.cli,
                    'peer',
                    'channel',
                    'join',
                    '-b',
                    '/etc/hyperledger/configtx/mychannel.block',
                );
            });
        });
        return Promise.all(peerPromises);
    }

    async deployChaincode(
        ccType: string,
        ccName: string,
        version: string,
        channelName: string,
        signaturePolicy: string,
    ): Promise<void> {
        let exists = false;
        let sequence = '1';
        const mangledName = ccName + version + channelName;
        const policy = this.runningChaincodes[mangledName];
        if (typeof policy !== 'undefined') {
            if (policy === signaturePolicy) {
                return;
            }
            // Already exists but different signature policy...
            // No need to re-install, just increment the sequence number and approve/commit new signature policy
            exists = true;
            const out = await dockerCommandWithTLS(
                'exec',
                'org1_cli',
                'peer',
                'lifecycle',
                'chaincode',
                'querycommitted',
                '-o',
                'orderer1.example.com:7050',
                '--channelID',
                channelName,
                '--name',
                ccName,
            );
            const pattern = new RegExp('.*Sequence: ([0-9]+),.*');
            const match = out.match(pattern);
            const sequenceNumber = assertDefined(match?.[1], `Chaincode ${ccName} not found on org1 peers`);
            sequence = (Number.parseInt(sequenceNumber) + 1).toString();
        }

        const ccPath = `/opt/gopath/src/github.com/chaincode/${ccType}/${ccName}`;
        const ccLabel = `${ccName}v${version}`;
        const ccPackage = `${ccName}.tar.gz`;

        // is there a collections_config.json file?
        let collectionsConfig: string[] = [];
        const collectionsFile = `chaincode/${ccType}/${ccName}/collections_config.json`;
        if (fs.existsSync(path.join(fixturesDir, collectionsFile))) {
            collectionsConfig = ['--collections-config', path.join('/opt/gopath/src/github.com', collectionsFile)];
        }

        if (!exists) {
            await this.#installChaincode({
                package: ccPackage,
                type: ccType,
                label: ccLabel,
                path: ccPath,
            });
        }

        await this.#approveChaincode(
            ccLabel,
            channelName,
            ccName,
            version,
            signaturePolicy,
            sequence,
            collectionsConfig,
        );

        await this.#commitChaincode(channelName, ccName, version, signaturePolicy, sequence, collectionsConfig);

        this.runningChaincodes[mangledName] = signaturePolicy;
        await sleep(10000);
    }

    async #approveChaincode(
        ccLabel: string,
        channelName: string,
        ccName: string,
        version: string,
        signaturePolicy: string,
        sequence: string,
        collectionsConfig: string[],
    ): Promise<string[]> {
        const promises = Object.values(orgs).map(async (orgInfo) =>
            this.#approveChaincodeForOrg(
                orgInfo,
                ccLabel,
                channelName,
                ccName,
                version,
                signaturePolicy,
                sequence,
                collectionsConfig,
            ),
        );
        return Promise.all(promises);
    }

    async #approveChaincodeForOrg(
        orgInfo: OrgInfo,
        ccLabel: string,
        channelName: string,
        ccName: string,
        version: string,
        signaturePolicy: string,
        sequence: string,
        collectionsConfig: string[],
    ): Promise<string> {
        const out = await dockerCommand('exec', orgInfo.cli, 'peer', 'lifecycle', 'chaincode', 'queryinstalled');

        const pattern = new RegExp('.*Package ID: (.*), Label: ' + ccLabel + '.*');
        const match = out.match(pattern);
        const packageID = assertDefined(match?.[1], `Chaincode ${ccLabel} not found on ${orgInfo.orgName} peers`);

        return dockerCommandWithTLS(
            'exec',
            orgInfo.cli,
            'peer',
            'lifecycle',
            'chaincode',
            'approveformyorg',
            '--package-id',
            packageID,
            '--channelID',
            channelName,
            '--name',
            ccName,
            '--version',
            version,
            '--signature-policy',
            signaturePolicy,
            '--sequence',
            sequence,
            '--waitForEvent',
            ...collectionsConfig,
        );
    }

    async #commitChaincode(
        channelName: string,
        ccName: string,
        version: string,
        signaturePolicy: string,
        sequence: string,
        collectionsConfig: string[],
    ): Promise<string> {
        return dockerCommandWithTLS(
            'exec',
            'org1_cli',
            'peer',
            'lifecycle',
            'chaincode',
            'commit',
            '--channelID',
            channelName,
            '--name',
            ccName,
            '--version',
            version,
            '--signature-policy',
            signaturePolicy,
            '--sequence',
            sequence,
            '--waitForEvent',
            '--peerAddresses',
            'peer0.org1.example.com:7051',
            '--peerAddresses',
            'peer0.org2.example.com:8051',
            '--tlsRootCertFiles',
            '/etc/hyperledger/configtx/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt',
            '--tlsRootCertFiles',
            '/etc/hyperledger/configtx/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt',
            ...collectionsConfig,
        );
    }

    async #installChaincode(chaincode: ChaincodeDefinition): Promise<void> {
        const orgInstalls = Object.values(orgs).map((orgInfo) => this.#installChaincodeToOrg(chaincode, orgInfo));
        await Promise.all(orgInstalls);
    }

    async #installChaincodeToOrg(chaincode: ChaincodeDefinition, orgInfo: OrgInfo): Promise<string[]> {
        await dockerCommand(
            'exec',
            orgInfo.cli,
            'peer',
            'lifecycle',
            'chaincode',
            'package',
            chaincode.package,
            '--lang',
            chaincode.type,
            '--label',
            chaincode.label,
            '--path',
            chaincode.path,
        );

        const peerInstalls = orgInfo.peers.map((peer) => {
            const env = 'CORE_PEER_ADDRESS=' + peer;
            return dockerCommand(
                'exec',
                '-e',
                env,
                orgInfo.cli,
                'peer',
                'lifecycle',
                'chaincode',
                'install',
                chaincode.package,
            );
        });
        return Promise.all(peerInstalls);
    }

    async stopPeer(peer: string): Promise<void> {
        await dockerCommand('stop', peer);
        this.runningPeers[peer] = false;
    }

    async startPeer(peer: string): Promise<void> {
        await dockerCommand('start', peer);
        this.runningPeers[peer] = true;
        await sleep(20000);
    }

    private async startAllPeers(): Promise<void> {
        const stoppedPeers = Object.keys(this.runningPeers).filter((peer) => !this.runningPeers[peer]);
        if (stoppedPeers.length === 0) {
            return;
        }

        const peerStarts = stoppedPeers.map(async (peer) => {
            await dockerCommand('start', peer);
            this.runningPeers[peer] = true;
        });
        await Promise.all(peerStarts);
        await sleep(20000);
    }
}
