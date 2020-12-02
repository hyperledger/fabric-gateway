/*
 *  Copyright 2020 IBM All Rights Reserved.
 *
 *  SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.nio.charset.StandardCharsets;
import java.security.SecureRandom;

import com.google.protobuf.ByteString;
import org.bouncycastle.util.encoders.Hex;
import org.hyperledger.fabric.protos.common.Common;

class TransactionContext {
    private static final int NONCE_LENGTH = 24;
    private static final SecureRandom RANDOM = new SecureRandom();

    private final SigningIdentity signingIdentity;
    private final byte[] nonce;
    private final String transactionId;
    private final Common.SignatureHeader signatureHeader;

    public TransactionContext(final SigningIdentity signingIdentity) {
        this.signingIdentity = signingIdentity;
        nonce = newNonce();

        transactionId = newTransactionId();
        signatureHeader = newSignatureHeader();
    }

    private static byte[] newNonce() {
        byte[] values = new byte[NONCE_LENGTH];
        RANDOM.nextBytes(values);
        return values;
    }

    private String newTransactionId() {
        byte[] saltedCreator = GatewayUtils.concat(nonce, signingIdentity.getCreator());
        byte[] rawTransactionId = signingIdentity.hash(saltedCreator);
        byte[] hexTransactionId = Hex.encode(rawTransactionId);
        return new String(hexTransactionId, StandardCharsets.UTF_8);
    }

    private Common.SignatureHeader newSignatureHeader() {
        return Common.SignatureHeader.newBuilder()
                .setCreator(ByteString.copyFrom(signingIdentity.getCreator()))
                .setNonce(ByteString.copyFrom(nonce))
                .build();
    }

    public String getTransactionId() {
        return transactionId;
    }

    public Common.SignatureHeader getSignatureHeader() {
        return signatureHeader;
    }
}
