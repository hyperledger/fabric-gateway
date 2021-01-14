/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

'use strict';

const { Given, When, Then, BeforeAll, AfterAll, setDefaultTimeout } = require('@cucumber/cucumber');
const { execFileSync, spawnSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const { connect, Signers } = require('fabric-gateway');
const chai = require('chai');
const expect = chai.expect;

setDefaultTimeout(30 * 1000);

const fixturesDir = path.resolve(__dirname, '..', '..', 'fixtures');
const dockerComposeDir = path.join(fixturesDir, 'docker-compose');
const dockerComposeFile = 'docker-compose-tls.yaml';

const TIMEOUTS = {
    HUGE_TIME: 20 * 60 * 1000,
    LONG_STEP: 240 * 1000,
    MED_STEP: 120 * 1000,
    SHORT_STEP: 60 * 1000,
    LONG_INC: 30 * 1000,
    MED_INC: 10 * 1000,
    SHORT_INC: 5 * 1000
};

const tlsOptions = [
    '--tls',
    'true',
    '--cafile',
    '/etc/hyperledger/configtx/crypto-config/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem'
];

const mspToOrgMap = {
    "Org1MSP": "org1.example.com"
};

const orgs = [
    {
        cli: "org1_cli",
        anchortx: "/etc/hyperledger/configtx/Org1MSPanchors.tx",
        peers: ["peer0.org1.example.com:7051", "peer1.org1.example.com:9051"]
    },
    {
        cli: "org2_cli",
        anchortx: "/etc/hyperledger/configtx/Org2MSPanchors.tx",
        peers: ["peer0.org2.example.com:8051", "peer1.org2.example.com:10051"]
    },
    {
        cli: "org3_cli",
        anchortx: "/etc/hyperledger/configtx/Org3MSPanchors.tx",
        peers: ["peer0.org3.example.com:11051"]
    }
];

const runningPeers = {
    'peer0.org1.example.com': true,
    'peer1.org1.example.com': true,
    'peer0.org2.example.com': true,
    'peer1.org2.example.com': true,
    'peer0.org3.example.com': true
};

let fabricRunning;
let channelsJoined;
let runningChaincodes;

BeforeAll(function() {
    fabricRunning = false;
    channelsJoined = false;
    runningChaincodes = {};
});

AfterAll(function() {
    const out = spawnSync('docker-compose', ['-f', dockerComposeFile, '-p', 'node', 'down'], { cwd: dockerComposeDir })
    console.log(out.output.toString());
})

Given('I have deployed a {word} Fabric network', { timeout: TIMEOUTS.LONG_STEP }, async (type) => {
    if (!fabricRunning) {
        // generate crypto material
        let out = execFileSync('./generate.sh', { cwd: fixturesDir })
        console.log(out.toString());

        out = spawnSync('docker-compose', ['-f', dockerComposeFile, '-p', 'node', 'up', '-d'], { cwd: dockerComposeDir })
        console.log(out.output.toString());

        fabricRunning = true;
        await sleep(20000);
    }
});

Given('I have created and joined all channels from the tls connection profile', { timeout: TIMEOUTS.LONG_STEP }, async function() {
    await startAllPeers();
    if (!channelsJoined) {
        dockerCommandWithTLS(
            'exec', 'org1_cli', 'peer', 'channel', 'create',
            '-o', 'orderer.example.com:7050',
            '-c', 'mychannel',
            '-f', '/etc/hyperledger/configtx/channel.tx',
            '--outputBlock', '/etc/hyperledger/configtx/mychannel.block');

        orgs.forEach(function(org) {
            org.peers.forEach(peer => {
                const env = 'CORE_PEER_ADDRESS=' + peer;
                dockerCommandWithTLS(
                    'exec', "-e", env, org.cli, 'peer', 'channel', 'join',
                    '-b', '/etc/hyperledger/configtx/mychannel.block');
            });
            dockerCommandWithTLS(
                'exec', org.cli, 'peer', 'channel', 'update',
                '-o', 'orderer.example.com:7050',
                '-c', 'mychannel',
                '-f', org.anchortx);
        });

        channelsJoined = true;
        await sleep(10000);
    }
});

Given(/^I deploy (\w+) chaincode named (\w+) at version ([^ ]+) for all organizations on channel (\w+) with endorsement policy (.+)$/, { timeout: TIMEOUTS.LONG_STEP }, async function(ccType, ccName, version, channelName, signaturePolicy) {
    const mangledName = ccName + version + channelName;
    if (runningChaincodes[mangledName]) {
        return;
    }

    let ccPath = 'github.com/chaincode/' + ccType + '/' + ccName
    if (ccType != "golang") {
        ccPath = "/opt/gopath/src/" + ccPath
    }
    const ccLabel = ccName + 'v' + version
    const ccPackage = ccName + '.tar.gz'

    orgs.forEach(function(org) {
        dockerCommand(
            'exec', org.cli, 'peer', 'lifecycle', 'chaincode', 'package', ccPackage,
            '--lang', ccType,
            '--label', ccLabel,
            '--path', ccPath);

        org.peers.forEach(function(peer) {
            const env = 'CORE_PEER_ADDRESS=' + peer;
            dockerCommand('exec', "-e", env, org.cli, 'peer', 'lifecycle', 'chaincode', 'install', ccPackage);
        });

        const out = dockerCommand('exec', org.cli, 'peer', 'lifecycle', 'chaincode', 'queryinstalled');

        const pattern = new RegExp('.*Package ID: (.*), Label: ' + ccLabel + '.*');
        const match = out.match(pattern);
        if (match === null || match.length < 2) {
            throw 'chaincode not found on org1 peer';
        }
        const packageID = match[1];

        dockerCommandWithTLS(
            'exec', org.cli, 'peer', 'lifecycle', 'chaincode',
            'approveformyorg', '--package-id', packageID, '--channelID', channelName, '--name', ccName,
            '--version', version, '--signature-policy', signaturePolicy,
            '--sequence', '1', '--waitForEvent');
    });

    // commit
    dockerCommandWithTLS(
        'exec', 'org1_cli', 'peer', 'lifecycle', 'chaincode', 'commit',
        '--channelID', channelName,
        '--name', ccName,
        '--version', version,
        '--signature-policy', signaturePolicy,
        '--sequence', '1',
        '--waitForEvent',
        '--peerAddresses', 'peer0.org1.example.com:7051',
        '--peerAddresses', 'peer0.org2.example.com:8051',
        '--tlsRootCertFiles',
        '/etc/hyperledger/configtx/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt',
        '--tlsRootCertFiles',
        '/etc/hyperledger/configtx/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt');

    runningChaincodes[mangledName] = true;
    await sleep(10000);
});

Given('I create a gateway for user {word} in MSP {word}', async function(user, mspId) {
    this.identity = await newIdentity(user, mspId);
    this.signer = await newSigner(user, mspId);
    delete this.gateway;
});

Given('I create a gateway without signer for user {word} in MSP {word}', async function(user, mspId) {
    this.identity = await newIdentity(user, mspId);
    delete this.gateway;
});

Given('I connect the gateway to {word}', async function(address) {
    const options = {
        url: address,
        signer: this.signer,
        identity: this.identity
    }
    this.gateway = await connect(options);
});

Given('I use the {word} network', function (channelName) {
    this.network = this.gateway.getNetwork(channelName);
});

Given('I use the {word} contract', function(contractName) {
    this.contract = this.network.getContract(contractName);
});

When(/I prepare to (evaluate|submit) an? ([^ ]+) transaction/, function(action, txnName) {
    this.txn = {
        name: txnName,
        options: {},
        contract: this.contract,
    };
    const invoke = getInvoke(action);
    this.txn.invoke = () => invoke(this.txn);
});

When(/I set the transaction arguments? to (.+)/, function(jsonArgs) {
    const args = JSON.parse(jsonArgs);
    this.txn.options.arguments = args;
});

When('I set transient data on the transaction to', function(dataTable) {
    const hash = dataTable.rowsHash();
    const transient = {};
    Object.keys(hash).forEach(key => { transient[key] = Buffer.from(hash[key]) });
    this.txn.options.transientData = transient;
});

When('I do off-line signing as user {word} in MSP {word}', async function(user, mspId) {
    this.txn.offlineSigner = await newSigner(user, mspId);
})

When('I invoke the transaction', async function() {
    this.txn.result = await this.txn.invoke();
});

Then('the transaction invocation should fail', async function() {
    try {
        this.txn.result = await this.txn.invoke();
    } catch (err) {
        return;
    }
    throw new Error(`Transaction invocation was expected to fail, but it returned: ${this.txn.result}`);
});

Then('the response should be JSON matching', function(docString) {
    const resultText = new TextDecoder().decode(this.txn.result);
    const actual = parseJson(resultText);
    const expected = parseJson(docString);
    expect(actual).to.eql(expected);
});

Then('the response should be {string}', function(expected) {
    const actual = new TextDecoder().decode(this.txn.result);
    expect(actual).to.equal(expected);
});

When(/I stop the peer named (.+)/, function(peer) {
    dockerCommand('stop', peer);
    runningPeers[peer] = false;
});

When(/I start the peer named (.+)/, async function(peer) {
    dockerCommand('start', peer);
    runningPeers[peer] = true;
    await sleep(20000);
});

function dockerCommand(...args) {
    const result = spawnSync('docker', args);
    const output = result.output.toString()
    console.log(output);
    // check return code
    return output;
}

function dockerCommandWithTLS(...args) {
    const allArgs = args.concat(tlsOptions);
    return dockerCommand(...allArgs);
}

function parseJson(json) {
    try {
        return JSON.parse(json);
    } catch (err) {
        err.message = err.message + ': ' + json;
        throw err;
    }
}

async function startAllPeers() {
    for (const peer in runningPeers) {
        if (!runningPeers[peer]) {
            dockerCommand('start', peer);
            runningPeers[peer] = true;
        }
    }
    await sleep(20000);
}

function getInvoke(action) {
    if ('evaluate' === action) {
        return evaluate;
    } else if ('submit' === action) {
        return submit;
    }
    throw new Error(`Unknown transaction action: ${this.txn.action}`)
}

async function evaluate(tx) {
    let proposal = tx.contract.newProposal(tx.name, tx.options);
    if (tx.offlineSigner) {
        proposal = offlineSign(proposal, tx.offlineSigner, tx.contract.newSignedProposal.bind(tx.contract));
    }

    return await proposal.evaluate();
}

async function submit(tx) {
    let proposal = tx.contract.newProposal(tx.name, tx.options);
    if (tx.offlineSigner) {
        proposal = offlineSign(proposal, tx.offlineSigner, tx.contract.newSignedProposal.bind(tx.contract));
    }
    
    let transaction = await proposal.endorse();
    if (tx.offlineSigner) {
        transaction = offlineSign(transaction, tx.offlineSigner, tx.contract.newSignedTransaction.bind(tx.contract));
    }

    return await transaction.submit();
}

function offlineSign(signable, sign, newInstance) {
    const signature = sign(signable.getDigest());
    return newInstance(signable.getBytes(), signature);
}

async function newIdentity(user, mspId) {
    const certificate = await readCertificate(user, mspId);
    return {
        mspId,
        credentials: certificate
    };
}

async function readCertificate(user, mspId) {
    const org = mspToOrgMap[mspId];
    const credentialsPath = getCredentialsPath(user, mspId);
    const certPath = path.join(credentialsPath, 'signcerts', `${user}@${org}-cert.pem`);
    return await fs.promises.readFile(certPath);
}

async function newSigner(user, mspId) {
    const privateKey = await readPrivateKey(user, mspId);
    return Signers.newECDSAPrivateKeySigner(privateKey);
}

async function readPrivateKey(user, mspId) {
    const credentialsPath = getCredentialsPath(user, mspId);
    const keyPath = path.join(credentialsPath, 'keystore', 'key.pem');
    const privateKeyPem = await fs.promises.readFile(keyPath);
    return crypto.createPrivateKey(privateKeyPem);
}

function getCredentialsPath(user, mspId) {
    const org = mspToOrgMap[mspId];
    return path.join(fixturesDir, 'crypto-material', 'crypto-config', 'peerOrganizations', `${org}`,
        'users', `${user}@${org}`, 'msp');
}

async function sleep(ms) {
    await new Promise(resolve => setTimeout(resolve, ms));
}
