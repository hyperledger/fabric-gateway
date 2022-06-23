/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.ArrayList;
import java.util.List;

import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;
import org.hyperledger.fabric.protos.common.ChannelHeader;
import org.hyperledger.fabric.protos.common.Envelope;
import org.hyperledger.fabric.protos.common.Header;
import org.hyperledger.fabric.protos.common.Payload;
import org.hyperledger.fabric.protos.peer.ChaincodeAction;
import org.hyperledger.fabric.protos.peer.ChaincodeActionPayload;
import org.hyperledger.fabric.protos.peer.ProposalResponsePayload;
import org.hyperledger.fabric.protos.peer.TransactionAction;

final class TransactionEnvelopeParser {
    private final String channelName;
    private final ByteString result;

    TransactionEnvelopeParser(final Envelope envelope) {
        try {
            Payload payload = Payload.parseFrom(envelope.getPayload());
            channelName = parseChannelName(payload.getHeader());
            result = parseResult(payload);
        } catch (InvalidProtocolBufferException e) {
            throw new IllegalArgumentException("Invalid transaction payload", e);
        }
    }

    public String getChannelName() {
        return channelName;
    }

    public ByteString getResult() {
        return result;
    }

    private String parseChannelName(final Header header) throws InvalidProtocolBufferException {
        ChannelHeader channelHeader = ChannelHeader.parseFrom(header.getChannelHeader());
        return channelHeader.getChannelId();
    }

    private ByteString parseResult(final Payload payload) throws InvalidProtocolBufferException {
        org.hyperledger.fabric.protos.peer.Transaction transaction = org.hyperledger.fabric.protos.peer.Transaction.parseFrom(payload.getData());

        List<InvalidProtocolBufferException> parseExceptions = new ArrayList<>();

        for (TransactionAction transactionAction : transaction.getActionsList()) {
            try {
                return parseResult(transactionAction);
            } catch (InvalidProtocolBufferException e) {
                parseExceptions.add(e);
            }
        }

        IllegalArgumentException e = new IllegalArgumentException("No proposal response found");
        parseExceptions.forEach(e::addSuppressed);
        throw e;
    }

    private ByteString parseResult(final TransactionAction transactionAction) throws InvalidProtocolBufferException {
        ChaincodeActionPayload actionPayload = ChaincodeActionPayload.parseFrom(transactionAction.getPayload());
        ProposalResponsePayload responsePayload = ProposalResponsePayload.parseFrom(actionPayload.getAction().getProposalResponsePayload());
        ChaincodeAction chaincodeAction = ChaincodeAction.parseFrom(responsePayload.getExtension());
        return chaincodeAction.getResponse().getPayload();
    }
}
