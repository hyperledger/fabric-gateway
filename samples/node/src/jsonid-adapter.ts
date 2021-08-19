/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */


import { Identity, Signer, signers } from 'fabric-gateway';
import fs from 'fs/promises';
import path from 'path';
import crypto from 'crypto';

/** Internal interface used to describe all the possible components
 * of the identity
 */
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
 * for the gateway. For example if you have an application wallet, or have exported IDs from SaaS
 * 
 * ```
 *   const jsonAdapter: JSONIDAdapter = new JSONIDAdapter(path.resolve(__dirname,'..','wallet'))
 *
 *   const gateway = connect({
 *       client,
 *       identity: await jsonAdapter.getIdentity("appuser"),
 *       signer: await jsonAdapter.getSigner("appuser"),
 *   });
 *  ```
 * 
 * Though they are JSON files, typically they files will have the .id extension. Therefore
 * if no extension is provided `.id` is added
 */
export default class JSONIDAdapter {

    private idFilesDir: string;
    private mspId: string = '';

    /** 
     * @param idFilesDir Directory to load the files from 
     * @param mspId optional MSPID to apply to all identities returned if they are missing it
     */
    public constructor(idFilesDir: string, mspId?: string) {
        this.idFilesDir = path.resolve(idFilesDir);
        
        if (mspId){
            this.mspId = mspId;
        }
        
    }

    private async readIDFile(idFile: string): Promise<JSONID> {
        let idJsonFile = path.resolve(path.join(this.idFilesDir, idFile));

        // check if there's no extension probably means it's a waller id file
        if (path.extname(idJsonFile)===""){
            idJsonFile = `${idJsonFile}.id`
        }

        let id: JSONID;
        let json = JSON.parse(await fs.readFile(idJsonFile, 'utf-8'));

        // look for the nested credentials element
        let credentials = json['credentials'];

        if (credentials){
            // v2 SDK Wallet format
            id = {
                name: idFile,
                cert: credentials['certificate'],
                ca: '',
                hsm: false,
                private_key: credentials['privateKey'],
                mspId : json.mspId
            }
        } else {
            // IBP exported ID style format
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

    /**
     * 
     * @param idFile the name of the identity to load (if no extension is provided `.id` is added)
     * @returns Identity to use with the GatewayBuilder
     */
    public async getIdentity(idFile: string): Promise<Identity> {
        const id = await this.readIDFile(idFile);
        
        const identity: Identity = {
            credentials: Buffer.from(id.cert),
            mspId: id.mspId
        };

        return identity;
    }

    /**
     * 
     * @param idFile the name of the identity to load (if no extension is provided `.id` is added)
     * @returns Signer to use with the GatewayBuilder
     */
    public async getSigner(idFile: string): Promise<Signer> {
        const id = await this.readIDFile(idFile);
        
        const privateKey = crypto.createPrivateKey(id.private_key);
        return signers.newPrivateKeySigner(privateKey);
    }

}