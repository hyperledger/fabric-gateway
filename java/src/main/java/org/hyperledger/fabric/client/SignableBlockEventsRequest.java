/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.protos.common.Envelope;

class SignableBlockEventsRequest implements Signable {
    private final SigningIdentity signingIdentity;
    private Envelope request;

    SignableBlockEventsRequest(final SigningIdentity signingIdentity, final Envelope request) {
        this.signingIdentity = signingIdentity;
        this.request = request;
    }

    protected Envelope getSignedRequest() {
        if (!isSigned()) {
            byte[] digest = getDigest();
            byte[] signature = signingIdentity.sign(digest);
            setSignature(signature);
        }

        return request;
    }

    @Override
    public byte[] getBytes() {
        return request.toByteArray();
    }

    @Override
    public byte[] getDigest() {
        byte[] message = request.getPayload().toByteArray();
        return signingIdentity.hash(message);
    }

    void setSignature(final byte[] signature) {
        request = request.toBuilder()
                .setSignature(ByteString.copyFrom(signature))
                .build();
    }

    private boolean isSigned() {
        return !request.getSignature().isEmpty();
    }
}
