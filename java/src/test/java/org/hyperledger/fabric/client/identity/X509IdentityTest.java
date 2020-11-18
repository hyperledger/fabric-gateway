/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.identity;

import java.nio.charset.StandardCharsets;
import java.security.cert.CertificateException;
import java.security.cert.X509Certificate;

import org.junit.jupiter.api.Test;

import static org.assertj.core.api.Assertions.assertThat;

public final class X509IdentityTest {
    private static final String mspId = "mspId";
    private static final X509Credentials credentials = new X509Credentials();

    private final X509Identity identity = new X509Identity(mspId, credentials.getCertificate());

    @Test
    void get_MSP_ID() {
        String actual = identity.getMspId();
        assertThat(actual).isEqualTo(mspId);
    }

    @Test
    void get_certificate() {
        X509Certificate actual = identity.getCertificate();
        assertThat(actual).isEqualTo(credentials.getCertificate());
    }

    @Test
    void get_credentials_returns_certificate_PEM() throws CertificateException {
        byte[] result = identity.getCredentials();

        X509Certificate certificate = Identities.readX509Certificate(new String(result, StandardCharsets.UTF_8));
        assertThat(certificate).isEqualTo(credentials.getCertificate());
    }

    @Test
    void equals_returns_true_for_equal_objects() throws CertificateException {
        // De/serialize credentials to ensure not just comparing the same objects
        X509Certificate certificate = Identities.readX509Certificate(credentials.getCertificatePem());
        Object other = new X509Identity(mspId, certificate);
        assertThat(identity).isEqualTo(other);
    }

    @Test
    void equals_returns_false_for_different_type() {
        assertThat(identity).isNotEqualTo(null);
    }

    @Test
    void equals_returns_false_for_unequal_MSP_ID() {
        Object other = new X509Identity("WRONG", credentials.getCertificate());
        assertThat(identity).isNotEqualTo(other);
    }

    @Test
    void equals_returns_false_for_unequal_certificate() {
        X509Credentials otherCredentials = new X509Credentials();
        Object other = new X509Identity(identity.getMspId(), otherCredentials.getCertificate());
        assertThat(identity).isNotEqualTo(other);
    }

    @Test
    void equal_objects_have_same_hashcode() throws CertificateException {
        // De/serialize credentials to ensure not just comparing the same objects
        X509Certificate certificate = Identities.readX509Certificate(credentials.getCertificatePem());
        X509Identity other = new X509Identity(mspId, certificate);

        assertThat(identity).hasSameHashCodeAs(other);
    }
}
