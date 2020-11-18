/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.identity;

import java.nio.charset.StandardCharsets;
import java.security.cert.X509Certificate;
import java.util.Arrays;
import java.util.Objects;

public final class X509Identity implements Identity {
    private final String mspId;
    private final X509Certificate certificate;
    private final byte[] credentials;

    public X509Identity(final String mspId, final X509Certificate certificate) {
        this.mspId = mspId;
        this.certificate = certificate;
        credentials = Identities.toPemString(certificate).getBytes(StandardCharsets.UTF_8);
    }

    public X509Certificate getCertificate() {
        return certificate;
    }

    @Override
    public String getMspId() {
        return mspId;
    }

    @Override
    public byte[] getCredentials() {
        return credentials.clone();
    }

    @Override
    public boolean equals(final Object other) {
        if (this == other) {
            return true;
        }
        if (!(other instanceof X509Identity)) {
            return false;
        }

        X509Identity that = (X509Identity) other;
        return Objects.equals(this.mspId, that.mspId)
                && Arrays.equals(this.credentials, that.credentials);
    }

    @Override
    public int hashCode() {
        return Objects.hash(mspId, Arrays.hashCode(credentials));
    }
}
