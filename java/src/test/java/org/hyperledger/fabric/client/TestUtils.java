/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.io.IOException;
import java.io.Reader;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.attribute.FileAttribute;
import java.util.concurrent.atomic.AtomicLong;

import org.hyperledger.fabric.client.identity.Identities;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.X509Credentials;
import org.hyperledger.fabric.client.identity.X509Identity;
import org.hyperledger.fabric.client.impl.GatewayImpl;
//import org.hyperledger.fabric.protos.peer.ProposalResponsePackage;


public final class TestUtils {
    private static final TestUtils INSTANCE = new TestUtils();
    private static final String TEST_FILE_PREFIX = "fgj-test-";
    private static final String UNUSED_FILE_PREFIX = "fgj-unused-";

    private final AtomicLong currentTransactionId = new AtomicLong();

    public static TestUtils getInstance() {
        return INSTANCE;
    }

    private TestUtils() { }

    public GatewayImpl.Builder newGatewayBuilder() throws IOException {
        X509Credentials credentials = new X509Credentials();

        GatewayImpl.Builder builder = (GatewayImpl.Builder)Gateway.createBuilder();
        Identity id = new X509Identity("msp1", credentials.getCertificate());
        builder.identity(id)
                .networkConfig("localhost:7053");
        return builder;
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
