/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Optional;

/**
 * Represents a smart contract instance in a network.
 * Applications should get a Contract instance from a Network using the
 * {@link Network#getContract(String) getContract} method.
 *
 * <p>The Contract allows applications to:</p>
 * <ul>
 *     <li>Evaluate transactions that query state from the ledger using {@link #evaluateTransaction(String, String...)}
 *     or {@link #evaluateTransaction(String, byte[][]) evaluateTransaction(String, byte[]...)}.</li>
 *     <li>Submit transactions that store state to the ledger using {@link #submitTransaction(String, String...)} or
 *     {@link #submitTransaction(String, byte[][]) submitTransaction(String, byte[]...)}.</li>
 * </ul>
 *
 * <p>For more complex transaction invocations, such as including transient data, the transaction proposal can be built
 * using {@link #newProposal(String)}. Once built, the proposal can either be evaluated, or can be sent for endorsement
 * and the resulting transaction object can be submitted to the orderer to be committed to the ledger. This flow can
 * also be used to asynchronously submit transactions, which allows the transaction result to be accessed prior to its
 * commit to the ledger.</p>
 *
 * <p>Evaluate transaction example:</p>
 * <pre>{@code
 *     byte[] result = contract.newProposal("transactionName")
 *             .addArguments("one", "two")
 *             // Specify additional proposal options, such as transient data
 *             .build()
 *             .evaluate();
 * }</pre>
 *
 * <p>Submit transaction example:</p>
 * <pre>{@code
 *     byte[] result = contract.newProposal("transactionName")
 *             .addArguments("one", "two")
 *             // Specify additional proposal options, such as transient data
 *             .build()
 *             .endorse()
 *             .submit();
 * }</pre>
 *
 * <p>Async submit example</p>
 * <pre>{@code
 *     SubmittedTransaction commit = contract.newProposal("transactionName")
 *             .build()
 *             .endorse()
 *             .submitAsync();
 *     byte[] result = commit.getResult();
 *
 *     // Update UI or reply to REST request before waiting for commit status
 *
 *     Status status = commit.getStatus();
 *     if (!status.isSuccessful()) {
 *         // Commit failure
 *     }
 * }</pre>
 *
 * <h2>Off-line signing</h2>
 *
 * <p>By default, proposal and transaction messages will be signed using the signing implementation specified when
 * connecting the Gateway. In cases where an external client holds the signing credentials, a signing implementation
 * can be omitted when connecting the Gateway and off-line signing can be carried out by:</p>
 * <ol>
 *     <li>Returning the serialized proposal, transaction or commit status message along with its digest to the client
 *     for them to generate a signature.</li>
 *     <li>On receipt of the serialized message and signature from the client, creating a signed proposal, transaction
 *     or commit using the Gateway's {@link Gateway#newSignedProposal(byte[], byte[])},
 *     {@link Gateway#newSignedTransaction(byte[], byte[])} or {@link Gateway#newSignedCommit(byte[], byte[])} methods
 *     respectively.</li>
 * </ol>
 *
 * <p>Signing of a proposal that can then be evaluated or endorsed:</p>
 * <pre>{@code
 *     Proposal unsignedProposal = contract.newProposal("transactionName").build();
 *     byte[] proposalBytes = unsignedProposal.getBytes();
 *     byte[] proposalDigest = unsignedProposal.getDigest();
 *     // Generate signature from digest
 *     Proposal signedProposal = gateway.newSignedProposal(proposalBytes, proposalSignature);
 * }</pre>
 *
 * <p>Signing of an endorsed transaction that can then be submitted to the orderer:</p>
 * <pre>{@code
 *     Transaction unsignedTransaction = signedProposal.endorse();
 *     byte[] transactionBytes = unsignedTransaction.getBytes();
 *     byte[] transactionDigest = unsignedTransaction.getDigest();
 *     // Generate signature from digest
 *     Transaction signedTransaction = gateway.newSignedTransaction(transactionBytes, transactionSignature);
 * }</pre>
 *
 * <p>Signing of a commit that can be used to obtain the status of a submitted transaction:</p>
 * <pre>{@code
 *     Commit unsignedCommit = signedTransaction.submitAsync();
 *     byte[] commitBytes = unsignedCommit.getBytes();
 *     byte[] commitDigest = unsignedCommit.getDigest();
 *     // Generate signature from digest
 *     Commit signedCommit = gateway.newSignedCommit(commitBytes, commitDigest);
 *
 *     byte[] result = signedTransaction.getResult();
 *     Status status = signedCommit.getStatus();
 * }</pre>
 */
public interface Contract {
    /**
     * Get the name of the chaincode that contains the smart contract.
     * @return Chaincode name.
     */
    String getChaincodeName();

    /**
     * Get the name of the smart contract within the chaincode. An empty value indicates that this Contract refers to
     * the chaincode's default smart contract.
     * @return An empty optional for the default smart contract; otherwise the contract name.
     */
    Optional<String> getContractName();

    /**
     * Submit a transaction to the ledger and return its result only after it is committed to the ledger. The
     * transaction function will be evaluated on endorsing peers and then submitted to the ordering service to be
     * committed to the ledger.
     *
     * <p>This method is equivalent to:</p>
     * <pre>{@code
     *     contract.newProposal(name)
     *             .build()
     *             .submit();
     * }</pre>
     * @param name Transaction function name.
     * @return Payload response from the transaction function.
     * @throws EndorseException if the endorse invocation fails.
     * @throws SubmitException if the submit invocation fails.
     * @throws CommitStatusException if the commit status invocation fails.
     * @throws CommitException if the transaction commits unsuccessfully.
     * @throws NullPointerException if the transaction name is null.
     */
    byte[] submitTransaction(String name) throws EndorseException, CommitException, SubmitException, CommitStatusException;

    /**
     * Submit a transaction to the ledger and return its result only after it is committed to the ledger. The
     * transaction function will be evaluated on endorsing peers and then submitted to the ordering service to be
     * committed to the ledger.
     * <p>This method is equivalent to:</p>
     * <pre>{@code
     *     contract.newProposal(name)
     *             .addArguments(arg1, arg2)
     *             .build()
     *             .submit();
     * }</pre>
     * @param name Transaction function name.
     * @param args Transaction function arguments.
     * @return Payload response from the transaction function.
     * @throws EndorseException if the endorse invocation fails.
     * @throws SubmitException if the submit invocation fails.
     * @throws CommitStatusException if the commit status invocation fails.
     * @throws CommitException if the transaction commits unsuccessfully.
     * @throws NullPointerException if the transaction name is null.
     */
    byte[] submitTransaction(String name, String... args) throws EndorseException, SubmitException, CommitStatusException, CommitException;

    /**
     * Submit a transaction to the ledger and return its result only after it is committed to the ledger. The
     * transaction function will be evaluated on endorsing peers and then submitted to the ordering service to be
     * committed to the ledger.
     *
     * <p>This method is equivalent to:</p>
     * <pre>{@code
     *     contract.newProposal(name)
     *             .addArguments(arg1, arg2)
     *             .build()
     *             .submit();
     * }</pre>
     * @param name Transaction function name.
     * @param args Transaction function arguments.
     * @return Payload response from the transaction function.
     * @throws EndorseException if the endorse invocation fails.
     * @throws SubmitException if the submit invocation fails.
     * @throws CommitStatusException if the commit status invocation fails.
     * @throws CommitException if the transaction commits unsuccessfully.
     * @throws NullPointerException if the transaction name is null.
     */
    byte[] submitTransaction(String name, byte[]... args) throws EndorseException, CommitException, SubmitException, CommitStatusException;

    /**
     * Evaluate a transaction function and return its results. A transaction proposal will be evaluated on endorsing
     * peers but the transaction will not be sent to the ordering service and so will not be committed to the ledger.
     * This can be used for querying the world state.
     *
     * <p>This method is equivalent to:</p>
     * <pre>{@code
     *     contract.newProposal(name)
     *             .build()
     *             .evaluate();
     * }</pre>
     * @param name Transaction function name.
     * @return Payload response from the transaction function.
     * @throws GatewayException if the gRPC service invocation fails.
     * @throws NullPointerException if the transaction name is null.
     */
    byte[] evaluateTransaction(String name) throws GatewayException;

    /**
     * Evaluate a transaction function and return its results. A transaction proposal will be evaluated on endorsing
     * peers but the transaction will not be sent to the ordering service and so will not be committed to the ledger.
     * This can be used for querying the world state.
     *
     * <p>This method is equivalent to:</p>
     * <pre>{@code
     *     contract.newProposal(name)
     *             .addArguments(arg1, arg2)
     *             .build()
     *             .evaluate();
     * }</pre>
     * @param name Transaction function name.
     * @param args Transaction function arguments.
     * @return Payload response from the transaction function.
     * @throws GatewayException if the gRPC service invocation fails.
     * @throws NullPointerException if the transaction name is null.
     */
    byte[] evaluateTransaction(String name, String... args) throws GatewayException;

    /**
     * Evaluate a transaction function and return its results. A transaction proposal will be evaluated on endorsing
     * peers but the transaction will not be sent to the ordering service and so will not be committed to the ledger.
     * This can be used for querying the world state.
     * <p>This method is equivalent to:</p>
     * <pre>{@code
     *     contract.newProposal(name)
     *             .addArguments(arg1, arg2)
     *             .build()
     *             .evaluate();
     * }</pre>
     * @param name Transaction function name.
     * @param args Transaction function arguments.
     * @return Payload response from the transaction function.
     * @throws GatewayException if the gRPC service invocation fails.
     * @throws NullPointerException if the transaction name is null.
     */
    byte[] evaluateTransaction(String name, byte[]... args) throws GatewayException;

    /**
     * Build a new proposal that can be evaluated or sent to peers for endorsement. Supports both asynchronous submit
     * of transactions and off-line signing flow.
     * @param transactionName The name of the transaction to be invoked.
     * @return A proposal builder.
     * @throws NullPointerException if the transaction name is null.
     */
    Proposal.Builder newProposal(String transactionName);
}
