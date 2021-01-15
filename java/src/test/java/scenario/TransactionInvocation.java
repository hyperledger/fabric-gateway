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
import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.ContractException;
import org.hyperledger.fabric.client.Proposal;
import org.hyperledger.fabric.client.Transaction;
import org.hyperledger.fabric.client.identity.Signer;

import static org.assertj.core.api.Assertions.assertThat;

public final class TransactionInvocation {
    private final Contract contract;
    private final Proposal.Builder proposalBuilder;
    private Callable<byte[]> action;
    private Signer offlineSigner;
    private String response;
    private Throwable error;

    private TransactionInvocation(final Contract contract, final String transactionName) {
        this.contract = contract;
        proposalBuilder = contract.newProposal(transactionName);
    }

    public void setTransient(final Map<String, byte[]> transientData) {
        proposalBuilder.putAllTransient(transientData);
    }

    public static TransactionInvocation prepareToSubmit(final Contract contract, final String transactionName) {
        TransactionInvocation invocation = new TransactionInvocation(contract, transactionName);
        invocation.action = invocation::submit;
        return invocation;
    }

    public static TransactionInvocation prepareToEvaluate(final Contract contract, final String transactionName) {
        TransactionInvocation invocation = new TransactionInvocation(contract, transactionName);
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

    private byte[] submit() throws ContractException, InvalidProtocolBufferException, GeneralSecurityException {
        Proposal proposal = proposalBuilder.build();
        proposal = offlineSign(proposal);

        Transaction transaction = proposal.endorse();
        transaction = offlineSign(transaction);

        return transaction.submitSync();
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

    private byte[] evaluate() throws InvalidProtocolBufferException, GeneralSecurityException {
        Proposal proposal = proposalBuilder.build();
        proposal = offlineSign(proposal);

        return proposal.evaluate();
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
}
