/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.io.IOException;
import java.io.Reader;
import java.io.UncheckedIOException;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;

import com.google.protobuf.ByteString;
import io.grpc.ManagedChannel;
import io.grpc.Server;
import io.grpc.inprocess.InProcessChannelBuilder;
import io.grpc.inprocess.InProcessServerBuilder;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.client.identity.Signers;
import org.hyperledger.fabric.client.identity.X509Credentials;
import org.hyperledger.fabric.client.identity.X509Identity;
import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.gateway.CommitStatusResponse;
import org.hyperledger.fabric.protos.gateway.EndorseResponse;
import org.hyperledger.fabric.protos.gateway.EvaluateResponse;
import org.hyperledger.fabric.protos.gateway.GatewayGrpc;
import org.hyperledger.fabric.protos.gateway.PreparedTransaction;
import org.hyperledger.fabric.protos.gateway.SubmitResponse;
import org.hyperledger.fabric.protos.peer.ProposalPackage;
import org.hyperledger.fabric.protos.peer.ProposalResponsePackage;
import org.hyperledger.fabric.protos.peer.ProposalResponsePackage.Response;
import org.hyperledger.fabric.protos.peer.TransactionPackage;

public final class TestUtils {
    private static final TestUtils INSTANCE = new TestUtils();

    private final AtomicLong currentTransactionId = new AtomicLong();
    private final X509Credentials credentials = new X509Credentials();

    public static TestUtils getInstance() {
        return INSTANCE;
    }

    private TestUtils() { }

    public X509Credentials getCredentials() {
        return credentials;
    }

    public ManagedChannel newChannelForService(GatewayGrpc.GatewayImplBase service) {
        String serverName = InProcessServerBuilder.generateName();
        Server server = InProcessServerBuilder.forName(serverName).addService(service).build();

        try {
            server.start();
        } catch (IOException e) {
            throw new UncheckedIOException(e);
        }

        return InProcessChannelBuilder.forName(serverName).directExecutor().build();
    }

    /**
     * Get a Gateway builder configured with a valid identity and signer.
     * @return A gateway builder implementation.
     */
    public Gateway.Builder newGatewayBuilder() {
        Identity id = new X509Identity("msp1", credentials.getCertificate());
        Signer signer = Signers.newPrivateKeySigner(credentials.getPrivateKey());
        return Gateway.newInstance()
                .identity(id)
                .signer(signer);
    }
    
    public EndorseResponse newEndorseResponse(String value, String channelName) {
        PreparedTransaction preparedTransaction = newPreparedTransaction(value, channelName);
        return EndorseResponse.newBuilder()
                .setPreparedTransaction(preparedTransaction.getEnvelope())
                .build();
    }
    
    public SubmitResponse newSubmitResponse() {
        return SubmitResponse.newBuilder()
                .build();
    }
    
    public EvaluateResponse newEvaluateResponse(String value) {
        return EvaluateResponse.newBuilder()
                .setResult(newResponse(value))
                .build();
    }
    
    public Response newResponse(String value) {
        return Response.newBuilder()
                .setPayload(ByteString.copyFromUtf8(value))
                .build();
    }

    public PreparedTransaction newPreparedTransaction(String result, String channelName) {
        Common.ChannelHeader channelHeader = Common.ChannelHeader.newBuilder()
                .setChannelId(channelName)
                .build();
        Common.Header header = Common.Header.newBuilder()
                .setChannelHeader(channelHeader.toByteString())
                .build();

        ProposalPackage.ChaincodeAction chaincodeAction = ProposalPackage.ChaincodeAction.newBuilder()
                .setResponse(newResponse(result))
                .build();
        ProposalResponsePackage.ProposalResponsePayload responsePayload = ProposalResponsePackage.ProposalResponsePayload.newBuilder()
                .setExtension(chaincodeAction.toByteString())
                .build();
        TransactionPackage.ChaincodeEndorsedAction endorsedAction = TransactionPackage.ChaincodeEndorsedAction.newBuilder()
                .setProposalResponsePayload(responsePayload.toByteString())
                .build();
        TransactionPackage.ChaincodeActionPayload actionPayload = TransactionPackage.ChaincodeActionPayload.newBuilder()
                .setAction(endorsedAction)
                .build();
        TransactionPackage.TransactionAction transactionAction = TransactionPackage.TransactionAction.newBuilder()
                .setPayload(actionPayload.toByteString())
                .build();
        TransactionPackage.Transaction transaction = TransactionPackage.Transaction.newBuilder()
                .addActions(transactionAction)
                .build();

        Common.Payload payload = Common.Payload.newBuilder()
                .setHeader(header)
                .setData(transaction.toByteString())
                .build();

        Common.Envelope envelope = Common.Envelope.newBuilder()
                .setPayload(payload.toByteString())
                .build();

        return PreparedTransaction.newBuilder()
                .setTransactionId(newFakeTransactionId())
                .setEnvelope(envelope)
                .build();
    }

    public CommitStatusResponse newCommitStatusResponse(TransactionPackage.TxValidationCode status) {
        return CommitStatusResponse.newBuilder()
                .setResult(status)
                .build();
    }

    public CommitStatusResponse newCommitStatusResponse(TransactionPackage.TxValidationCode status, long blockNumber) {
        return CommitStatusResponse.newBuilder()
                .setResult(status)
                .setBlockNumber(blockNumber)
                .build();
    }

    private String newFakeTransactionId() {
        return Long.toHexString(currentTransactionId.incrementAndGet());
    }

    /**
     * Create a new Reader instance that will fail on any read attempt with the provided exception message.
     * @param failMessage Read exception message.
     * @return A reader.
     */
    public Reader newFailingReader(final String failMessage) {
        return new Reader() {
            @Override
            public int read(char[] cbuf, int offset, int length) throws IOException {
                throw new IOException(failMessage);
            }
            @Override
            public void close() {
                // do nothing
            }
        };
    }

    public void shutdownChannel(final ManagedChannel channel, final long timeout, final TimeUnit timeUnit) {
        if (channel.isShutdown()) {
            return;
        }

        try {
            channel.shutdownNow().awaitTermination(timeout, timeUnit);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }
}
