/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package com.example;

import java.io.IOException;
import java.io.InputStream;
import java.io.StringReader;
import java.nio.file.Files;
import java.nio.file.Path;
import java.security.InvalidKeyException;
import java.security.PrivateKey;
import java.security.cert.CertificateException;
import java.security.cert.X509Certificate;

import java.util.logging.Logger;
import javax.json.Json;
import javax.json.JsonObject;

import org.hyperledger.fabric.client.identity.Identities;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.client.identity.Signers;
import org.hyperledger.fabric.client.identity.X509Identity;

/**
 * This class can be used to map identities in a variety of JSON formats to the
 * Identity and Signers required for the gateway. For example if you have an
 * application wallet, or have exported IDs from SaaS
 * 
 * ``` 
 * // setup the loading of the identities JsonIdAdapter idWallet = new
 * JsonIdAdapter(Paths.get(walletDir)); Identity id =
 * idWallet.getIdentity(userid); Signer signer = idWallet.getSigner(userid);
 * Gateway.Builder builder =
 * Gateway.newInstance().identity(id).signer(signer).connection(grpcChannel);
 * 
 * ```
 * 
 * Though they are JSON files, typically they files will have the .id extension.
 * Therefore if no extension is provided `.id` is added
 */
public class JsonIdAdapter {
    private static final Logger LOGGER = Logger.getLogger(JsonIdAdapter.class.getName());
    public static final String JSON_VERSION = "version";
    public static final String JSON_TYPE = "type";
    public static final String JSON_MSP_ID = "mspId";

    private static final String JSON_CREDENTIALS = "credentials";
    private static final String JSON_CERTIFICATE = "certificate";
    private static final String JSON_PRIVATE_KEY = "privateKey";
    private static final String DATA_FILE_EXTENTION = ".id";

    private final Path idFilesDir;

    /** 
     * @param idFilesDir
     */
    public JsonIdAdapter(final Path idFilesDir) {
        this.idFilesDir = idFilesDir;

        if (!Files.exists(idFilesDir)) {
            throw new RuntimeException("Directory " + idFilesDir + " does not exist");
        }

        LOGGER.info("Reading for " + idFilesDir);

    }

    private Path getPathForLabel(final String label) {
        return idFilesDir.resolve(label + DATA_FILE_EXTENTION);
    }

    public Identity getIdentity(String id) throws CertificateException, IOException {
        JsonObject json = readFile(id);
        String mspId = json.getString(JSON_MSP_ID);
        JsonObject credentials = json.getJsonObject(JSON_CREDENTIALS);
        String certificatePem = credentials.getString(JSON_CERTIFICATE);

        X509Certificate certificate = Identities.readX509Certificate(new StringReader(certificatePem));
        return new X509Identity(mspId, certificate);
    }

    public Signer getSigner(String id) throws InvalidKeyException, IOException {
        JsonObject json = readFile(id);
        JsonObject credentials = json.getJsonObject(JSON_CREDENTIALS);
        String privateKeyPem = credentials.getString(JSON_PRIVATE_KEY);

        PrivateKey privateKey = Identities.readPrivateKey(new StringReader(privateKeyPem));

        return Signers.newPrivateKeySigner(privateKey);
    }

    private JsonObject readFile(final String label) throws IOException {
        InputStream identityData = Files.newInputStream(getPathForLabel(label));
        JsonObject identityJson = Json.createReader(identityData).readObject();
        return identityJson;
    }
}
