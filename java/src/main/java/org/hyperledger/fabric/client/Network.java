/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Iterator;

import com.google.protobuf.InvalidProtocolBufferException;

/**
 * The Network represents a Fabric network (channel). Network instances are obtained from a Gateway using the
 * {@link Gateway#getNetwork(String)} method.
 *
 * <p>The Network provides the ability for applications to:</p>
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
     * Get events emitted by transaction functions of a specific chaincode. Note that the returned {@link Iterator} may
     * throw {@link io.grpc.StatusRuntimeException} from any of its methods if a gRPC connection error occurs.
     * @param chaincodeId A chaincode ID.
     * @return Ordered sequence of events.
     */
    Iterator<ChaincodeEvent> getChaincodeEvents(String chaincodeId);

    /**
     * Create a chaincode events request, which can be used to obtain events emitted by transaction functions of a
     * specific chaincode. Supports off-line signing flow.
     * @param chaincodeId A chaincode ID.
     * @return A chaincode events request.
     */
    ChaincodeEventsSupplier newChaincodeEvents(String chaincodeId);

    /**
     * Create a chaincode events request with the specified digital signature, which can be used to obtain events
     * emitted by transaction functions of a specific chaincode. Supports off-line signing flow.
     * @param bytes Serialized chaincode events request.
     * @param signature Digital signature.
     * @return A signed chaincode events request.
     * @throws InvalidProtocolBufferException if the supplied chaincode events request bytes are not valid.
     */
    ChaincodeEventsSupplier newSignedChaincodeEvents(byte[] bytes, byte[] signature) throws InvalidProtocolBufferException;
}
