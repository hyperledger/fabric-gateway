/*
 *  Copyright 2020 IBM All Rights Reserved.
 *
 *  SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.List;
import java.util.stream.Collectors;
import java.util.stream.Stream;

import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;
import org.hyperledger.fabric.gateway.Event;
import org.hyperledger.fabric.gateway.PreparedTransaction;
import org.hyperledger.fabric.gateway.ProposedTransaction;
import org.hyperledger.fabric.gateway.Result;
import org.hyperledger.fabric.protos.peer.Chaincode;
import org.hyperledger.fabric.protos.peer.ProposalPackage;

public class GatewayServiceStub {
    private static final TestUtils utils = TestUtils.getInstance();

    public PreparedTransaction endorse(final ProposedTransaction request) {
        return utils.newPreparedTransaction("PAYLOAD", "SIGNATURE");
    }

    public Stream<Event> submit(final PreparedTransaction request) {
        return Stream.of(newEvent(), newEvent());
    }

    public Result evaluate(final ProposedTransaction request) {
        String resultPayload = newPayload(request);
        return utils.newResult(resultPayload);
    }

    private Event newEvent() {
        return Event.newBuilder()
                .setValue(ByteString.copyFromUtf8("EVENT"))
                .build();
    }

    private String newPayload(ProposedTransaction request) {
        // create a mock payload string by concatenating the chaincode id, tx name and arguments from the request
        try {
            ProposalPackage.Proposal proposal = ProposalPackage.Proposal.parseFrom(request.getProposal().getProposalBytes());
            ProposalPackage.ChaincodeProposalPayload chaincodeProposalPayload = ProposalPackage.ChaincodeProposalPayload.parseFrom(proposal.getPayload());
            Chaincode.ChaincodeInvocationSpec chaincodeInvocationSpec = Chaincode.ChaincodeInvocationSpec.parseFrom(chaincodeProposalPayload.getInput());
            String chaincodeId = chaincodeInvocationSpec.getChaincodeSpec().getChaincodeId().getName();
            List<ByteString> args = chaincodeInvocationSpec.getChaincodeSpec().getInput().getArgsList();
            String payload = chaincodeId + args.stream().map(arg -> arg.toStringUtf8()).collect(Collectors.joining());
            return payload;
        } catch (InvalidProtocolBufferException ex) {
            return ex.getMessage();
        }
    }
}
