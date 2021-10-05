/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import io.grpc.Channel;
import org.hyperledger.fabric.protos.gateway.ChaincodeEventsResponse;
import org.hyperledger.fabric.protos.gateway.CommitStatusResponse;
import org.hyperledger.fabric.protos.gateway.EndorseRequest;
import org.hyperledger.fabric.protos.gateway.EndorseResponse;
import org.hyperledger.fabric.protos.gateway.EvaluateRequest;
import org.hyperledger.fabric.protos.gateway.EvaluateResponse;
import org.hyperledger.fabric.protos.gateway.GatewayGrpc;
import org.hyperledger.fabric.protos.gateway.SignedChaincodeEventsRequest;
import org.hyperledger.fabric.protos.gateway.SignedCommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.SubmitRequest;
import org.hyperledger.fabric.protos.gateway.SubmitResponse;

final class GatewayClient {
    private final GatewayGrpc.GatewayBlockingStub blockingStub;
    private final GatewayGrpc.GatewayStub asyncStub;

    GatewayClient(final Channel channel) {
        this.blockingStub = GatewayGrpc.newBlockingStub(channel);
        this.asyncStub = GatewayGrpc.newStub(channel);
    }

    public EvaluateResponse evaluate(final EvaluateRequest request) {
        return blockingStub.evaluate(request);
    }

    public EndorseResponse endorse(final EndorseRequest request) {
        return blockingStub.endorse(request);
    }

    public SubmitResponse submit(final SubmitRequest request) {
        return blockingStub.submit(request);
    }

    public CommitStatusResponse commitStatus(final SignedCommitStatusRequest request) {
        return blockingStub.commitStatus(request);
    }

    public CloseableIterator<ChaincodeEventsResponse> chaincodeEvents(final SignedChaincodeEventsRequest request) {
        final ResponseObserver<SignedChaincodeEventsRequest, ChaincodeEventsResponse> responseObserver = new ResponseObserver<>();
        asyncStub.chaincodeEvents(request, responseObserver);
        return responseObserver.iterator();
    }
}
