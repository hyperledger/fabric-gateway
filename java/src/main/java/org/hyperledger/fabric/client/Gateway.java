/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.io.IOException;

import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.client.impl.GatewayImpl;

/**
 * The Gateway provides the connection point for an application to access the Fabric network as a specific user. It is
 * instantiated from a Builder instance that is created using {@link #createBuilder()} and configured using a gateway URL
 * and a signing identity. It can then be connected to a fabric network using the
 * {@link Builder#connect()} method. Once connected, it can then access individual {@link Network} instances (channels)
 * using the {@link #getNetwork(String) getNetwork} method which in turn can access the {@link Contract} installed on a
 * network and {@link Contract#submitTransaction(String, String...) submit transactions} to the ledger.
 *
 * <p>Gateway instances should be reused for multiple transaction invocations and only closed once connection to the
 * Fabric network is no longer required.</p>
 *
 * <pre><code>
 *     Gateway.Builder builder = Gateway.createBuilder()
 *             .identity(identity)
 *             .networkConfig(url);
 *
 *     try (Gateway gateway = builder.connect()) {
 *         Network network = gateway.getNetwork("mychannel");
 *         // Interactions with the network
 *     }
 * </code></pre>
 */
public interface Gateway extends AutoCloseable {
    /**
     * Returns an object representing a network.
     *
     * @param networkName The name of the network (channel name)
     * @return {@link Network}
     * @throws GatewayRuntimeException if a configuration or infrastructure error causes a failure.
     */
    Network getNetwork(String networkName);

    /**
     * Get the client identity associated with the Gateway connection.
     * @return A client identity.
     */
    Identity getIdentity();

    /**
     * Get the signing implementation associated with the Gateway connection.
     * @return A signing implementation
     */
    Signer getSigner();

    /**
     * Creates a gateway builder which is used to configure the gateway options
     * prior to connecting to the Fabric network.
     *
     * @return A gateway connection.
     */
    static Builder createBuilder() {
        return new GatewayImpl.Builder();
    }

    /**
     * Close the gateway connection and all associated resources, including removing listeners attached to networks and
     * contracts created by the gateway.
     */
    void close();

    /**
     *
     * The Gateway Builder interface defines the options that can be configured
     * prior to connection.
     * An instance of builder is created using the static method
     * {@link Gateway#createBuilder()}.  Every method on the builder object will return
     * a reference to the same builder object allowing them to be chained together in
     * a single line.
     *
     */
    interface Builder {
        /**
         * Specifies the path to the common connection profile.
         * @param url The url of the gateway.
         * @return The builder instance, allowing multiple configuration options to be chained.
         * @throws IOException if the config file does not exist, or is not JSON or YAML format,
         * or contains invalid information.
         */
        Builder networkConfig(String url) throws IOException;

        /**
         * Specifies the identity that is to be used to connect to the network. All operations
         * under this gateway connection will be performed using this identity.
         * @param identity An identity
         * @return The builder instance, allowing multiple configuration options to be chained.
         */
        Builder identity(Identity identity);

        /**
         * Specify the signing implementation used to sign message sent to the Gateway.
         * @param signer A signing implementation.
         * @return The builder instance, allowing multiple configuration options to be chained.
         */
        Builder signer(Signer signer);

        /**
         * Connects to the gateway using the specified options.
         * @return The connected {@link Gateway} object.
         */
        Gateway connect();
    }
}
