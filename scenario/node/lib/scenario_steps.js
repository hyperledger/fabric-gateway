/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

'use strict';

const { Given, When, Then, BeforeAll, AfterAll, After } = require('cucumber');
const { execFileSync, spawnSync, spawn } = require('child_process');
const fs = require('fs');
const { Gateway, Signer } = require('fabric-gateway');
const chai = require('chai');
const expect = chai.expect;

const fixturesDir = __dirname + '/../../fixtures';
const dockerComposeDir = fixturesDir + '/docker-compose';
const dockerComposeFile = 'docker-compose-tls.yaml';
const gatewayDir = __dirname + '/../../../bin';

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

BeforeAll(() => {
  this.fabricRunning = false;
  this.channelsJoined = false;
  this.runningChaincodes = {};
});

AfterAll(() => {
  if (this.gatewayProcess) {
    process.kill(-this.gatewayProcess.pid);
  }
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

Given('I have a gateway for {word}', (mspid) => {
  if (!this.gatewayProcess) {
    const env = { DISCOVERY_AS_LOCALHOST: 'TRUE' };
    Object.assign(env, process.env);
    this.gatewayProcess = spawn('./gateway', [
      '-h', 'peer0.org1.example.com',
      '-p', '7051',
      '-m', mspid,
      '-cert', '../scenario/fixtures/crypto-material/crypto-config/peerOrganizations/org1.example.com/users/User2@org1.example.com/msp/signcerts/User2@org1.example.com-cert.pem',
      '-key', '../scenario/fixtures/crypto-material/crypto-config/peerOrganizations/org1.example.com/users/User2@org1.example.com/msp/keystore/key.pem',
      '-tlscert', '../scenario/fixtures/crypto-material/crypto-config/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem'
    ], {
      cwd: gatewayDir,
      env: env,
      detached: true
    });
    this.gatewayProcess.stdout.on('data', (data) => {
      console.log(data.toString());
    });

    this.gatewayProcess.stderr.on('data', (data) => {
      console.error(data.toString());
    });

  }
});

Given(/I deploy (\w+) chaincode named (\w+) at version ([^ ]+) for all organizations on channel (\w+) with endorsement policy ([^ ]+) and arguments(.+)/, { timeout: TIMEOUTS.LONG_STEP }, async (ccType, ccName, version, channelName, policyType, argsJSON) => {
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
  if (match.length < 2) {
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

Given('I have a gateway as user {word} using the tls connection profile', (user) => {
	if(!this.gateway) {
		const mspid = "Org1MSP";
		const certPath = fixturesDir + "/crypto-material/crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem"
		const keyPath = fixturesDir + "/crypto-material/crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore/key.pem"
    const cert = fs.readFileSync(certPath);
    const key = fs.readFileSync(keyPath);

    const signer = new Signer(mspid, cert, key);
    this.gateway = new Gateway();
    this.gateway.connect('localhost:1234', signer);
}
});

Given('I connect the gateway', () => {
  // no op
});

Given('I use the {word} network', (channelName) => {
  this.network = this.gateway.getNetwork(channelName);
});

Given('I use the {word} contract', (contractName) => {
  this.contract = this.network.getContract(contractName);
});

When(/I prepare to (evaluate|submit) an? ([^ ]+) transaction/, (action, txnName) => {
  this.txn = this.contract.createTransaction(txnName);
  if(action === 'evaluate') {
    this.txn = this.contract.prepareToEvaluate(txnName);
  } else {
    this.txn = this.contract.prepareToSubmit(txnName);
  }
});

When(/I set the transaction arguments? to (.+)/, (jsonArgs) => {
  const args = JSON.parse(jsonArgs);
  this.txn.setArgs(...args);
});

When('I set transient data on the transaction to', (dataTable) => {
  const hash = dataTable.rowsHash();
  const transient = {};
  Object.keys(hash).map(key => { transient[key] = Buffer.from(hash[key]) });
  this.txn.setTransient(transient);
});

When('I invoke the transaction', async () => {
  this.txnResult = await this.txn.invoke();
});

Then('the response should be JSON matching', (docString) => {
  const response = JSON.parse(this.txnResult);
  const expected = JSON.parse(docString);
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

