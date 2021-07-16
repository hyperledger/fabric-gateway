/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Iterator;

import com.google.protobuf.ByteString;
import org.hyperledger.fabric.protos.gateway.ChaincodeEventsResponse;
import org.hyperledger.fabric.protos.gateway.GatewayGrpc;
import org.hyperledger.fabric.protos.gateway.SignedChaincodeEventsRequest;

/**
 * A Fabric Gateway call to obtain chaincode events. Supports off-line signing flow using
 * {@link Network#newSignedChaincodeEventsRequest(byte[], byte[])}.
 */
public final class ChaincodeEventsRequest implements Signable {
    private final GatewayGrpc.GatewayBlockingStub client;
    private final SigningIdentity signingIdentity;
    private SignedChaincodeEventsRequest signedRequest;

    ChaincodeEventsRequest(final GatewayGrpc.GatewayBlockingStub client, final SigningIdentity signingIdentity,
                            final SignedChaincodeEventsRequest signedRequest) {
        this.client = client;
        this.signingIdentity = signingIdentity;
        this.signedRequest = signedRequest;
    }

    /**
     * Get events emitted by transaction functions of a specific chaincode. Note that the returned {@link Iterator} may
     * throw {@link io.grpc.StatusRuntimeException} during iteration if a gRPC connection error occurs.
     * @return Ordered sequence of events.
     */
    public Iterator<ChaincodeEvent> getEvents() {
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
