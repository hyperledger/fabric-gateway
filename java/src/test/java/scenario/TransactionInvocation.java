/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scenario;

import java.util.Map;
import java.util.concurrent.Callable;

import org.hyperledger.fabric.client.ContractException;
import org.hyperledger.fabric.client.Proposal;

import static org.assertj.core.api.Assertions.assertThat;

public final class TransactionInvocation {
    private final Proposal.Builder proposalBuilder;
    private Callable<byte[]> action;
    private String response;
    private Throwable error;

    private TransactionInvocation(Proposal.Builder proposalBuilder) {
        this.proposalBuilder = proposalBuilder;
    }

    public void setTransient(Map<String, byte[]> transientData) {
        proposalBuilder.putAllTransient(transientData);
    }

    public static TransactionInvocation prepareToSubmit(Proposal.Builder proposalBuilder) {
        TransactionInvocation invocation = new TransactionInvocation(proposalBuilder);
        invocation.action = invocation::submit;
        return invocation;
    }

    public static TransactionInvocation prepareToEvaluate(Proposal.Builder proposalBuilder) {
        TransactionInvocation invocation = new TransactionInvocation(proposalBuilder);
        invocation.action = invocation::evaluate;
        return invocation;
    }

    public void setArguments(String[] args) {
        proposalBuilder.addArguments(args);
    }

    public void invoke() {
        try {
            byte[] result = action.call();
            setResponse(result);
        } catch (Exception e) {
            setError(e);
        }
    }

    private byte[] submit() throws ContractException {
        return proposalBuilder.build().endorse().submitSync();
    }

    private byte[] evaluate() {
        return proposalBuilder.build().evaluate();
    }

    private void setResponse(byte[] response) {
        this.response = ScenarioSteps.newString(response);
        error = null;
    }

    private void setError(Throwable error) {
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
