/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

'use strict';

const {Gateway, Signer} = require('./sdk/sdk');
const fs = require('fs');

(async() => {
    try {
        const mspid = "Org1MSP"
		const certPath = "../../scenario/fixtures/crypto-material/crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem"
		const keyPath = "../../scenario/fixtures/crypto-material/crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore/key.pem"
        const cert = fs.readFileSync(certPath);
        const key = fs.readFileSync(keyPath);
        const signer = new Signer(mspid, cert, key);
        const gateway = new Gateway();
        gateway.connect('localhost:1234', signer);
        const network = gateway.getNetwork('mychannel');
        const contract = network.getContract('fabcar');
        let result = await contract.evaluateTransaction('queryAllCars');
        console.log(result);
        await contract.submitTransaction("createCar", "CAR12", "VW", "Polo", "Grey", "Mary");
	    result = await contract.evaluateTransaction("queryCar", "CAR12");
        console.log(result);
	    await contract.submitTransaction("changeCarOwner", "CAR12", "Archie");
    	result = await contract.evaluateTransaction("queryCar", "CAR12");
        console.log(result);
    } catch(err) {
        console.log(err);
    }
})()
