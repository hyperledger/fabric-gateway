/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Iterator;
import java.util.function.Supplier;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.protos.gateway.ChaincodeEventsResponse;
import org.hyperledger.fabric.protos.gateway.GatewayGrpc;
import org.hyperledger.fabric.protos.gateway.SignedChaincodeEventsRequest;

/**
 * A Fabric Gateway call to obtain chaincode events. Supports off-line signing flow using {@link Network#newSignedChaincodeEvents(byte[], byte[])}.
 */
final class ChaincodeEventsSupplier implements Signable, Supplier<Iterator<ChaincodeEvent>> {
    private final GatewayGrpc.GatewayBlockingStub client;
    private final SigningIdentity signingIdentity;
    private SignedChaincodeEventsRequest signedRequest;

    ChaincodeEventsSupplier(final GatewayGrpc.GatewayBlockingStub client, final SigningIdentity signingIdentity,
                            final SignedChaincodeEventsRequest signedRequest) {
        this.client = client;
        this.signingIdentity = signingIdentity;
        this.signedRequest = signedRequest;
    }

    @Override
    public Iterator<ChaincodeEvent> get() {
        sign();
        Iterator<ChaincodeEventsResponse> responseIter = client.chaincodeEvents(signedRequest);
        return new ChaincodeEventIterator(responseIter);
    }

    @Override
    public byte[] getBytes() {
        return signedRequest.toByteArray();
    }

    @Override
    public byte[] getDigest() {
        byte[] message = signedRequest.getRequest().toByteArray();
        return signingIdentity.hash(message);
    }

    void setSignature(final byte[] signature) {
        signedRequest = signedRequest.toBuilder()
                .setSignature(ByteString.copyFrom(signature))
                .build();
    }

    private void sign() {
        if (isSigned()) {
            return;
        }

        byte[] digest = getDigest();
        byte[] signature = signingIdentity.sign(digest);
        setSignature(signature);
    }

    private boolean isSigned() {
        return !signedRequest.getSignature().isEmpty();
    }
}
