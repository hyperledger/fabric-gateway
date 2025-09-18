/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import io.grpc.CallOptions;
import io.grpc.Channel;
import io.grpc.ClientCall;
import io.grpc.ClientInterceptor;
import io.grpc.MethodDescriptor;
import io.grpc.stub.AbstractStub;
import java.util.function.UnaryOperator;

final class DefaultCallOptions {
    private final UnaryOperator<CallOptions> evaluate;
    private final UnaryOperator<CallOptions> endorse;
    private final UnaryOperator<CallOptions> submit;
    private final UnaryOperator<CallOptions> commitStatus;
    private final UnaryOperator<CallOptions> chaincodeEvents;
    private final UnaryOperator<CallOptions> blockEvents;
    private final UnaryOperator<CallOptions> filteredBlockEvents;
    private final UnaryOperator<CallOptions> blockAndPrivateDataEvents;

    private DefaultCallOptions(final Builder builder) {
        evaluate = builder.evaluate;
        endorse = builder.endorse;
        submit = builder.submit;
        commitStatus = builder.commitStatus;
        chaincodeEvents = builder.chaincodeEvents;
        blockEvents = builder.blockEvents;
        filteredBlockEvents = builder.filteredBlockEvents;
        blockAndPrivateDataEvents = builder.blockAndPrivateDataEvents;
    }

    static Builder newBuiler() {
        return new Builder();
    }

    <T extends AbstractStub<T>> T applyEvaluate(final T stub, final UnaryOperator<CallOptions> additional) {
        return applyOptions(applyOptions(stub, additional), evaluate);
    }

    <T extends AbstractStub<T>> T applyEndorse(final T stub, final UnaryOperator<CallOptions> additional) {
        return applyOptions(applyOptions(stub, additional), endorse);
    }

    <T extends AbstractStub<T>> T applySubmit(final T stub, final UnaryOperator<CallOptions> additional) {
        return applyOptions(applyOptions(stub, additional), submit);
    }

    <T extends AbstractStub<T>> T applyCommitStatus(final T stub, final UnaryOperator<CallOptions> additional) {
        return applyOptions(applyOptions(stub, additional), commitStatus);
    }

    <T extends AbstractStub<T>> T applyChaincodeEvents(final T stub, final UnaryOperator<CallOptions> additional) {
        return applyOptions(applyOptions(stub, additional), chaincodeEvents);
    }

    <T extends AbstractStub<T>> T applyBlockEvents(final T stub, final UnaryOperator<CallOptions> additional) {
        return applyOptions(applyOptions(stub, additional), blockEvents);
    }

    <T extends AbstractStub<T>> T applyFilteredBlockEvents(final T stub, final UnaryOperator<CallOptions> additional) {
        return applyOptions(applyOptions(stub, additional), filteredBlockEvents);
    }

    <T extends AbstractStub<T>> T applyBlockAndPrivateDataEvents(
            final T stub, final UnaryOperator<CallOptions> additional) {
        return applyOptions(applyOptions(stub, additional), blockAndPrivateDataEvents);
    }

    private static <T extends AbstractStub<T>> T applyOptions(final T stub, final UnaryOperator<CallOptions> operator) {
        if (operator == null) {
            return stub;
        }

        return stub.withInterceptors(new ClientInterceptor() {
            @Override
            @SuppressWarnings("PMD.TypeParameterNamingConventions")
            public <ReqT, RespT> ClientCall<ReqT, RespT> interceptCall(
                    final MethodDescriptor<ReqT, RespT> methodDescriptor,
                    final CallOptions callOptions,
                    final Channel channel) {
                return channel.newCall(methodDescriptor, operator.apply(callOptions));
            }
        });
    }

    static final class Builder {
        private UnaryOperator<CallOptions> evaluate;
        private UnaryOperator<CallOptions> endorse;
        private UnaryOperator<CallOptions> submit;
        private UnaryOperator<CallOptions> commitStatus;
        private UnaryOperator<CallOptions> chaincodeEvents;
        private UnaryOperator<CallOptions> blockEvents;
        private UnaryOperator<CallOptions> filteredBlockEvents;
        private UnaryOperator<CallOptions> blockAndPrivateDataEvents;

        private Builder() {
            // Nothing to do
        }

        Builder evaluate(final UnaryOperator<CallOptions> options) {
            evaluate = options;
            return this;
        }

        Builder endorse(final UnaryOperator<CallOptions> options) {
            endorse = options;
            return this;
        }

        Builder submit(final UnaryOperator<CallOptions> options) {
            submit = options;
            return this;
        }

        Builder commitStatus(final UnaryOperator<CallOptions> options) {
            commitStatus = options;
            return this;
        }

        Builder chaincodeEvents(final UnaryOperator<CallOptions> options) {
            chaincodeEvents = options;
            return this;
        }

        Builder blockEvents(final UnaryOperator<CallOptions> options) {
            blockEvents = options;
            return this;
        }

        Builder filteredBlockEvents(final UnaryOperator<CallOptions> options) {
            filteredBlockEvents = options;
            return this;
        }

        Builder blockAndPrivateDataEvents(final UnaryOperator<CallOptions> options) {
            blockAndPrivateDataEvents = options;
            return this;
        }

        DefaultCallOptions build() {
            return new DefaultCallOptions(this);
        }
    }
}
