/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.identity;

import java.nio.charset.StandardCharsets;
import java.security.GeneralSecurityException;
import java.security.KeyPair;
import java.security.KeyPairGenerator;
import java.security.NoSuchAlgorithmException;

import org.bouncycastle.jce.provider.BouncyCastleProvider;
import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

public final class SignerTest {
    private static final X509Credentials CREDENTIALS = new X509Credentials();
    private static final byte[] DIGEST = "DIGEST".getBytes(StandardCharsets.UTF_8);

    @Test
    void new_signer_from_unsupported_private_key_type_throws_IllegalArgumentException() throws NoSuchAlgorithmException {
        KeyPairGenerator generator = KeyPairGenerator.getInstance("DSA", new BouncyCastleProvider());
        generator.initialize(2048);
        KeyPair keyPair = generator.generateKeyPair();

        assertThatThrownBy(() -> Signers.newPrivateKeySigner(keyPair.getPrivate()))
                .isInstanceOf(IllegalArgumentException.class);
    }

    @Test
    void sign_valid_digest() throws GeneralSecurityException {
        Signer signer = Signers.newPrivateKeySigner(CREDENTIALS.getPrivateKey());
        byte[] signature = signer.sign(DIGEST);

        assertThat(signature).isNotEmpty();
    }

    @Test
    void sign_null_digest_throws_NullPointerException() {
        Signer signer = Signers.newPrivateKeySigner(CREDENTIALS.getPrivateKey());

        assertThatThrownBy(() -> signer.sign(null))
            .isInstanceOf(NullPointerException.class);
    }
}
