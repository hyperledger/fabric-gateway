/*
 * Copyright 2022 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;


import javax.json.JsonObject;
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
import java.nio.file.Paths;
import java.nio.file.StandardOpenOption;
import java.util.Collections;
import java.util.EnumSet;
import java.util.Set;
import javax.json.Json;
import javax.json.JsonReader;
import javax.json.JsonWriter;


/**
 * FileCheckpointer to store checkpointer state during file read write operations.
 */
public final class FileCheckpointer implements Checkpointer {

    private long blockNumber;
    private String transactionId = "";
    private String path = "";
    private final Reader fileReader;
    private final Writer fileWriter;
    private static final Set<OpenOption> OPEN_OPTIONS = Collections.unmodifiableSet(EnumSet.of(
            StandardOpenOption.CREATE,
            StandardOpenOption.READ,
            StandardOpenOption.WRITE
    ));
    private final FileChannel fileChannel;
    private static final String CONFIG_KEY_BLOCK = "blockNumber";
    private static final String CONFIG_KEY_TRANSACTIONID = "transactionId";

    FileCheckpointer(final String path) throws IOException {
        this.path = path;

        fileChannel = FileChannel.open(Paths.get(path), OPEN_OPTIONS);
        lockFile();

        CharsetEncoder utf8Encoder = StandardCharsets.UTF_8.newEncoder();
        fileWriter = Channels.newWriter(fileChannel, utf8Encoder, -1);

        CharsetDecoder utf8Decoder = StandardCharsets.UTF_8.newDecoder();
        fileReader = Channels.newReader(fileChannel, utf8Decoder, -1);
        this.loadFromFile();
        this.saveToFile();

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
        this.blockNumber = blockNumber + 1;
        this.transactionId = "";
        this.saveToFile();
    }

    @Override
    public void checkpointTransaction(final long blockNumber, final String transactionID) throws IOException {
        this.blockNumber = blockNumber;
        this.transactionId = transactionID;
        this.saveToFile();
    }

    @Override
    public void  checkpointChaincodeEvent(final ChaincodeEvent event) throws IOException {
        checkpointTransaction(event.getBlockNumber(), event.getTransactionId());
    }

    @Override
    public long getBlockNumber() {
        return this.blockNumber;
    }

    @Override
    public String getTransactionId() {
        return this.transactionId;
    }

    private void loadFromFile() throws IOException {
            JsonObject data = this.readFile();
            if (data != null) {
                parseJson(data);
            }
    }

    private JsonObject readFile() throws IOException {
        fileChannel.position(0);
        JsonReader jsonReader = Json.createReader(fileReader);

        try {
            return jsonReader.readObject();
        } catch (RuntimeException e) {
//            throw new IOException("Failed to parse checkpoint data from file: " + path, e);
        }
        return null;
    }

    private void parseJson(final JsonObject json) throws IOException {
        try {
            blockNumber = json.getJsonNumber(CONFIG_KEY_BLOCK).longValue();
            transactionId = json.getString(CONFIG_KEY_TRANSACTIONID);
        } catch (RuntimeException e) {
            throw new IOException("Bad format of checkpoint data from file: " + path, e);
        }
    }

    private void saveToFile() throws IOException {
        JsonObject jsonData = buildJson();
        fileChannel.position(0);
        saveJson(jsonData);
        fileChannel.truncate(fileChannel.position());
    }

    private void saveJson(final JsonObject json) throws IOException {
        JsonWriter jsonWriter = Json.createWriter(fileWriter);
        try {
            jsonWriter.writeObject(json);
        } catch (RuntimeException e) {
            throw new IOException("Failed to write checkpoint data to file: " + path, e);
        }
        fileWriter.flush();
    }

    private JsonObject buildJson() {
        return Json.createObjectBuilder()
                .add(CONFIG_KEY_BLOCK, blockNumber)
                .add(CONFIG_KEY_TRANSACTIONID, transactionId)
                .build();
    }

    /**
     * Releases the resources and closes the file channel.
     * @throws IOException
     */

    public void close() throws IOException {
        fileChannel.close(); // Also releases lock
    }
}

