/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.identity;

import org.bouncycastle.jce.provider.BouncyCastleProvider;
import org.hyperledger.fabric.client.Hash;
import org.junit.jupiter.api.Test;

import java.nio.charset.StandardCharsets;
import java.security.GeneralSecurityException;
import java.security.KeyPair;
import java.security.KeyPairGenerator;
import java.security.NoSuchAlgorithmException;
import java.security.Provider;
import java.security.Signature;
import java.security.cert.X509Certificate;
import java.util.Arrays;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

public final class SignerTest {
    private static final Provider PROVIDER = new BouncyCastleProvider();
    private static final byte[] MESSAGE = "MESSAGE".getBytes(StandardCharsets.UTF_8);

    private static void assertValidSignature(Signature verifier, X509Certificate certificate, final byte[] signature) throws GeneralSecurityException {
        verifier.initVerify(certificate);
        verifier.update(MESSAGE);
        assertThat(verifier.verify(signature))
                .withFailMessage("invalid signature: %s", Arrays.toString(signature))
                .isTrue();
    }

    @Test
    void new_signer_from_unsupported_private_key_type_throws_IllegalArgumentException() throws NoSuchAlgorithmException {
        KeyPairGenerator generator = KeyPairGenerator.getInstance("DSA", new BouncyCastleProvider());
        generator.initialize(2048);
        KeyPair keyPair = generator.generateKeyPair();

        assertThatThrownBy(() -> Signers.newPrivateKeySigner(keyPair.getPrivate()))
                .isInstanceOf(IllegalArgumentException.class);
    }

    @Test
    void sign_with_P256_key() throws GeneralSecurityException {
        X509Credentials credentials = new X509Credentials();
        Signer signer = Signers.newPrivateKeySigner(credentials.getPrivateKey());
        byte[] signature = signer.sign(Hash.SHA256.apply(MESSAGE));

        Signature verifier = Signature.getInstance("SHA256withECDSA", PROVIDER);
        assertValidSignature(verifier, credentials.getCertificate(), signature);
    }

    @Test
    void sign_null_digest_throws_NullPointerException() {
        X509Credentials credentials = new X509Credentials();
        Signer signer = Signers.newPrivateKeySigner(credentials.getPrivateKey());

        assertThatThrownBy(() -> signer.sign(null))
            .isInstanceOf(NullPointerException.class);
    }

    @Test
    void sign_with_P384_key() throws GeneralSecurityException {
        X509Credentials credentials = new X509Credentials(X509Credentials.Curve.P384);
        Signer signer = Signers.newPrivateKeySigner(credentials.getPrivateKey());
        byte[] signature = signer.sign(Hash.SHA384.apply(MESSAGE));

        Signature verifier = Signature.getInstance("SHA384withECDSA", PROVIDER);
        assertValidSignature(verifier, credentials.getCertificate(), signature);
    }

    @Test
    void sign_with_Ed25519_key() throws GeneralSecurityException {
        X509Credentials credentials = new X509Credentials(X509Credentials.Curve.Ed25519);
        Signer signer = Signers.newPrivateKeySigner(credentials.getPrivateKey());
        byte[] signature = signer.sign(MESSAGE);

        Signature verifier = Signature.getInstance(credentials.getPrivateKey().getAlgorithm(), PROVIDER);
        assertValidSignature(verifier, credentials.getCertificate(), signature);
    }
}
