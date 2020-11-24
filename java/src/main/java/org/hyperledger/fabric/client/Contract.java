/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Optional;
import java.util.concurrent.TimeoutException;

/**
 * Represents a smart contract instance in a network.
 * Applications should get a Contract instance from a Network using the
 * {@link Network#getContract(String) getContract} method.
 *
 * <p>The Contract allows applications to:</p>
 * <ul>
 *     <li>Submit transactions that store state to the ledger using {@link #submitTransaction(String, String...)}.</li>
 *     <li>Evaluate transactions that query state from the ledger using {@link #evaluateTransaction(String, String...)}.</li>
 * </ul>
 *
 * @see <a href="https://hyperledger-fabric.readthedocs.io/en/release-2.2/developapps/application.html#construct-request">Developing Fabric Applications - Construct request</a>
 */
public interface Contract {
    /**
     * Submit a transaction to the ledger. The transaction function {@code name}
     * will be evaluated on the endorsing peers and then submitted to the ordering service
     * for committing to the ledger.
     * This function is equivalent to calling {@code createTransaction(name).submit()}.
     *
     * @param name Transaction function name.
     * @param args Transaction function arguments.
     * @return Payload response from the transaction function.
     * @throws ContractException if the transaction is rejected.
     * @throws TimeoutException If the transaction was successfully submitted to the orderer but
     * timed out before a commit event was received from peers.
     * @throws InterruptedException if the current thread is interrupted while waiting.
     * @throws GatewayRuntimeException if an underlying infrastructure failure occurs.
     *
     * @see <a href="https://hyperledger-fabric.readthedocs.io/en/release-2.2/developapps/application.html#submit-transaction">Developing Fabric Applications - Submit transaction</a>
     */
    byte[] submitTransaction(String name, String... args) throws ContractException, TimeoutException, InterruptedException;

    /**
     * Evaluate a transaction function and return its results.
     * The transaction function {@code name}
     * will be evaluated on the endorsing peers but the responses will not be sent to
     * the ordering service and hence will not be committed to the ledger.
     * This is used for querying the world state.
     * This function is equivalent to calling {@code createTransaction(name).evaluate()}.
     *
     * @param name Transaction function name.
     * @param args Transaction function arguments.
     * @return Payload response from the transaction function.
     * @throws ContractException if no peers are reachable or an error response is returned.
     */
    byte[] evaluateTransaction(String name, String... args) throws ContractException;

    Proposal newProposal(String transactionName);
    Proposal newSignedProposal(byte[] proposalBytes, byte[] signature);
    Transaction newSignedTransaction(byte[] transactionBytes, byte[] signature);
    String getChaincodeId();
    Optional<String> getContractName();
}
