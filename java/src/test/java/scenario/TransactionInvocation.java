/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scenario;

import java.security.GeneralSecurityException;
import java.util.Map;
import java.util.concurrent.Callable;

import org.hyperledger.fabric.client.Commit;
import org.hyperledger.fabric.client.CommitStatusException;
import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.EndorseException;
import org.hyperledger.fabric.client.Gateway;
import org.hyperledger.fabric.client.GatewayException;
import org.hyperledger.fabric.client.Proposal;
import org.hyperledger.fabric.client.Status;
import org.hyperledger.fabric.client.SubmitException;
import org.hyperledger.fabric.client.SubmittedTransaction;
import org.hyperledger.fabric.client.Transaction;
import org.hyperledger.fabric.client.identity.Signer;

import static org.assertj.core.api.Assertions.assertThat;

public final class TransactionInvocation {
    private final Gateway gateway;
    private final Proposal.Builder proposalBuilder;
    private Callable<byte[]> action;
    private Signer offlineSigner;
    private String response;
    private Throwable error;
    private long blockNumber;

    private TransactionInvocation(final Gateway gateway, final Contract contract, final String transactionName) {
        this.gateway = gateway;
        proposalBuilder = contract.newProposal(transactionName);
    }

    public void setTransient(final Map<String, String> transientData) {
        transientData.forEach(proposalBuilder::putTransient);
    }

    public void setEndorsingOrgs(final String[] orgs) {
        proposalBuilder.setEndorsingOrganizations(orgs);
    }

    public static TransactionInvocation prepareToSubmit(final Gateway gateway, final Contract contract, final String transactionName) {
        TransactionInvocation invocation = new TransactionInvocation(gateway, contract, transactionName);
        invocation.action = invocation::submit;
        return invocation;
    }

    public static TransactionInvocation prepareToEvaluate(final Gateway gateway, final Contract contract, final String transactionName) {
        TransactionInvocation invocation = new TransactionInvocation(gateway, contract, transactionName);
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

    private byte[] submit() throws GeneralSecurityException, EndorseException, SubmitException, CommitStatusException {
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

    private Proposal offlineSign(final Proposal proposal) throws GeneralSecurityException {
        if (null == offlineSigner) {
            return proposal;
        }

        byte[] signature = offlineSigner.sign(proposal.getDigest());
        return gateway.newSignedProposal(proposal.getBytes(), signature);
    }

    private Transaction offlineSign(final Transaction transaction) throws GeneralSecurityException {
        if (null == offlineSigner) {
            return transaction;
        }

        byte[] signature = offlineSigner.sign(transaction.getDigest());
        return gateway.newSignedTransaction(transaction.getBytes(), signature);
    }

    private Commit offlineSign(final Commit commit) throws GeneralSecurityException {
        if (null == offlineSigner) {
            return commit;
        }

        byte[] signature = offlineSigner.sign(commit.getDigest());
        return gateway.newSignedCommit(commit.getBytes(), signature);
    }

    private byte[] evaluate() throws GeneralSecurityException, GatewayException {
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
