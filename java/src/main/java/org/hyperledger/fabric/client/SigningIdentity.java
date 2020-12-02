/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.security.GeneralSecurityException;
import java.util.function.Function;

import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;

final class SigningIdentity {
    private final Identity identity;
    private final Signer signer;
    private final Function<byte[], byte[]> hash = Hash::sha256;
    private final byte[] creator;

    SigningIdentity(final Identity identity, final Signer signer) {
        this.identity = identity;
        this.signer = signer;
        this.creator = GatewayUtils.serializeIdentity(identity);
    }

    public Identity getIdentity() {
        return identity;
    }

    public byte[] hash(byte[] message) {
        return hash.apply(message);
    }

    public byte[] sign(byte[] digest) {
        try {
            return signer.sign(digest);
        } catch (GeneralSecurityException e) {
            throw new GatewayRuntimeException(e);
        }
    }

    public byte[] getCreator() {
        return creator;
    }
}
