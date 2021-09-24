/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scenario;

import java.security.GeneralSecurityException;
import java.util.Map;
import java.util.concurrent.Callable;

import com.google.protobuf.InvalidProtocolBufferException;
import org.hyperledger.fabric.client.Commit;
import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.Network;
import org.hyperledger.fabric.client.Proposal;
import org.hyperledger.fabric.client.Status;
import org.hyperledger.fabric.client.SubmittedTransaction;
import org.hyperledger.fabric.client.Transaction;
import org.hyperledger.fabric.client.identity.Signer;

import static org.assertj.core.api.Assertions.assertThat;

public final class TransactionInvocation {
    private final Network network;
    private final Contract contract;
    private final Proposal.Builder proposalBuilder;
    private Callable<byte[]> action;
    private Signer offlineSigner;
    private String response;
    private Throwable error;
    private long blockNumber;

    private TransactionInvocation(final Network network, final Contract contract, final String transactionName) {
        this.network = network;
        this.contract = contract;
        proposalBuilder = contract.newProposal(transactionName);
    }

    public void setTransient(final Map<String, String> transientData) {
        transientData.entrySet().forEach(entry -> proposalBuilder.putTransient(entry.getKey(), entry.getValue()));
    }

    public void setEndorsingOrgs(final String[] orgs) {
        proposalBuilder.setEndorsingOrganizations(orgs);
    }

    public static TransactionInvocation prepareToSubmit(final Network network, final Contract contract, final String transactionName) {
        TransactionInvocation invocation = new TransactionInvocation(network, contract, transactionName);
        invocation.action = invocation::submit;
        return invocation;
    }

    public static TransactionInvocation prepareToEvaluate(final Network network, final Contract contract, final String transactionName) {
        TransactionInvocation invocation = new TransactionInvocation(network, contract, transactionName);
        invocation.action = invocation::evaluate;
        return invocation;
    }

    public void setArguments(final String[] args) {
        proposalBuilder.addArguments(args);
    }

    public void setOfflineSigner(final Signer signer) {
        offlineSigner = signer;
    }

    public void invoke() {
        try {
            byte[] result = action.call();
            setResponse(result);
        } catch (Exception e) {
            setError(e);
        }
    }

    private byte[] submit() throws InvalidProtocolBufferException, GeneralSecurityException {
        Proposal unsignedProposal = proposalBuilder.build();
        Proposal signedProposal = offlineSign(unsignedProposal);

        Transaction unsignedTransaction = signedProposal.endorse();
        Transaction signedTransaction = offlineSign(unsignedTransaction);

        SubmittedTransaction submitted = signedTransaction.submitAsync();
        Commit commit = offlineSign(submitted);

        Status status = commit.getStatus();
        blockNumber = status.getBlockNumber();

        if (!status.isSuccessful()) {
            throw new RuntimeException("Transaction commit failed with status: " + status.getCode());
        }

        return submitted.getResult();
    }

    private Proposal offlineSign(final Proposal proposal) throws GeneralSecurityException, InvalidProtocolBufferException {
        if (null == offlineSigner) {
            return proposal;
        }

        byte[] signature = offlineSigner.sign(proposal.getDigest());
        return contract.newSignedProposal(proposal.getBytes(), signature);
    }

    private Transaction offlineSign(final Transaction transaction) throws GeneralSecurityException, InvalidProtocolBufferException {
        if (null == offlineSigner) {
            return transaction;
        }

        byte[] signature = offlineSigner.sign(transaction.getDigest());
        return contract.newSignedTransaction(transaction.getBytes(), signature);
    }

    private Commit offlineSign(final Commit commit) throws InvalidProtocolBufferException, GeneralSecurityException {
        if (null == offlineSigner) {
            return commit;
        }

        byte[] signature = offlineSigner.sign(commit.getDigest());
        return network.newSignedCommit(commit.getBytes(), signature);
    }

    private byte[] evaluate() throws InvalidProtocolBufferException, GeneralSecurityException {
        Proposal unsignedProposal = proposalBuilder.build();
        Proposal signedProposal = offlineSign(unsignedProposal);

        return signedProposal.evaluate();
    }

    private void setResponse(final byte[] response) {
        this.response = ScenarioSteps.newString(response);
        error = null;
    }

    private void setError(final Throwable error) {
        this.error = error;
        response = null;
    }

    public String getResponse() {
        assertThat(response)
                .withFailMessage(() -> "No transaction response. Failed with error: " + error)
                .isNotNull();
        return response;
    }

    public Throwable getError() {
        assertThat(error)
                .withFailMessage(() -> "No transaction error. Response was: " + response)
                .isNotNull();
        return error;
    }

    public long getBlockNumber() {
        return blockNumber;
    }
}
