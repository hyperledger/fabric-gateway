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
import org.hyperledger.fabric.protos.gateway.ChaincodeEventsResponse;
import org.hyperledger.fabric.protos.gateway.CommitStatusResponse;
import org.hyperledger.fabric.protos.gateway.EndorseRequest;
import org.hyperledger.fabric.protos.gateway.EndorseResponse;
import org.hyperledger.fabric.protos.gateway.EvaluateRequest;
import org.hyperledger.fabric.protos.gateway.EvaluateResponse;
import org.hyperledger.fabric.protos.gateway.SignedChaincodeEventsRequest;
import org.hyperledger.fabric.protos.gateway.SignedCommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.SubmitRequest;
import org.hyperledger.fabric.protos.gateway.SubmitResponse;
import org.hyperledger.fabric.protos.peer.Chaincode;
import org.hyperledger.fabric.protos.peer.ProposalPackage;
import org.hyperledger.fabric.protos.peer.ProposalPackage.SignedProposal;
import org.hyperledger.fabric.protos.peer.TransactionPackage;

public class GatewayServiceStub {
    private static final TestUtils utils = TestUtils.getInstance();

    public EndorseResponse endorse(final EndorseRequest request) {
        return utils.newEndorseResponse("PAYLOAD", request.getChannelId());
    }

    public SubmitResponse submit(final SubmitRequest request) {
        return utils.newSubmitResponse();
    }

    public EvaluateResponse evaluate(final EvaluateRequest request) {
        String payload = newPayload(request.getProposedTransaction());
        return utils.newEvaluateResponse(payload);
    }

    public CommitStatusResponse commitStatus(final SignedCommitStatusRequest request) {
        return utils.newCommitStatusResponse(TransactionPackage.TxValidationCode.VALID);
    }

    public Stream<ChaincodeEventsResponse> chaincodeEvents(final SignedChaincodeEventsRequest request) {
        return Stream.empty();
    }

    private String newPayload(SignedProposal requestProposal) {
        // create a mock payload string by concatenating the chaincode name, tx name and arguments from the request
        try {
            ProposalPackage.Proposal proposal = ProposalPackage.Proposal.parseFrom(requestProposal.getProposalBytes());
            ProposalPackage.ChaincodeProposalPayload chaincodeProposalPayload = ProposalPackage.ChaincodeProposalPayload.parseFrom(proposal.getPayload());
            Chaincode.ChaincodeInvocationSpec chaincodeInvocationSpec = Chaincode.ChaincodeInvocationSpec.parseFrom(chaincodeProposalPayload.getInput());
            String chaincodeId = chaincodeInvocationSpec.getChaincodeSpec().getChaincodeId().getName();
            List<ByteString> args = chaincodeInvocationSpec.getChaincodeSpec().getInput().getArgsList();
            return chaincodeId + args.stream().map(ByteString::toStringUtf8).collect(Collectors.joining());
        } catch (InvalidProtocolBufferException ex) {
            return ex.getMessage();
        }
    }
}
