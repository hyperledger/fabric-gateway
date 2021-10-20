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
import org.hyperledger.fabric.protos.peer.ProposalResponsePackage.Response;
import org.hyperledger.fabric.protos.peer.TransactionPackage;

public final class TestUtils {
    private static final TestUtils INSTANCE = new TestUtils();
    private static final String TEST_FILE_PREFIX = "fgw-test-";
    private static final String UNUSED_FILE_PREFIX = "fgw-unused-";

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
    
    public EndorseResponse newEndorseResponse(String value) {
        PreparedTransaction preparedTransaction = newPreparedTransaction(value);
        return EndorseResponse.newBuilder()
                .setPreparedTransaction(preparedTransaction.getEnvelope())
                .setResult(preparedTransaction.getResult())
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

    public PreparedTransaction newPreparedTransaction(String payload) {
        Common.Envelope envelope = Common.Envelope.newBuilder()
                .setPayload(ByteString.copyFromUtf8(payload))
                .build();
        return PreparedTransaction.newBuilder()
                .setTransactionId(newFakeTransactionId())
                .setEnvelope(envelope)
                .setResult(newResponse(payload))
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
     * Create a new temporary directory that will be deleted when the JVM exits.
     * @param attributes Attributes to be assigned to the directory.
     * @return The temporary directory.
     * @throws IOException On error.
     */
    public Path createTempDirectory(FileAttribute<?>... attributes) throws IOException {
        Path tempDir = Files.createTempDirectory(TEST_FILE_PREFIX, attributes);
        tempDir.toFile().deleteOnExit();
        return tempDir;
    }

    /**
     * Create a new temporary file within a specific directory that will be deleted when the JVM exits.
     * @param directory A directory in which to create the file.
     * @param attributes Attributes to be assigned to the file.
     * @return The temporary file.
     * @throws IOException On error.
     */
    public Path createTempFile(Path directory, FileAttribute<?>... attributes) throws IOException {
        Path tempFile = Files.createTempFile(directory, TEST_FILE_PREFIX, null, attributes);
        tempFile.toFile().deleteOnExit();
        return tempFile;
    }

    /**
     * Create a new temporary file that will be deleted when the JVM exits.
     * @param attributes Attributes to be assigned to the file.
     * @return The temporary file.
     * @throws IOException On error.
     */
    public Path createTempFile(FileAttribute<?>... attributes) throws IOException {
        Path tempFile = Files.createTempFile(TEST_FILE_PREFIX, null, attributes);
        tempFile.toFile().deleteOnExit();
        return tempFile;
    }

    /**
     * Get a temporary file name within a specific directory that currently does not exist, and mark the file for
     * deletion when the JVM exits.
     * @param directory Parent directory for the file.
     * @return The temporary file.
     * @throws IOException On error.
     */
    public Path getUnusedFilePath(Path directory) throws IOException {
        Path tempFile = Files.createTempFile(directory, UNUSED_FILE_PREFIX, null);
        Files.delete(tempFile);
        tempFile.toFile().deleteOnExit();
        return tempFile;
    }

    /**
     * Get a temporary file name that currently does not exist, and mark the file for deletion when the JVM exits.
     * @return The temporary file.
     * @throws IOException On error.
     */
    public Path getUnusedFilePath() throws IOException {
        Path tempFile = Files.createTempFile(UNUSED_FILE_PREFIX, null);
        Files.delete(tempFile);
        tempFile.toFile().deleteOnExit();
        return tempFile;
    }

    /**
     * Get a temporary directory name that currently does not exist, and mark the directory for deletion when the JVM
     * exits.
     * @return The temporary directory.
     * @throws IOException On error.
     */
    public Path getUnusedDirectoryPath() throws IOException {
        Path tempDir = Files.createTempDirectory(UNUSED_FILE_PREFIX);
        tempDir.toFile().deleteOnExit();
        Files.delete(tempDir);
        return tempDir;
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
}
