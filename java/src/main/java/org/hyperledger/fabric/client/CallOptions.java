/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Objects;

final class CallOptions {
    private final List<CallOption> evaluate;
    private final List<CallOption> endorse;
    private final List<CallOption> submit;
    private final List<CallOption> commitStatus;
    private final List<CallOption> chaincodeEvents;

    private CallOptions(final Builder builder) {
        this.evaluate = Collections.unmodifiableList(new ArrayList<>(builder.evaluate));
        this.endorse = Collections.unmodifiableList(new ArrayList<>(builder.endorse));
        this.submit = Collections.unmodifiableList(new ArrayList<>(builder.submit));
        this.commitStatus = Collections.unmodifiableList(new ArrayList<>(builder.commitStatus));
        this.chaincodeEvents = Collections.unmodifiableList(new ArrayList<>(builder.chaincodeEvents));
    }

    public static Builder newBuiler() {
        return new Builder();
    }

    private static List<CallOption> append(final List<CallOption> current, final CallOption... additional) {
        List<CallOption> result = new ArrayList<>(current);
        Collections.addAll(result, additional);
        return result;
    }

    public List<CallOption> getEvaluate(final CallOption... additional) {
        return append(evaluate, additional);
    }

    public List<CallOption> getEndorse(final CallOption... additional) {
        return append(endorse, additional);
    }

    public List<CallOption> getSubmit(final CallOption... additional) {
        return append(submit, additional);
    }

    public List<CallOption> getCommitStatus(final CallOption... additional) {
        return append(commitStatus, additional);
    }

    public List<CallOption> getChaincodeEvents(final CallOption... additional) {
        return append(chaincodeEvents, additional);
    }

    public static final class Builder {
        private List<CallOption> evaluate = Collections.emptyList();
        private List<CallOption> endorse = Collections.emptyList();
        private List<CallOption> submit = Collections.emptyList();
        private List<CallOption> commitStatus = Collections.emptyList();
        private List<CallOption> chaincodeEvents = Collections.emptyList();

        private Builder() {
            // Nothing to do
        }

        public Builder evaluate(final List<CallOption> options) {
            Objects.requireNonNull(options, "evaluate");
            evaluate = options;
            return this;
        }

        public Builder endorse(final List<CallOption> options) {
            Objects.requireNonNull(options, "endorse");
            endorse = options;
            return this;
        }

        public Builder submit(final List<CallOption> options) {
            Objects.requireNonNull(options, "submit");
            submit = options;
            return this;
        }

        public Builder commitStatus(final List<CallOption> options) {
            Objects.requireNonNull(options, "commitStatus");
            commitStatus = options;
            return this;
        }

        public Builder chaincodeEvents(final List<CallOption> options) {
            Objects.requireNonNull(options, "chaincodeEvents");
            chaincodeEvents = options;
            return this;
        }

        public CallOptions build() {
            return new CallOptions(this);
        }
    }
}
