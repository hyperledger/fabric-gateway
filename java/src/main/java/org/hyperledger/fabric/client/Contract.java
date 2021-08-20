/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Optional;

import com.google.protobuf.InvalidProtocolBufferException;

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
 * and the resulting transaction object can be submitted to the orderer to be committed to the ledger.</p>
 *
 * <h3>Evaluate transaction example</h3>
 * <pre>{@code
 *     byte[] result = contract.newProposal("transactionName")
 *             .addArguments("one", "two")
 *             // Specify additional proposal options, such as transient data
 *             .build()
 *             .evaluate();
 * }</pre>
 *
 * <h3>Submit transaction example</h3>
 * <pre>{@code
 *     byte[] result = contract.newProposal("transactionName")
 *             .addArguments("one", "two")
 *             // Specify additional proposal options, such as transient data
 *             .build()
 *             .endorse()
 *             .submit();
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
 *     <li>On receipt of the serialized message and signature from the client, creating a signed proposal or transaction
 *     using the Contract's {@link #newSignedProposal(byte[], byte[])} or {@link #newSignedTransaction(byte[], byte[])}
 *     methods respectively,  or creating a signed commit using the Network's
 *     {@link Network#newSignedCommit(byte[], byte[])} method.</li>
 * </ol>
 *
 * <h3>Off-line signing examples</h3>
 * <p>Signing of a proposal that can then be evaluated or endorsed:</p>
 * <pre>{@code
 *     Proposal unsignedProposal = contract.newProposal("transactionName").build();
 *     byte[] proposalBytes = unsignedProposal.getBytes();
 *     byte[] proposalDigest = unsignedProposal.getDigest();
 *     // Generate signature from digest
 *     Proposal signedProposal = contract.newSignedProposal(proposalBytes, proposalSignature);
 * }</pre>
 *
 * <p>Signing of an endorsed transaction that can then be submitted to the orderer:</p>
 * <pre>{@code
 *     Transaction unsignedTransaction = signedProposal.endorse();
 *     byte[] transactionBytes = unsignedTransaction.getBytes();
 *     byte[] transactionDigest = unsignedTransaction.getDigest();
 *     // Generate signature from digest
 *     Transaction signedTransaction = contract.newSignedTransaction(transactionBytes, transactionSignature);
 * }</pre>
 *
 * <p>Signing of a commit that can be used to obtain the status of a submitted transaction:</p>
 * <pre>{@code
 *     Commit unsignedCommit = signedTransaction.submitAsync();
 *     byte[] commitBytes = unsignedCommit.getBytes();
 *     byte[] commitDigest = unsignedCommit.getDigest();
 *     // Generate signature from digest
 *     Commit signedCommit = network.newSignedCommit(commitBytes, commitDigest);
 * }</pre>
 *
 * @see <a href="https://hyperledger-fabric.readthedocs.io/en/latest/developapps/application.html#construct-request">Developing Fabric Applications - Construct request</a>
 */
public interface Contract {
    /**
     * Get the identifier of the chaincode that contains the smart contract.
     * @return Chaincode ID.
     */
    String getChaincodeId();

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
     * @param name Transaction function name.
     * @return Payload response from the transaction function.
     * @throws CommitException if the transaction fails to commit successfully.
     * @throws io.grpc.StatusRuntimeException if the gRPC service invocation fails.
     */
    byte[] submitTransaction(String name) throws CommitException;

    /**
     * Submit a transaction to the ledger and return its result only after it is committed to the ledger. The
     * transaction function will be evaluated on endorsing peers and then submitted to the ordering service to be
     * committed to the ledger.
     * @param name Transaction function name.
     * @param args Transaction function arguments.
     * @return Payload response from the transaction function.
     * @throws CommitException if the transaction fails to commit successfully.
     * @throws io.grpc.StatusRuntimeException if the gRPC service invocation fails.
     */
    byte[] submitTransaction(String name, String... args) throws CommitException;

    /**
     * Submit a transaction to the ledger and return its result only after it is committed to the ledger. The
     * transaction function will be evaluated on endorsing peers and then submitted to the ordering service to be
     * committed to the ledger.
     * @param name Transaction function name.
     * @param args Transaction function arguments.
     * @return Payload response from the transaction function.
     * @throws CommitException if the transaction fails to commit successfully.
     * @throws io.grpc.StatusRuntimeException if the gRPC service invocation fails.
     */
    byte[] submitTransaction(String name, byte[]... args) throws CommitException;

    /**
     * Evaluate a transaction function and return its results. A transaction proposal will be evaluated on endorsing
     * peers but the transaction will not be sent to the ordering service and so will not be committed to the ledger.
     * This can be used for querying the world state.
     * @param name Transaction function name.
     * @return Payload response from the transaction function.
     * @throws io.grpc.StatusRuntimeException if the gRPC service invocation fails.
     */
    byte[] evaluateTransaction(String name);

    /**
     * Evaluate a transaction function and return its results. A transaction proposal will be evaluated on endorsing
     * peers but the transaction will not be sent to the ordering service and so will not be committed to the ledger.
     * This can be used for querying the world state.
     * @param name Transaction function name.
     * @param args Transaction function arguments.
     * @return Payload response from the transaction function.
     * @throws io.grpc.StatusRuntimeException if the gRPC service invocation fails.
     */
    byte[] evaluateTransaction(String name, String... args);

    /**
     * Evaluate a transaction function and return its results. A transaction proposal will be evaluated on endorsing
     * peers but the transaction will not be sent to the ordering service and so will not be committed to the ledger.
     * This can be used for querying the world state.
     * @param name Transaction function name.
     * @param args Transaction function arguments.
     * @return Payload response from the transaction function.
     * @throws io.grpc.StatusRuntimeException if the gRPC service invocation fails.
     */
    byte[] evaluateTransaction(String name, byte[]... args);

    /**
     * Build a new proposal that can be evaluated or sent to peers for endorsement.
     * @param transactionName The name of the transaction to be invoked.
     * @return A proposal builder.
     */
    Proposal.Builder newProposal(String transactionName);

    /**
     * Create a proposal with the specified digital signature. Supports off-line signing flow.
     * @param proposalBytes The proposal.
     * @param signature A digital signature.
     * @return A signed proposal.
     * @throws InvalidProtocolBufferException if the supplied proposal bytes are not a valid proposal.
     */
    Proposal newSignedProposal(byte[] proposalBytes, byte[] signature) throws InvalidProtocolBufferException;

    /**
     * Create a transaction with the specified digital signature. Supports off-line signing flow.
     * @param transactionBytes The transaction.
     * @param signature A digital signature.
     * @return A signed transaction.
     * @throws InvalidProtocolBufferException if the supplied transaction bytes are not a valid transaction.
     */
    Transaction newSignedTransaction(byte[] transactionBytes, byte[] signature) throws InvalidProtocolBufferException;
}
