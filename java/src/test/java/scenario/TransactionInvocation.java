/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scenario;

import java.util.Map;
import java.util.concurrent.Callable;

import org.hyperledger.fabric.client.Proposal;

import static org.assertj.core.api.Assertions.assertThat;

public final class TransactionInvocation {
    private final Proposal proposal;
    private final boolean expectSuccess;
    private String response;
    private Throwable error;
    private String action = null;
    private String[] args = null;

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
        proposal.setTransient(transientData);
    }

    public static TransactionInvocation prepareToSubmit(Proposal proposal) {
        TransactionInvocation ti = new TransactionInvocation(proposal, true);
        ti.action = "submit";
        return ti;
    }

    public static TransactionInvocation prepareToEvaluate(Proposal proposal) {
        TransactionInvocation ti = new TransactionInvocation(proposal, true);
        ti.action = "evaluate";
        return ti;
    }

    public void setArgs(String[] args) {
        this.args = args;
    }

    public void invokeTxn() {
        if (action.equals("submit")) {
            submit(args);
        } else if (action.equals("evaluate")) {
            evaluate(args);
        }
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
