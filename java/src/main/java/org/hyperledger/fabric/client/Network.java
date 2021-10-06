/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import com.google.protobuf.InvalidProtocolBufferException;

/**
 * The Network represents a Fabric network (channel). Network instances are obtained from a Gateway using the
 * {@link Gateway#getNetwork(String)} method.
 *
 * <p>The Network provides the ability for applications to:</p>
 * <ul>
 *     <li>Obtain a specific smart contract deployed to the network using {@link #getContract(String)}, in order to
 *     submit and evaluate transactions for that smart contract.</li>
 *     <li>Listen for events emitted when blocks are committed to the ledger using {@link #getChaincodeEvents(String)}
 *     or {@link #newChaincodeEventsRequest(String)}.</li>
 * </ul>
 *
 * <h3>Chaincode events example</h3>
 * <pre>{@code
 *     Iterator<ChaincodeEvent> events = network.newChaincodeEventsRequest("basic")
 *             .startBlock(101)
 *             .build()
 *             .getEvents();
 *     events.forEachRemaining(event -> {
 *         // Process event
 *     });
 * }</pre>
 */
public interface Network {
    /**
     * Get an instance of a contract on the current network.
     * @param chaincodeId The name of the chaincode that implements the smart contract.
     * @return The contract object.
     * @throws NullPointerException if the chaincode ID is null.
     */
    Contract getContract(String chaincodeId);

    /**
     * Get an instance of a contract on the current network.  If the chaincode instance contains more
     * than one smart contract class (available using the latest chaincode programming model), then an
     * individual class can be selected.
     * @param chaincodeId The name of the chaincode that implements the smart contract.
     * @param contractName The name of the smart contract within the chaincode.
     * @return The contract object.
     * @throws NullPointerException if the chaincode ID is null.
     */
    Contract getContract(String chaincodeId, String contractName);

    /**
     * Get the name of the network channel.
     * @return The network name.
     */
    String getName();

    /**
     * Create a commit with the specified digital signature, which can be used to access information about a
     * transaction that is committed to the ledger. Supports off-line signing flow.
     * @param bytes Serialized commit status request.
     * @param signature Digital signature.
     * @return A signed commit status request.
     * @throws InvalidProtocolBufferException if the supplied commit bytes are not a valid commit.
     */
    Commit newSignedCommit(byte[] bytes, byte[] signature) throws InvalidProtocolBufferException;

    /**
     * Get events emitted by transaction functions of a specific chaincode from the next committed block. Note that the
     * returned iterator may throw {@link io.grpc.StatusRuntimeException} during iteration if a gRPC connection error
     * occurs.
     * @param chaincodeId A chaincode ID.
     * @return Ordered sequence of events.
     * @throws NullPointerException if the chaincode ID is null.
     * @see #newChaincodeEventsRequest(String)
     */
    CloseableIterator<ChaincodeEvent> getChaincodeEvents(String chaincodeId);

    /**
     * Build a new chaincode events request, which can be used to obtain events emitted by transaction functions of a
     * specific chaincode. This can be used to specify a specific ledger start position. Supports offline signing flow.
     * @param chaincodeId A chaincode ID.
     * @return A chaincode events request builder.
     * @throws NullPointerException if the chaincode ID is null.
     */
    ChaincodeEventsRequest.Builder newChaincodeEventsRequest(String chaincodeId);

    /**
     * Create a chaincode events request with the specified digital signature, which can be used to obtain events
     * emitted by transaction functions of a specific chaincode. Supports off-line signing flow.
     * @param bytes Serialized chaincode events request.
     * @param signature Digital signature.
     * @return A signed chaincode events request.
     * @throws InvalidProtocolBufferException if the supplied chaincode events request bytes are not valid.
     */
    ChaincodeEventsRequest newSignedChaincodeEventsRequest(byte[] bytes, byte[] signature) throws InvalidProtocolBufferException;
}
