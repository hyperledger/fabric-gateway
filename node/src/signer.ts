/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { marshal } from './impl/protoutils'
import { KEYUTIL } from 'jsrsasign';
import * as crypto from 'crypto';
import * as elliptic from 'elliptic';

const EC = elliptic.ec;
const curves: any =  elliptic.curves;
const ecdsaCurve = curves['p256'];
const ecdsa = new EC(ecdsaCurve);

export class Signer {
    private readonly mspid: string;
    private readonly cert: Buffer;
    private readonly signKey: elliptic.ec.KeyPair;
    private readonly serialized: Uint8Array;

    constructor(mspid: string, certPem: Buffer, keyPem: Buffer) {
        this.mspid = mspid;
        this.cert = certPem;
        const key: any = KEYUTIL.getKey(keyPem.toString()); // convert the pem encoded key to hex encoded private key
        this.signKey = ecdsa.keyFromPrivate(key.prvKeyHex, 'hex');
        this.serialized = marshal('msp.SerializedIdentity', {
            mspid: mspid,
            idBytes: certPem
        })
    }

    sign(msg: crypto.BinaryLike) {
        const hash = crypto.createHash('sha256');
        hash.update(msg);
        const digest: any = hash.digest();
        const sig = ecdsa.sign(Buffer.from(digest, 'hex'), this.signKey);
        _preventMalleability(sig, ecdsaCurve);
        return sig.toDER();
    }

    serialize() {
        return this.serialized;
    }
}

function _preventMalleability(sig: elliptic.ec.Signature, curve: any) {

    const halfOrder = curve.n.shrn(1);
    if (!halfOrder) {
        throw new Error('Can not find the half order needed to calculate "s" value for immalleable signatures. Unsupported curve name: ' + curve);
    }

    if (sig.s.cmp(halfOrder) === 1) {
        const bigNum = curve.n;
        sig.s = bigNum.sub(sig.s);
    }

    return sig;
}
