/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scenario;

import java.util.Map;
import java.util.concurrent.Callable;
import java.util.function.Consumer;

import org.hyperledger.fabric.client.Proposal;

import static org.assertj.core.api.Assertions.assertThat;

public final class TransactionInvocation {
    private final Proposal proposal;
    private final boolean expectSuccess;
    private String response;
    private Throwable error;
    private Consumer<String[]> action;
    private String[] args = new String[0];

    public static TransactionInvocation expectFail(Proposal proposal) {
        return new TransactionInvocation(proposal, false);
    }

    public static TransactionInvocation expectSuccess(Proposal proposal) {
        return new TransactionInvocation(proposal, true);
    }

    private TransactionInvocation(Proposal proposal, boolean expectSuccess) {
        this.proposal = proposal;
        this.expectSuccess = expectSuccess;
    }

    public void setTransient(Map<String, byte[]> transientData) {
        proposal.putAllTransient(transientData);
    }

    public static TransactionInvocation prepareToSubmit(Proposal proposal) {
        TransactionInvocation invocation = new TransactionInvocation(proposal, true);
        invocation.action = invocation::submit;
        return invocation;
    }

    public static TransactionInvocation prepareToEvaluate(Proposal proposal) {
        TransactionInvocation invocation = new TransactionInvocation(proposal, true);
        invocation.action = invocation::evaluate;
        return invocation;
    }

    public void setArgs(String[] args) {
        this.args = args;
    }

    public void invokeTxn() {
        action.accept(args);
    }

    public void submit(String... args) {
        invoke(() -> proposal.addArguments(args).endorse().submitSync());
    }

    public void evaluate(String... args) {
        invoke(() -> proposal.addArguments(args).evaluate());
    }

    private void invoke(Callable<byte[]> invocationFn) {
        try {
            byte[] result = invocationFn.call();
            setResponse(result);
        } catch (Exception e) {
            setError(e);
        }
    }

    private void setResponse(byte[] response) {
        String text = ScenarioSteps.newString(response);
        assertThat(expectSuccess)
                .withFailMessage("Response received for transaction that was expected to fail: %s", text)
                .isTrue();
        this.response = text;
        this.error = null;
    }

    private void setError(Throwable error) {
        if(expectSuccess) {
            error.printStackTrace();
        }
        assertThat(expectSuccess)
                .withFailMessage("Error received for transaction that was expected to succeed: %s", error)
                .isFalse();
        this.response = null;
        this.error = error;
    }

    public String getResponse() {
        assertThat(response)
                .withFailMessage("No transaction response")
                .isNotNull();
        return response;
    }

    public Throwable getError() {
        assertThat(error)
                .withFailMessage("No transaction error")
                .isNotNull();
        return error;
    }
}
