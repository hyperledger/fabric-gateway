/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

/**
 * The Network represents a Fabric network (channel). Network instances are obtained from a Gateway using the
 * {@link Gateway#getNetwork(String)} method.
 *
 * <p>The Network provides the ability for applications to:</p>
 * <ul>
 *     <li>Obtain a specific smart contract deployed to the network using {@link #getContract(String)}, in order to
 *     submit and evaluate transactions for that smart contract.</li>
 *     <li>Listen for events emitted when blocks are committed to the ledger using
 *     {@link #getChaincodeEvents(String, CallOption...)} or {@link #newChaincodeEventsRequest(String)}.</li>
 * </ul>
 *
 * <p>Chaincode events example:</p>
 * <pre>{@code
 *     try (CloseableIterator<ChaincodeEvent> events = network.getChaincodeEvents("chaincodeName")) {
 *         events.forEachRemaining(event -> {
 *             // Process event
 *         });
 *     }
 * }</pre>
 * *
 * <p>Chaincode event replay example</p>
 * <pre>{@code
 *     ChaincodeEventsRequest request = network.newChaincodeEventsRequest("chaincodeName")
 *             .startBlock(blockNumber)
 *             .build();
 *     try (CloseableIterator<ChaincodeEvent> events = request.getEvents()) {
 *         events.forEachRemaining(event -> {
 *             // Process event
 *         });
 *     }
 * }</pre>
 */
public interface Network {
    /**
     * Get an instance of a contract on the current network.
     * @param chaincodeName The name of the chaincode that implements the smart contract.
     * @return The contract object.
     * @throws NullPointerException if the chaincode name is null.
     */
    Contract getContract(String chaincodeName);

    /**
     * Get an instance of a contract on the current network.  If the chaincode instance contains more
     * than one smart contract class (available using the latest chaincode programming model), then an
     * individual class can be selected.
     * @param chaincodeName The name of the chaincode that implements the smart contract.
     * @param contractName The name of the smart contract within the chaincode.
     * @return The contract object.
     * @throws NullPointerException if the chaincode name is null.
     */
    Contract getContract(String chaincodeName, String contractName);

    /**
     * Get the name of the network channel.
     * @return The network name.
     */
    String getName();

    /**
     * Get events emitted by transaction functions of a specific chaincode from the next committed block. The Java gRPC
     * implementation may not begin reading events until the first use of the returned iterator.
     * <p>Note that the returned iterator may throw {@link io.grpc.StatusRuntimeException} during iteration if a gRPC connection error
     * occurs.</p>
     * @param chaincodeName A chaincode name.
     * @param options Call options.
     * @return Ordered sequence of events.
     * @throws NullPointerException if the chaincode name is null.
     * @see #newChaincodeEventsRequest(String)
     */
    CloseableIterator<ChaincodeEvent> getChaincodeEvents(String chaincodeName, CallOption... options);

    /**
     * Build a new chaincode events request, which can be used to obtain events emitted by transaction functions of a
     * specific chaincode. This can be used to specify a specific ledger start position. Supports offline signing flow.
     * @param chaincodeName A chaincode name.
     * @return A chaincode events request builder.
     * @throws NullPointerException if the chaincode name is null.
     */
    ChaincodeEventsRequest.Builder newChaincodeEventsRequest(String chaincodeName);
}
