/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;

import java.security.GeneralSecurityException;
import java.util.function.Function;

final class SigningIdentity {
    private final Identity identity;
    private final Function<byte[], byte[]> hash;
    private final Signer signer;
    private final byte[] creator;

    SigningIdentity(final Identity identity, final Function<byte[], byte[]> hash, final Signer signer) {
        this.identity = identity;
        this.hash = hash;
        this.signer = signer;

        GatewayUtils.requireNonNullArgument(this.identity, "No identity supplied");
        GatewayUtils.requireNonNullArgument(this.hash, "No hash implementation supplied");
        GatewayUtils.requireNonNullArgument(this.signer, "No signing implementation supplied");

        this.creator = GatewayUtils.serializeIdentity(identity);
    }

    public Identity getIdentity() {
        return identity;
    }

    public byte[] hash(final byte[] message) {
        return hash.apply(message);
    }

    public byte[] sign(final byte[] digest) {
        try {
            return signer.sign(digest);
        } catch (GeneralSecurityException e) {
            throw new RuntimeException(e);
        }
    }

    public byte[] getCreator() {
        return creator;
    }
}
