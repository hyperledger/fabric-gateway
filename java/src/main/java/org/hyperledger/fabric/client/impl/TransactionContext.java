/*
 *  Copyright 2020 IBM All Rights Reserved.
 *
 *  SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import java.nio.charset.StandardCharsets;
import java.security.SecureRandom;
import java.util.function.Function;

import com.google.protobuf.ByteString;
import org.bouncycastle.util.encoders.Hex;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.protos.common.Common;

class TransactionContext {
    private static final int NONCE_LENGTH = 24;
    private static final SecureRandom RANDOM = new SecureRandom();

    private final byte[] serializedIdentity;
    private final Function<byte[], byte[]> hasher;
    private final byte[] nonce;

    public TransactionContext(Identity identity, Function<byte[], byte[]> hasher) {
        this.hasher = hasher;
        serializedIdentity = GatewayUtils.serializeIdentity(identity);
        nonce = generateNonce();
    }

    private static byte[] generateNonce() {
        byte[] values = new byte[NONCE_LENGTH];
        RANDOM.nextBytes(values);
        return values;
    }

    public String getTransactionId() {
        byte[] digest = GatewayUtils.concat(nonce, serializedIdentity);
        byte[] hash = hasher.apply(digest);
        byte[] hex = Hex.encode(hash);
        return new String(hex, StandardCharsets.UTF_8);
    }

    public Common.SignatureHeader getSignatureHeader() {
        return Common.SignatureHeader.newBuilder()
                .setCreator(ByteString.copyFrom(serializedIdentity))
                .setNonce(ByteString.copyFrom(nonce))
                .build();
    }
}
