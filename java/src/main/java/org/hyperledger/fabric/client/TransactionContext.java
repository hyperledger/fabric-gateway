/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import com.google.protobuf.ByteString;
import org.bouncycastle.util.encoders.Hex;
import org.hyperledger.fabric.protos.common.SignatureHeader;

import java.nio.charset.StandardCharsets;
import java.security.SecureRandom;

final class TransactionContext {
    private static final int NONCE_LENGTH = 24;
    private static final SecureRandom RANDOM = new SecureRandom();

    private final SigningIdentity signingIdentity;
    private final byte[] nonce;
    private final String transactionId;
    private final SignatureHeader signatureHeader;

    TransactionContext(final SigningIdentity signingIdentity) {
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
        byte[] rawTransactionId = Hash.SHA256.apply(saltedCreator);
        byte[] hexTransactionId = Hex.encode(rawTransactionId);
        return new String(hexTransactionId, StandardCharsets.UTF_8);
    }

    private SignatureHeader newSignatureHeader() {
        return SignatureHeader.newBuilder()
                .setCreator(ByteString.copyFrom(signingIdentity.getCreator()))
                .setNonce(ByteString.copyFrom(nonce))
                .build();
    }

    public String getTransactionId() {
        return transactionId;
    }

    public SignatureHeader getSignatureHeader() {
        return signatureHeader;
    }
}
