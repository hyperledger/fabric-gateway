/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import io.grpc.Channel;

import java.util.function.Function;

import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;

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
 * <pre><code>
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
 * </code></pre>
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
     * @throws GatewayRuntimeException if a configuration or infrastructure error causes a failure.
     */
    Network getNetwork(String networkName);

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
         * Specifies the Gateway endpoint address in the form {@code "host:post"}. The connection to the specified
         * endpoint will be closed when the Gateway instance is closed.
         * @param url Endpoint address.
         * @return The builder instance, allowing multiple configuration options to be chained.
         */
        Builder endpoint(String url);

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
         * Connects to the gateway using the specified options.
         * @return The connected {@link Gateway} object.
         */
        Gateway connect();
    }
}
