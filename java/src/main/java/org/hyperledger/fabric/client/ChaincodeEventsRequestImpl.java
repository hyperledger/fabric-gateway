/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.function.UnaryOperator;

import com.google.protobuf.ByteString;
import io.grpc.CallOptions;
import org.hyperledger.fabric.protos.gateway.ChaincodeEventsResponse;
import org.hyperledger.fabric.protos.gateway.SignedChaincodeEventsRequest;

final class ChaincodeEventsRequestImpl implements ChaincodeEventsRequest {
    private final GatewayClient client;
    private final SigningIdentity signingIdentity;
    private SignedChaincodeEventsRequest signedRequest;

    ChaincodeEventsRequestImpl(final GatewayClient client, final SigningIdentity signingIdentity,
                               final SignedChaincodeEventsRequest signedRequest) {
        this.client = client;
        this.signingIdentity = signingIdentity;
        this.signedRequest = signedRequest;
    }

    @Override
    public CloseableIterator<ChaincodeEvent> getEvents(final UnaryOperator<CallOptions> options) {
        sign();
        CloseableIterator<ChaincodeEventsResponse> responseIter = client.chaincodeEvents(signedRequest, options);
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
