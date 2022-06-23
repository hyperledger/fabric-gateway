/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import io.grpc.CallOptions;
import io.grpc.Channel;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;

import java.util.function.Function;
import java.util.function.UnaryOperator;

/**
 * The Gateway provides the connection point for an application to access the Fabric network as a specific user. It is
 * instantiated from a Builder instance that is created using {@link #newInstance()} and configured using a gateway URL
 * and a signing identity. It can then be connected to a fabric network using the
 * {@link Builder#connect()} method. Once connected, it can then access individual {@link Network} instances (channels)
 * using the {@link #getNetwork(String) getNetwork} method which in turn can access the {@link Contract} installed on a
 * network and {@link Contract#submitTransaction(String, String...) submit transactions} to the ledger.
 *
 * <p>Gateway instances should be reused for multiple transaction invocations and only closed once connection to the
 * Fabric network is no longer required.</p>
 *
 * <p>Multiple Gateway instances may share the same underlying gRPC connection by supplying the gRPC {@code Channel} as
 * an option to the Gateway connect.</p>
 *
 * <pre>{@code
 *     Identity identity = new X509Identity("mspId", certificate);
 *     Signer signer = Signers.newPrivateKeySigner(privateKey);
 *
 *     Gateway.Builder builder = Gateway.newInstance()
 *             .identity(identity)
 *             .signer(signer)
 *             .connection(grpcChannel);
 *
 *     try (Gateway gateway = builder.connect()) {
 *         Network network = gateway.getNetwork("channel");
 *         // Interactions with the network
 *     }
 * }</pre>
 */
public interface Gateway extends AutoCloseable {
    /**
     * Creates a gateway builder which is used to configure and connect a new Gateway instance.
     * @return A gateway builder.
     */
    static Builder newInstance() {
        return new GatewayImpl.Builder();
    }

    /**
     * Returns the identity used to interact with Fabric.
     * @return A client identity.
     */
    Identity getIdentity();

    /**
     * Returns an object representing a network.
     *
     * @param networkName The name of the network (channel name)
     * @return A network.
     * @throws NullPointerException if the network name is null.
     */
    Network getNetwork(String networkName);

    /**
     * Create a proposal with the specified digital signature. Supports off-line signing flow.
     * @param proposalBytes The proposal.
     * @param signature A digital signature.
     * @return A signed proposal.
     * @throws IllegalArgumentException if the supplied proposal bytes are not a valid proposal.
     */
    Proposal newSignedProposal(byte[] proposalBytes, byte[] signature);

    /**
     * Recreates a proposal from serialized data.
     * @param proposalBytes The proposal.
     * @return A  proposal.
     * @throws IllegalArgumentException if the supplied proposal bytes are not a valid proposal.
     */
    Proposal newProposal(byte[] proposalBytes);

    /**
     * Create a transaction with the specified digital signature. Supports off-line signing flow.
     * @param transactionBytes The transaction.
     * @param signature A digital signature.
     * @return A signed transaction.
     * @throws IllegalArgumentException if the supplied transaction bytes are not a valid transaction.
     */
    Transaction newSignedTransaction(byte[] transactionBytes, byte[] signature);

    /**
     * Recreates a transaction from serialized data.
     * @param transactionBytes The transaction.
     * @return A transaction.
     * @throws IllegalArgumentException if the supplied transaction bytes are not a valid transaction.
     */
    Transaction newTransaction(byte[] transactionBytes);

    /**
     * Create a commit with the specified digital signature, which can be used to access information about a
     * transaction that is committed to the ledger. Supports off-line signing flow.
     * @param bytes Serialized commit status request.
     * @param signature Digital signature.
     * @return A signed commit status request.
     * @throws IllegalArgumentException if the supplied commit bytes are not a valid commit.
     */
    Commit newSignedCommit(byte[] bytes, byte[] signature);

    /**
     * Recreates a commit from serialized data.
     * @param bytes Serialized commit status request.
     * @return A signed commit status request.
     * @throws IllegalArgumentException if the supplied commit bytes are not a valid commit.
     */
    Commit newCommit(byte[] bytes);

    /**
     * Create a chaincode events request with the specified digital signature, which can be used to obtain events
     * emitted by transaction functions of a specific chaincode. Supports off-line signing flow.
     * @param bytes Serialized chaincode events request.
     * @param signature Digital signature.
     * @return A signed chaincode events request.
     * @throws IllegalArgumentException if the supplied chaincode events request bytes are not valid.
     */
    ChaincodeEventsRequest newSignedChaincodeEventsRequest(byte[] bytes, byte[] signature);

    /**
     * Recreates a chaincode events request from serialized data.
     * @param bytes Serialized chaincode events request.
     * @return A chaincode events request.
     * @throws IllegalArgumentException if the supplied chaincode events request bytes are not valid.
     */
    ChaincodeEventsRequest newChaincodeEventsRequest(byte[] bytes);

    /**
     * Create a block events request with the specified digital signature, which can be used to receive block events.
     * Supports off-line signing flow.
     * @param bytes Serialized block events request.
     * @param signature Digital signature.
     * @return A signed block events request.
     * @throws IllegalArgumentException if the supplied request bytes are not valid.
     */
    BlockEventsRequest newSignedBlockEventsRequest(byte[] bytes, byte[] signature);

    /**
     * Recreates a block events request from serialized bytes.
     * @param bytes Serialized block events request.
     * @return A block events request.
     * @throws IllegalArgumentException if the supplied request bytes are not valid.
     */
    BlockEventsRequest newBlockEventsRequest(byte[] bytes);

    /**
     * Create a filtered block events request with the specified digital signature, which can be used to receive
     * filtered block events.* Supports off-line signing flow.
     * @param bytes Serialized filtered block events request.
     * @param signature Digital signature.
     * @return A signed filtered block events request.
     * @throws IllegalArgumentException if the supplied request bytes are not valid.
     */
    FilteredBlockEventsRequest newSignedFilteredBlockEventsRequest(byte[] bytes, byte[] signature);

    /**
     * Recreates a filtered block event request from serialized data.
     * @param bytes Serialized filtered block events request.
     * @return A filtered block events request.
     * @throws IllegalArgumentException if the supplied request bytes are not valid.
     */
    FilteredBlockEventsRequest newFilteredBlockEventsRequest(byte[] bytes);

    /**
     * Create a block and private data events request with the specified digital signature, which can be used to
     * receive block and private data events. Supports off-line signing flow.
     * @param bytes Serialized block and private data events request.
     * @param signature Digital signature.
     * @return A signed block and private data events request.
     * @throws IllegalArgumentException if the supplied request bytes are not valid.
     */
    BlockAndPrivateDataEventsRequest newSignedBlockAndPrivateDataEventsRequest(byte[] bytes, byte[] signature);

    /**
     * Recreate a block and private data events request from serialized data.
     * @param bytes Serialized block and private data events request.
     * @return A block and private data events request.
     * @throws IllegalArgumentException if the supplied request bytes are not valid.
     */
    BlockAndPrivateDataEventsRequest newBlockAndPrivateDataEventsRequest(byte[] bytes);

    /**
     * Close the gateway connection and all associated resources, including removing listeners attached to networks and
     * contracts created by the gateway.
     */
    void close();

    /**
     * The builder is used to specify the options used when connecting a Gateway. An instance of builder is created
     * using the static method {@link Gateway#newInstance()}. Every method on the builder object will return
     * a reference to the same builder object allowing them to be chained together in a single line, terminating with
     * a call to {@link #connect()} to complete connection of the Gateway.
     */
    interface Builder {
        /**
         * Specifies an existing gRPC connection to be used by the Gateway. The connection will not be closed when the
         * Gateway instance is closed. This allows multiple Gateway instances to share a gRPC connection.
         * @param grpcChannel A gRPC connection.
         * @return The builder instance, allowing multiple configuration options to be chained.
         */
        Builder connection(Channel grpcChannel);

        /**
         * Specifies the client identity used to connect to the network. All interactions will the Fabric network using
         * this Gateway will be performed by this identity.
         * @param identity An identity.
         * @return The builder instance, allowing multiple configuration options to be chained.
         */
        Builder identity(Identity identity);

        /**
         * Specify the signing implementation used to sign messages sent to the Fabric network.
         * @param signer A signing implementation.
         * @return The builder instance, allowing multiple configuration options to be chained.
         */
        Builder signer(Signer signer);

        /**
         * Specify the hashing implementation used to generate digests of messages sent to the Fabric network.
         * @param hash A hashing function.
         * @return The builder instance, allowing multiple configuration options to be chained.
         */
        Builder hash(Function<byte[], byte[]> hash);

        /**
         * Specify the default call options for evaluating transactions.
         * <p>A call of:</p>
         * <pre>{@code
         *     evaluateOptions(CallOption.deadlineAfter(5, TimeUnit.SECONDS))
         * }</pre>
         * <p>is equivalent to:</p>
         * <pre>{@code
         *     evaluateOptions(options -> options.withDeadlineAfter(5, TimeUnit.SECONDS))
         * }</pre>
         * @param options Call options.
         * @return The builder instance, allowing multiple configuration options to be chained.
         * @deprecated Replaced by {@link #evaluateOptions(UnaryOperator)}.
         */
        @Deprecated
        default Builder evaluateOptions(CallOption... options) {
            return evaluateOptions(GatewayUtils.asCallOptions(options));
        }

        /**
         * Specify the default call options for evaluating transactions.
         * @param options Function that transforms call options.
         * @return The builder instance, allowing multiple configuration options to be chained.
         */
        Builder evaluateOptions(UnaryOperator<CallOptions> options);

        /**
         * Specify the default call options for endorsements.
         * <p>A call of:</p>
         * <pre>{@code
         *     endorseOptions(CallOption.deadlineAfter(5, TimeUnit.SECONDS))
         * }</pre>
         * <p>is equivalent to:</p>
         * <pre>{@code
         *     endorseOptions(options -> options.withDeadlineAfter(5, TimeUnit.SECONDS))
         * }</pre>
         * @param options Call options.
         * @return The builder instance, allowing multiple configuration options to be chained.
         * @deprecated Replaced by {@link #endorseOptions(UnaryOperator)}.
         */
        @Deprecated
        default Builder endorseOptions(CallOption... options) {
            return endorseOptions(GatewayUtils.asCallOptions(options));
        }

        /**
         * Specify the default call options for endorsements.
         * @param options Function that transforms call options.
         * @return The builder instance, allowing multiple configuration options to be chained.
         */
        Builder endorseOptions(UnaryOperator<CallOptions> options);

        /**
         * Specify the default call options for submit of transactions to the orderer.
         * <p>A call of:</p>
         * <pre>{@code
         *     submitOptions(CallOption.deadlineAfter(5, TimeUnit.SECONDS))
         * }</pre>
         * <p>is equivalent to:</p>
         * <pre>{@code
         *     submitOptions(options -> options.withDeadlineAfter(5, TimeUnit.SECONDS))
         * }</pre>
         * @param options Call options.
         * @return The builder instance, allowing multiple configuration options to be chained.
         * @deprecated Replaced by {@link #submitOptions(UnaryOperator)}.
         */
        @Deprecated
        default Builder submitOptions(CallOption... options) {
            return submitOptions(GatewayUtils.asCallOptions(options));
        }

        /**
         * Specify the default call options for submit of transactions to the orderer.
         * @param options Function that transforms call options.
         * @return The builder instance, allowing multiple configuration options to be chained.
         */
        Builder submitOptions(UnaryOperator<CallOptions> options);

        /**
         * Specify the default call options for retrieving transaction commit status.
         * <p>A call of:</p>
         * <pre>{@code
         *     commitStatusOptions(CallOption.deadlineAfter(5, TimeUnit.SECONDS))
         * }</pre>
         * <p>is equivalent to:</p>
         * <pre>{@code
         *     commitStatusOptions(options -> options.withDeadlineAfter(5, TimeUnit.SECONDS))
         * }</pre>
         * @param options Call options.
         * @return The builder instance, allowing multiple configuration options to be chained.
         * @deprecated Replaced by {@link #commitStatusOptions(UnaryOperator)}.
         */
        @Deprecated
        default Builder commitStatusOptions(CallOption... options) {
            return commitStatusOptions(GatewayUtils.asCallOptions(options));
        }

        /**
         * Specify the default call options for retrieving transaction commit status.
         * @param options Function that transforms call options.
         * @return The builder instance, allowing multiple configuration options to be chained.
         */
        Builder commitStatusOptions(UnaryOperator<CallOptions> options);

        /**
         * Specify the default call options for chaincode events.
         * <p>A call of:</p>
         * <pre>{@code
         *     chaincodeEventsOptions(CallOption.deadlineAfter(5, TimeUnit.SECONDS))
         * }</pre>
         * <p>is equivalent to:</p>
         * <pre>{@code
         *     chaincodeEventsOptions(options -> options.withDeadlineAfter(5, TimeUnit.SECONDS))
         * }</pre>
         * @param options Call options.
         * @return The builder instance, allowing multiple configuration options to be chained.
         * @deprecated Replaced by {@link #chaincodeEventsOptions(UnaryOperator)}.
         */
        @Deprecated
        default Builder chaincodeEventsOptions(CallOption... options) {
            return chaincodeEventsOptions(GatewayUtils.asCallOptions(options));
        }

        /**
         * Specify the default call options for chaincode events.
         * @param options Function that transforms call options.
         * @return The builder instance, allowing multiple configuration options to be chained.
         */
        Builder chaincodeEventsOptions(UnaryOperator<CallOptions> options);

        /**
         * Specify the default call options for block events.
         * @param options Function that transforms call options.
         * @return The builder instance, allowing multiple configuration options to be chained.
         */
        Builder blockEventsOptions(UnaryOperator<CallOptions> options);

        /**
         * Specify the default call options for filtered block events.
         * @param options Function that transforms call options.
         * @return The builder instance, allowing multiple configuration options to be chained.
         */
        Builder filteredBlockEventsOptions(UnaryOperator<CallOptions> options);

        /**
         * Specify the default call options for block and private data events.
         * @param options Function that transforms call options.
         * @return The builder instance, allowing multiple configuration options to be chained.
         */
        Builder blockAndPrivateDataEventsOptions(UnaryOperator<CallOptions> options);

        /**
         * Connects to the gateway using the specified options.
         * @return The connected {@link Gateway} object.
         */
        Gateway connect();
    }
}
