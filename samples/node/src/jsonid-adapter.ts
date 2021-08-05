/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */


import { Identity, Signer, signers } from 'fabric-gateway';
import fs from 'fs';
import path from 'path';
import crypto from 'crypto';

interface JSONID {
    name: string,
    cert: string,
    ca: string,
    hsm: boolean,
    private_key: string,
    mspId: string
}

/**
 * This class can be used to map identities in a variety of JSON formats to the Identity and Signers required
 * for the gateway
 * 
 *  JSONIDAdapter jsonAdapter = new JsonAdapter("walletdir");
 *  const gateway = connect({
 *       client,
 *       identity: await jsonAdapter.getIdentity("AppAdmin"),
 *       signer: await jsonAdapter.getSigner("AppAdmin")
 *  });
 */
export default class JSONIDAdapter {

    private idFilesDir: string;
    private mspId: string = '';

    /** 
     * @param dFilesDir Directory to load the files from 
     * @param mspId optional MSPID to apply to all identities returned if they are missing it
     */
    public constructor(idFilesDir: string, mspId?: string) {
        this.idFilesDir = path.resolve(idFilesDir);
        if (!fs.existsSync(idFilesDir)) {
            throw new Error(`Can't locate the [${idFilesDir}]`);
        }

        if (mspId){
            this.mspId = mspId;
        }
        
    }

    private readIDFile(idFile: string): JSONID {
        let idJsonFile = path.resolve(path.join(this.idFilesDir, idFile));

        // check if there's no extension probably means it's a waller id file
        if (path.extname(idJsonFile)===""){
            idJsonFile = `${idJsonFile}.id`
        }

        if (!fs.existsSync(idJsonFile)) {
            throw new Error(`Can't locate the id file [${idJsonFile}]`);
        }
        let id: JSONID;
        let json = JSON.parse(fs.readFileSync(idJsonFile, 'utf-8'));

        if (json['credentials']){
            // v2 wallet format
            id = {
                name: idFile,
                cert: json['credentials']['certificate'],
                ca: '',
                hsm: false,
                private_key: Buffer.from(json.private_key, 'base64').toString(),
                mspId : json.mspId
            }
        } else {
            // ibp style format
            id = {
                name: json.name ,
                cert: Buffer.from(json.cert, 'base64').toString(),
                ca: Buffer.from(json.ca, 'base64').toString(),
                hsm: json.js,
                private_key: Buffer.from(json.private_key, 'base64').toString(),
                mspId: this.mspId
            }
        }
        return id;
    }

    public async getIdentity(idFile: string): Promise<Identity> {
        const id = this.readIDFile(idFile);
        
        const identity: Identity = {
            credentials: Buffer.from(id.cert),
            mspId: id.mspId
        };

        return identity;
    }

    public async getSigner(idFile: string): Promise<Signer> {
        const id = this.readIDFile(idFile);
        const privateKeyPem: string = Buffer.from(id.private_key, 'base64').toString();
        const privateKey = crypto.createPrivateKey(privateKeyPem);
        return signers.newPrivateKeySigner(privateKey);
    }

}