/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;


/**
 * A Network object represents the set of peers in a Fabric network (channel).
 * Applications should get a Network instance from a Gateway using the
 * {@link Gateway#getNetwork(String)} method.
 *
 * <p>The Network object provides the ability for applications to:</p>
 * <ul>
 *     <li>Obtain a specific smart contract deployed to the network using {@link #getContract(String)}, in order to
 *     submit and evaluate transactions for that smart contract.</li>
 * </ul>
 *
 * @see <a href="https://hyperledger-fabric.readthedocs.io/en/release-1.4/developapps/application.html#network-channel">Developing Fabric Applications - Network Channel</a>
 */
public interface Network {
    /**
     * Get an instance of a contract on the current network.
     * @param chaincodeId The name of the chaincode that implements the smart contract.
     * @return The contract object.
     */
    Contract getContract(String chaincodeId);

    /**
     * Get an instance of a contract on the current network.  If the chaincode instance contains more
     * than one smart contract class (available using the latest chaincode programming model), then an
     * individual class can be selected.
     * @param chaincodeId The name of the chaincode that implements the smart contract.
     * @param name The class name of the smart contract within the chaincode.
     * @return The contract object.
     */
    Contract getContract(String chaincodeId, String name);

    /**
     * Get a reference to the owning Gateway connection.
     * @return The owning gateway.
     */
    Gateway getGateway();

    /**
     * Get the name of the network channel.
     * @return The network name.
     */
    String getName();

}
