/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
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

BeforeAll(() => {
  this.fabricRunning = false;
  this.channelsJoined = false;
  this.runningChaincodes = {};
});

AfterAll(() => {
  const out = spawnSync('docker-compose', ['-f', dockerComposeFile, '-p', 'node', 'down'], { cwd: dockerComposeDir })
  console.log(out.output.toString());
})

Given('I have deployed a {word} Fabric network', { timeout: TIMEOUTS.LONG_STEP }, async (type) => {
  if (!this.fabricRunning) {
    // generate crypto material
    let out = execFileSync('./generate.sh', { cwd: fixturesDir })
    console.log(out.toString());

    out = spawnSync('docker-compose', ['-f', dockerComposeFile, '-p', 'node', 'up', '-d'], { cwd: dockerComposeDir })
    console.log(out.output.toString());

    this.fabricRunning = true;
    await new Promise(r => setTimeout(r, 20000));
  }
});

Given('I have created and joined all channels from the tls connection profile', { timeout: TIMEOUTS.LONG_STEP }, async () => {
  if (!this.channelsJoined) {
    dockerCommandWithTLS(
      'exec', 'org1_cli', 'peer', 'channel', 'create',
      '-o', 'orderer.example.com:7050',
      '-c', 'mychannel',
      '-f', '/etc/hyperledger/configtx/channel.tx',
      '--outputBlock', '/etc/hyperledger/configtx/mychannel.block');

    dockerCommandWithTLS(
      'exec', 'org1_cli', 'peer', 'channel', 'join',
      '-b', '/etc/hyperledger/configtx/mychannel.block');

    dockerCommandWithTLS(
      'exec', 'org2_cli', 'peer', 'channel', 'join',
      '-b', '/etc/hyperledger/configtx/mychannel.block');

    dockerCommandWithTLS(
      'exec', 'org1_cli', 'peer', 'channel', 'update',
      '-o', 'orderer.example.com:7050',
      '-c', 'mychannel',
      '-f', '/etc/hyperledger/configtx/Org1MSPanchors.tx');

    dockerCommandWithTLS(
      'exec', 'org2_cli', 'peer', 'channel', 'update',
      '-o', 'orderer.example.com:7050',
      '-c', 'mychannel',
      '-f', '/etc/hyperledger/configtx/Org2MSPanchors.tx');

    this.channelsJoined = true;
    await new Promise(r => setTimeout(r, 10000));
  }
});

Given(/I deploy (\w+) chaincode named (\w+) at version ([^ ]+) for all organizations on channel (\w+) with endorsement policy ([^ ]+)/, { timeout: TIMEOUTS.LONG_STEP }, async (ccType, ccName, version, channelName, policyType) => {
  const mangledName = ccName + version + channelName;
  if (this.runningChaincodes[mangledName]) {
    return;
  }

  const ccPath = '/opt/gopath/src/github.com/chaincode/' + ccType + '/' + ccName
  const ccLabel = ccName + 'v' + version
  const ccPackage = ccName + '.tar.gz'

  // org1
  dockerCommand(
    'exec', 'org1_cli', 'peer', 'lifecycle', 'chaincode', 'package', ccPackage,
    '--lang', ccType,
    '--label', ccLabel,
    '--path', ccPath);

  dockerCommand('exec', 'org1_cli', 'peer', 'lifecycle', 'chaincode', 'install', ccPackage);

  let out = dockerCommand('exec', 'org1_cli', 'peer', 'lifecycle', 'chaincode', 'queryinstalled');

  let pattern = new RegExp('.*Package ID: (.*), Label: ' + ccLabel + '.*');
  let match = out.match(pattern);
  if (match === null || match.length < 2) {
    throw 'chaincode not found on org1 peer';
  }
  let packageID = match[1];

  dockerCommandWithTLS(
    'exec', 'org1_cli', 'peer', 'lifecycle', 'chaincode',
    'approveformyorg', '--package-id', packageID, '--channelID', channelName, '--name', ccName,
    '--version', version, '--signature-policy', `AND('Org1MSP.member','Org2MSP.member')`,
    '--sequence', '1', '--waitForEvent');

  // org2
  dockerCommand(
    'exec', 'org2_cli', 'peer', 'lifecycle', 'chaincode', 'package', ccPackage,
    '--lang', ccType,
    '--label', ccLabel,
    '--path', ccPath);

  dockerCommand('exec', 'org2_cli', 'peer', 'lifecycle', 'chaincode', 'install', ccPackage);

  out = dockerCommand('exec', 'org2_cli', 'peer', 'lifecycle', 'chaincode', 'queryinstalled');

  pattern = new RegExp('.*Package ID: (.*), Label: ' + ccLabel + '.*');
  match = out.match(pattern);
  if (match.length < 2) {
    throw 'chaincode not found on org2 peer';
  }
  packageID = match[1];

  dockerCommandWithTLS(
    'exec', 'org2_cli', 'peer', 'lifecycle', 'chaincode',
    'approveformyorg', '--package-id', packageID, '--channelID', channelName, '--name', ccName,
    '--version', version, '--signature-policy', `AND('Org1MSP.member','Org2MSP.member')`,
    '--sequence', '1', '--waitForEvent');

  // commit
  dockerCommandWithTLS(
    'exec', 'org1_cli', 'peer', 'lifecycle', 'chaincode',
    'commit', '--channelID', channelName, '--name', ccName, '--version', version,
    '--signature-policy', 'AND(\'Org1MSP.member\',\'Org2MSP.member\')', '--sequence', '1',
    '--waitForEvent', '--peerAddresses', 'peer0.org1.example.com:7051', '--peerAddresses',
    'peer0.org2.example.com:8051',
    '--tlsRootCertFiles',
    '/etc/hyperledger/configtx/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt',
    '--tlsRootCertFiles',
    '/etc/hyperledger/configtx/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt');

  this.runningChaincodes[mangledName] = true;
  await new Promise(r => setTimeout(r, 10000));
});

Given('I create a gateway for user {word} in MSP {word}', async (user, mspId) => {
  const org = mspToOrgMap[mspId];
  const credentialsPath = path.join(fixturesDir, 'crypto-material', 'crypto-config', 'peerOrganizations', `${org}`,
    'users', `${user}@${org}`, 'msp');

  const certPath = path.join(credentialsPath, 'signcerts', `${user}@${org}-cert.pem`);
  const certificate = await fs.promises.readFile(certPath);
  this.identity = {
    mspId,
    credentials: certificate
  };

  const keyPath = path.join(credentialsPath, 'keystore', 'key.pem');
  const privateKeyPem = await fs.promises.readFile(keyPath);
  const privateKey = crypto.createPrivateKey(privateKeyPem);
  this.signer = Signers.newECDSAPrivateKeySigner(privateKey);

  delete this.gateway;
});

Given('I connect the gateway to {word}', async (address) => {
  const options = {
    url: address,
    signer: this.signer,
    identity: this.identity
  }
  this.gateway = await connect(options);
});

Given('I use the {word} network', (channelName) => {
  this.network = this.gateway.getNetwork(channelName);
});

Given('I use the {word} contract', (contractName) => {
  this.contract = this.network.getContract(contractName);
});

When(/I prepare to (evaluate|submit) an? ([^ ]+) transaction/, (action, txnName) => {
  this.txn = {
    type: action,
    name: txnName,
    options: {},
  };
});

When(/I set the transaction arguments? to (.+)/, (jsonArgs) => {
  const args = JSON.parse(jsonArgs);
  this.txn.options.arguments = args;
});

When('I set transient data on the transaction to', (dataTable) => {
  const hash = dataTable.rowsHash();
  const transient = {};
  Object.keys(hash).forEach(key => { transient[key] = Buffer.from(hash[key]) });
  this.txn.options.transientData = transient;
});

When('I invoke the transaction', async () => {
  const proposal = this.contract.newProposal(this.txn.name, this.txn.options);
  if (this.txn.type === 'evaluate') {
    this.txn.result = await proposal.evaluate();
  } else if (this.txn.type === 'submit') {
    const transaction = await proposal.endorse();
    this.txn.result = await transaction.submit();
  } else {
    throw new Error(`Unknown transaction type: ${this.txn.type}`);
  }
});

Then('the response should be JSON matching', (docString) => {
  const resultText = new TextDecoder().decode(this.txn.result);
  const response = parseJson(resultText);
  const expected = parseJson(docString);
  expect(response).to.eql(expected);
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
