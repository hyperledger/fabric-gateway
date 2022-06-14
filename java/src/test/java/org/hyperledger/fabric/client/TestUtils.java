/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.io.IOException;
import java.io.Reader;
import java.io.UncheckedIOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.attribute.FileAttribute;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicLong;
import java.util.function.Consumer;
import java.util.function.Function;
import java.util.stream.Stream;

import com.google.protobuf.ByteString;
import io.grpc.BindableService;
import io.grpc.ManagedChannel;
import io.grpc.Server;
import io.grpc.inprocess.InProcessChannelBuilder;
import io.grpc.inprocess.InProcessServerBuilder;
import io.grpc.stub.ServerCallStreamObserver;
import io.grpc.stub.StreamObserver;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.client.identity.Signers;
import org.hyperledger.fabric.client.identity.X509Credentials;
import org.hyperledger.fabric.client.identity.X509Identity;
import org.hyperledger.fabric.protos.common.ChannelHeader;
import org.hyperledger.fabric.protos.common.Envelope;
import org.hyperledger.fabric.protos.common.Header;
import org.hyperledger.fabric.protos.common.Payload;
import org.hyperledger.fabric.protos.gateway.CommitStatusResponse;
import org.hyperledger.fabric.protos.gateway.EndorseResponse;
import org.hyperledger.fabric.protos.gateway.EvaluateResponse;
import org.hyperledger.fabric.protos.gateway.PreparedTransaction;
import org.hyperledger.fabric.protos.gateway.SubmitResponse;
import org.hyperledger.fabric.protos.peer.ChaincodeAction;
import org.hyperledger.fabric.protos.peer.ChaincodeActionPayload;
import org.hyperledger.fabric.protos.peer.ChaincodeEndorsedAction;
import org.hyperledger.fabric.protos.peer.ProposalResponsePayload;
import org.hyperledger.fabric.protos.peer.Response;
import org.hyperledger.fabric.protos.peer.Transaction;
import org.hyperledger.fabric.protos.peer.TransactionAction;
import org.hyperledger.fabric.protos.peer.TxValidationCode;

public final class TestUtils {
    private static final TestUtils INSTANCE = new TestUtils();
    private static final String TEST_FILE_PREFIX = "fg-test-";

    private final AtomicLong currentTransactionId = new AtomicLong();
    private final X509Credentials credentials = new X509Credentials();

    public static TestUtils getInstance() {
        return INSTANCE;
    }

    private TestUtils() { }

    public X509Credentials getCredentials() {
        return credentials;
    }

    public ManagedChannel newChannelForServices(BindableService service, BindableService... additionalServices) {
        String serverName = InProcessServerBuilder.generateName();
        InProcessServerBuilder serverBuilder = InProcessServerBuilder.forName(serverName).addService(service);
        for (BindableService additionalService : additionalServices) {
            serverBuilder.addService(additionalService);
        }
        Server server = serverBuilder.build();

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
        ChannelHeader channelHeader = ChannelHeader.newBuilder()
                .setChannelId(channelName)
                .build();
        Header header = Header.newBuilder()
                .setChannelHeader(channelHeader.toByteString())
                .build();

        ChaincodeAction chaincodeAction = ChaincodeAction.newBuilder()
                .setResponse(newResponse(result))
                .build();
        ProposalResponsePayload responsePayload = ProposalResponsePayload.newBuilder()
                .setExtension(chaincodeAction.toByteString())
                .build();
        ChaincodeEndorsedAction endorsedAction = ChaincodeEndorsedAction.newBuilder()
                .setProposalResponsePayload(responsePayload.toByteString())
                .build();
        ChaincodeActionPayload actionPayload = ChaincodeActionPayload.newBuilder()
                .setAction(endorsedAction)
                .build();
        TransactionAction transactionAction = TransactionAction.newBuilder()
                .setPayload(actionPayload.toByteString())
                .build();
        Transaction transaction = Transaction.newBuilder()
                .addActions(transactionAction)
                .build();

        Payload payload = Payload.newBuilder()
                .setHeader(header)
                .setData(transaction.toByteString())
                .build();

        Envelope envelope = Envelope.newBuilder()
                .setPayload(payload.toByteString())
                .build();

        return PreparedTransaction.newBuilder()
                .setTransactionId(newFakeTransactionId())
                .setEnvelope(envelope)
                .build();
    }

    public CommitStatusResponse newCommitStatusResponse(TxValidationCode status) {
        return CommitStatusResponse.newBuilder()
                .setResult(status)
                .build();
    }

    public CommitStatusResponse newCommitStatusResponse(TxValidationCode status, long blockNumber) {
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

    public <Request, Response> void invokeStubUnaryCall(
            final Function<Request, Response> stubCall,
            final Request request,
            final StreamObserver<Response> responseObserver
    ) {
        try {
            Response response = stubCall.apply(request);
            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(e);
        }
    }

    public <Request, Response> void invokeStubServerStreamingCall(
            final Function<Request, Stream<Response>> stubCall,
            final Request request,
            final StreamObserver<Response> responseObserver
    ) {
        try {
            stubCall.apply(request).forEachOrdered(responseObserver::onNext);
            responseObserver.onCompleted();
        } catch (Exception e) {
            responseObserver.onError(e);
        }
    }

    public <Request, Response> StreamObserver<Request> invokeStubDuplexCall(
            final Function<Stream<Request>, Stream<Response>> stubCall,
            final ServerCallStreamObserver<Response> responseObserver,
            final int initialRequestCount
    ) {
        BlockingQueue<Request> requestQueue = new LinkedBlockingQueue<>();
        CountDownLatch requestCountLatch = new CountDownLatch(initialRequestCount);
        CompletableFuture<Void> responseFuture = CompletableFuture.completedFuture(null);

        try {
            Stream<Response> responses = stubCall.apply(requestQueue.stream()); // Stub invocation may throw exception
            responseFuture = CompletableFuture.runAsync(() -> {
                try {
                    requestCountLatch.await();
                    responses.forEachOrdered(responseObserver::onNext);
                    responseObserver.onCompleted();
                } catch (Throwable t) {
                    responseObserver.onError(t);
                }
            });
        } catch (Exception e) {
            responseObserver.onError(e);
        }

        CompletableFuture<Void> finalResponseFuture = responseFuture;
        responseObserver.setOnCancelHandler(() -> finalResponseFuture.cancel(true)); // Avoids gRPC error if cancel is called more than once
        return streamObserverFromQueue(
                requestQueue,
                request -> requestCountLatch.countDown(),
                throwable -> { },
                () -> finalResponseFuture.cancel(true)
        );
    }

    private <T> StreamObserver<T> streamObserverFromQueue(
            final BlockingQueue<T> queue,
            final Consumer<T> onNextListener,
            final Consumer<Throwable> onError,
            final Runnable onCompleted
    ) {
        return new StreamObserver<T>() {
            @Override
            public void onNext(final T request) {
                try {
                    onNextListener.accept(request);
                    queue.put(request);
                } catch (InterruptedException e) {
                    onError.accept(e);
                }
            }

            @Override
            public void onError(final Throwable t) {
                onError.accept(t);
            }

            @Override
            public void onCompleted() {
                onCompleted.run();
            }
        };
    }

    public Path createTempFile(String sufix, FileAttribute<?>... attributes) throws IOException {
        Path tempFile = Files.createTempFile(TEST_FILE_PREFIX, sufix, attributes);
        tempFile.toFile().deleteOnExit();
        return tempFile;
    }
}
