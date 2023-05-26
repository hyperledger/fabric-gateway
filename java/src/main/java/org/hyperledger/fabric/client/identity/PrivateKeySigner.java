/*
 * Copyright 2023 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.identity;

import org.bouncycastle.jce.provider.BouncyCastleProvider;

import java.security.GeneralSecurityException;
import java.security.PrivateKey;
import java.security.Provider;
import java.security.Signature;

final class PrivateKeySigner implements Signer {
    private static final Provider PROVIDER = new BouncyCastleProvider();

    private final PrivateKey privateKey;
    private final String algorithm;

    PrivateKeySigner(final PrivateKey privateKey, final String algorithm) {
        this.privateKey = privateKey;
        this.algorithm = algorithm;
    }

    @Override
    public byte[] sign(final byte[] digest) throws GeneralSecurityException {
        Signature signer = Signature.getInstance(algorithm, PROVIDER);
        signer.initSign(privateKey);
        signer.update(digest);
        return signer.sign();
    }
}
