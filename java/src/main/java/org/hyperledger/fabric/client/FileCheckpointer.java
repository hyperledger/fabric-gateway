/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import com.google.gson.Gson;
import com.google.gson.JsonObject;
import com.google.gson.stream.JsonReader;
import com.google.gson.stream.JsonWriter;

import java.io.IOException;
import java.io.Reader;
import java.io.Writer;
import java.nio.channels.Channels;
import java.nio.channels.FileChannel;
import java.nio.channels.FileLock;
import java.nio.channels.OverlappingFileLockException;
import java.nio.charset.CharsetDecoder;
import java.nio.charset.CharsetEncoder;
import java.nio.charset.StandardCharsets;
import java.nio.file.OpenOption;
import java.nio.file.Path;
import java.nio.file.StandardOpenOption;
import java.util.Collections;
import java.util.EnumSet;
import java.util.Optional;
import java.util.OptionalLong;
import java.util.Set;

/**
 * Checkpointer implementation backed by persistent file storage.
 * It can be used to checkpoint progress after successfully processing events, allowing eventing to be resumed from this point.
 */
public final class FileCheckpointer implements Checkpointer, AutoCloseable {
    private static final String CONFIG_KEY_BLOCK = "blockNumber";
    private static final String CONFIG_KEY_TRANSACTIONID = "transactionId";
    private static final Set<OpenOption> OPEN_OPTIONS = Collections.unmodifiableSet(EnumSet.of(
            StandardOpenOption.CREATE,
            StandardOpenOption.READ,
            StandardOpenOption.WRITE
    ));

    private OptionalLong blockNumber = OptionalLong.empty();
    private String transactionId;
    private final Path path;
    private final Reader fileReader;
    private final Writer fileWriter;
    private final FileChannel fileChannel;
    private final Gson gson = new Gson();

    /**
     * To create a checkpointer instance backed by persistent file storage.
     * @param path Path of the file which has to store the checkpointer state.
     * @throws IOException if the file cannot be opened, is unwritable, or contains invalid checkpointer state data.
     */
    public FileCheckpointer(final Path path) throws IOException {
        this.path = path;

        fileChannel = FileChannel.open(path, OPEN_OPTIONS);
        lockFile();

        CharsetEncoder utf8Encoder = StandardCharsets.UTF_8.newEncoder();
        fileWriter = Channels.newWriter(fileChannel, utf8Encoder, -1);

        CharsetDecoder utf8Decoder = StandardCharsets.UTF_8.newDecoder();
        fileReader = Channels.newReader(fileChannel, utf8Decoder, -1);
        if (fileChannel.size() > 0) {
            load();
        }
        save();
    }

    private void lockFile() throws IOException {
        final FileLock fileLock;
        try {
            fileLock = fileChannel.tryLock();
        } catch (OverlappingFileLockException e) {
            throw new IOException("File is already locked: " + path, e);
        }
        if (fileLock == null) {
            throw new IOException("Another process holds an overlapping lock for file: " + path);
        }
    }

    @Override
    public void checkpointBlock(final long blockNumber) throws IOException {
        checkpointTransaction(blockNumber + 1, null);
    }

    @Override
    public void checkpointTransaction(final long blockNumber, final String transactionID) throws IOException {
        this.blockNumber = OptionalLong.of(blockNumber);
        this.transactionId = transactionID;
        save();
    }

    @Override
    public void checkpointChaincodeEvent(final ChaincodeEvent event) throws IOException {
        checkpointTransaction(event.getBlockNumber(), event.getTransactionId());
    }

    @Override
    public OptionalLong getBlockNumber() {
        return blockNumber;
    }

    @Override
    public Optional<String> getTransactionId() {
        return Optional.ofNullable(transactionId);
    }

    private void load() throws IOException {
        JsonObject data = readFile();
        if (data != null) {
            parseJson(data);
        }
    }

    private JsonObject readFile() throws IOException {
        fileChannel.position(0);
        JsonReader jsonReader = new JsonReader(fileReader);
        try {
            return gson.fromJson(jsonReader, JsonObject.class);
        } catch (RuntimeException e) {
            throw new IOException("Failed to parse checkpoint data from file: " + path, e);
        }
    }

    private void parseJson(final JsonObject json) throws IOException {
        try {
            blockNumber = json.has(CONFIG_KEY_BLOCK) ? OptionalLong.of(json.get(CONFIG_KEY_BLOCK).getAsLong()) : OptionalLong.empty();
            transactionId = json.has(CONFIG_KEY_TRANSACTIONID) ? json.get(CONFIG_KEY_TRANSACTIONID).getAsString() : null;
        } catch (RuntimeException e) {
            throw new IOException("Bad format of checkpoint data from file: " + path, e);
        }
    }

    private void save() throws IOException {
        JsonObject jsonData = buildJson();
        fileChannel.position(0);
        saveJson(jsonData);
        fileChannel.truncate(fileChannel.position());
    }

    private void saveJson(final JsonObject json) throws IOException {
        JsonWriter jsonWriter = new JsonWriter(fileWriter);
        try {
            gson.toJson(json, jsonWriter);
        } catch (RuntimeException e) {
            throw new IOException("Failed to write checkpoint data to file: " + path, e);
        }
        fileWriter.flush();
    }

    private JsonObject buildJson() {
        JsonObject object = new JsonObject();
        blockNumber.ifPresent(block -> object.addProperty(CONFIG_KEY_BLOCK, block));
        if (transactionId != null) {
            object.addProperty(CONFIG_KEY_TRANSACTIONID, transactionId);
        }
        return object;
    }

    /**
     * Releases the resources and closes the file channel.
     * @throws IOException if an I/O error occurs.
     */
    public void close() throws IOException {
        fileChannel.close(); // Also releases lock
    }

    /**
     * Commits file changes to the storage device.
     * @throws IOException if an I/O error occurs.
     */
    public void sync() throws IOException {
        fileChannel.force(false);
    }
}
