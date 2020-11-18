/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.identity;

import java.nio.charset.StandardCharsets;
import java.security.GeneralSecurityException;
import java.security.interfaces.ECPrivateKey;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

public final class ECDSASignerTest {
    private static final X509Credentials CREDENTIALS = new X509Credentials();
    private static final byte[] DIGEST = "DIGEST".getBytes(StandardCharsets.UTF_8);

    private final Signer signer = Signers.newPrivateKeySigner((ECPrivateKey) CREDENTIALS.getPrivateKey());

    @Test
    void sign_valid_digest() throws GeneralSecurityException {
        byte[] signature = signer.sign(DIGEST);
        assertThat(signature).isNotEmpty();
    }

    @Test
    void sign_null_digest_throws_NullPointerException() {
        assertThatThrownBy(() -> signer.sign(null))
            .isInstanceOf(NullPointerException.class);
    }
}
