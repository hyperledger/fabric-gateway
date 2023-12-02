/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.function.UnaryOperator;

import io.grpc.CallOptions;
import org.hyperledger.fabric.protos.common.Block;
import org.hyperledger.fabric.protos.peer.BlockAndPrivateData;
import org.hyperledger.fabric.protos.peer.FilteredBlock;

/**
 * Network represents a network of nodes that are members of a specific Fabric channel. Network instances are obtained
 * from a Gateway using the {@link Gateway#getNetwork(String)} method.
 *
 * <p>The Network provides the ability for applications to:</p>
 * <ul>
 *     <li>Obtain a specific smart contract deployed to the network using {@link #getContract(String)}, in order to
 *     submit and evaluate transactions for that smart contract.</li>
 *     <li>Listen for chaincode events emitted by transactions when they are committed to the ledger using
 *     {@link #getChaincodeEvents(String)} or {@link #newChaincodeEventsRequest(String)}.</li>
 *     <li>Listen for block events emitted when blocks are committed to the ledger:
 *         <ul>
 *             <li><strong>Blocks</strong> using {@link #getBlockEvents()} or {@link #newBlockEventsRequest()}.</li>
 *             <li><strong>Filtered blocks</strong> {@link #getFilteredBlockEvents()} or {@link #newFilteredBlockEventsRequest()}.</li>
 *             <li><strong>Blocks and private data </strong> {@link #getBlockAndPrivateDataEvents()} or {@link #newBlockAndPrivateDataEventsRequest()}.</li>
 *         </ul>
 *     </li>
 * </ul>
 *
 * <p>To safely handle connection errors during eventing, it is recommended to use a checkpointer to track eventing
 * progress. This allows eventing to be resumed with no loss or duplication of events.</p>
 *
 * <p>Chaincode events example</p>
 * <pre>{@code
 *     Checkpointer checkpointer = new InMemoryCheckpointer();
 *     while (true) {
 *         ChaincodeEventsRequest request = network.newChaincodeEventsRequest("chaincodeName")
 *                 .checkpoint(checkpointer)
 *                 .startBlock(blockNumber) // Ignored if the checkpointer has checkpoint state
 *                 .build();
 *         try (CloseableIterator<ChaincodeEvent> events = request.getEvents()) {
 *             events.forEachRemaining(event -> {
 *                 // Process then checkpoint event
 *                 checkpointer.checkpointChaincodeEvent(event);
 *             });
 *         } catch (GatewayRuntimeException e) {
 *             // Connection error
 *         }
 *     }
 * }</pre>
 *
 * <p>Block events example</p>
 * <pre>{@code
 *     Checkpointer checkpointer = new InMemoryCheckpointer();
 *     while (true) {
 *         ChaincodeEventsRequest request = network.newBlockEventsRequest()
 *                 .checkpoint(checkpointer)
 *                 .startBlock(blockNumber) // Ignored if the checkpointer has checkpoint state
 *                 .build();
 *         try (CloseableIterator<Block> events = request.getEvents()) {
 *             events.forEachRemaining(event -> {
 *                 // Process then checkpoint block
 *                 checkpointer.checkpointBlock(event.getHeader().getNumber());
 *             });
 *         } catch (GatewayRuntimeException e) {
 *             // Connection error
 *         }
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
     * @return Ordered sequence of events.
     * @throws NullPointerException if the chaincode name is null.
     * @see #newChaincodeEventsRequest(String)
     */
    default CloseableIterator<ChaincodeEvent> getChaincodeEvents(String chaincodeName) {
        return getChaincodeEvents(chaincodeName, GatewayUtils.asCallOptions());
    }

    /**
     * Get events emitted by transaction functions of a specific chaincode from the next committed block. The Java gRPC
     * implementation may not begin reading events until the first use of the returned iterator.
     * <p>Note that the returned iterator may throw {@link io.grpc.StatusRuntimeException} during iteration if a gRPC connection error
     * occurs.</p>
     * @param chaincodeName A chaincode name.
     * @param options Function that transforms call options.
     * @return Ordered sequence of events.
     * @throws NullPointerException if the chaincode name is null.
     * @see #newChaincodeEventsRequest(String)
     */
    CloseableIterator<ChaincodeEvent> getChaincodeEvents(String chaincodeName, UnaryOperator<CallOptions> options);

    /**
     * Get events emitted by transaction functions of a specific chaincode from the next committed block. The Java gRPC
     * implementation may not begin reading events until the first use of the returned iterator.
     * <p>Note that the returned iterator may throw {@link io.grpc.StatusRuntimeException} during iteration if a gRPC connection error
     * occurs.</p>
     * @param chaincodeName A chaincode name.
     * @param options Call options.
     * @return Ordered sequence of events.
     * @throws NullPointerException if the chaincode name is null.
     * @deprecated Replaced by {@link #getChaincodeEvents(String, UnaryOperator)}.
     */
    @Deprecated
    default CloseableIterator<ChaincodeEvent> getChaincodeEvents(String chaincodeName, CallOption... options) {
        return getChaincodeEvents(chaincodeName, GatewayUtils.asCallOptions(options));
    }

    /**
     * Build a new chaincode events request, which can be used to obtain events emitted by transaction functions of a
     * specific chaincode. This can be used to specify a specific ledger start position. Supports offline signing flow.
     * @param chaincodeName A chaincode name.
     * @return A chaincode events request builder.
     * @throws NullPointerException if the chaincode name is null.
     */
    ChaincodeEventsRequest.Builder newChaincodeEventsRequest(String chaincodeName);

    /**
     * Get block events. The Java gRPC implementation may not begin reading events until the first use of the returned
     * iterator.
     * <p>Note that the returned iterator may throw {@link io.grpc.StatusRuntimeException} during iteration if a gRPC
     * connection error occurs.</p>
     * @return Ordered sequence of events.
     * @see #newBlockEventsRequest()
     */
    default CloseableIterator<Block> getBlockEvents() {
        return getBlockEvents(GatewayUtils.asCallOptions());
    }

    /**
     * Get block events. The Java gRPC implementation may not begin reading events until the first use of the returned
     * iterator.
     * <p>Note that the returned iterator may throw {@link io.grpc.StatusRuntimeException} during iteration if a gRPC
     * connection error occurs.</p>
     * @param options Function that transforms call options.
     * @return Ordered sequence of events.
     * @see #newBlockEventsRequest()
     */
    CloseableIterator<Block> getBlockEvents(UnaryOperator<CallOptions> options);

    /**
     * Build a request to receive block events. This can be used to specify a specific ledger start position. Supports
     * offline signing flow.
     * @return A block events request builder.
     * @throws NullPointerException if the chaincode name is null.
     */
    BlockEventsRequest.Builder newBlockEventsRequest();

    /**
     * Get filtered block events. The Java gRPC implementation may not begin reading events until the first use of the
     * returned iterator.
     * <p>Note that the returned iterator may throw {@link io.grpc.StatusRuntimeException} during iteration if a gRPC
     * connection error occurs.</p>
     * @return Ordered sequence of events.
     * @see #newFilteredBlockEventsRequest()
     */
    default CloseableIterator<FilteredBlock> getFilteredBlockEvents() {
        return getFilteredBlockEvents(GatewayUtils.asCallOptions());
    }

    /**
     * Get filtered block events. The Java gRPC implementation may not begin reading events until the first use of the
     * returned iterator.
     * <p>Note that the returned iterator may throw {@link io.grpc.StatusRuntimeException} during iteration if a gRPC
     * connection error occurs.</p>
     * @param options Function that transforms call options.
     * @return Ordered sequence of events.
     * @see #newFilteredBlockEventsRequest()
     */
    CloseableIterator<FilteredBlock> getFilteredBlockEvents(UnaryOperator<CallOptions> options);

    /**
     * Build a request to receive filtered block events. This can be used to specify a specific ledger start position.
     * Supports offline signing flow.
     * @return A filtered block events request builder.
     * @throws NullPointerException if the chaincode name is null.
     */
    FilteredBlockEventsRequest.Builder newFilteredBlockEventsRequest();

    /**
     * Get block and private data events. The Java gRPC implementation may not begin reading events until the first
     * use of the* returned iterator.
     * <p>Note that the returned iterator may throw {@link io.grpc.StatusRuntimeException} during iteration if a gRPC
     * connection error occurs.</p>
     * @return Ordered sequence of events.
     * @see #newBlockAndPrivateDataEventsRequest()
     */
    default CloseableIterator<BlockAndPrivateData> getBlockAndPrivateDataEvents() {
        return getBlockAndPrivateDataEvents(GatewayUtils.asCallOptions());
    }

    /**
     * Get block and private data events. The Java gRPC implementation may not begin reading events until the first
     * use of the* returned iterator.
     * <p>Note that the returned iterator may throw {@link io.grpc.StatusRuntimeException} during iteration if a gRPC
     * connection error occurs.</p>
     * @param options Function that transforms call options.
     * @return Ordered sequence of events.
     * @see #newBlockAndPrivateDataEventsRequest()
     */
    CloseableIterator<BlockAndPrivateData> getBlockAndPrivateDataEvents(UnaryOperator<CallOptions> options);

    /**
     * Build a request to receive block and private data events. This can be used to specify a specific ledger start
     * position. Supports offline signing flow.
     * @return A block and private data events request builder.
     * @throws NullPointerException if the chaincode name is null.
     */
    BlockAndPrivateDataEventsRequest.Builder newBlockAndPrivateDataEventsRequest();
}
