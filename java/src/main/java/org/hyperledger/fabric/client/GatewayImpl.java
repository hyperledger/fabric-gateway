/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import java.util.Arrays;
import java.util.Objects;
import java.util.function.Function;

import com.google.protobuf.InvalidProtocolBufferException;
import io.grpc.Channel;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.protos.common.ChannelHeader;
import org.hyperledger.fabric.protos.common.Envelope;
import org.hyperledger.fabric.protos.common.Header;
import org.hyperledger.fabric.protos.gateway.CommitStatusRequest;
import org.hyperledger.fabric.protos.gateway.PreparedTransaction;
import org.hyperledger.fabric.protos.gateway.ProposedTransaction;
import org.hyperledger.fabric.protos.gateway.SignedChaincodeEventsRequest;
import org.hyperledger.fabric.protos.gateway.SignedCommitStatusRequest;

final class GatewayImpl implements Gateway {
    public static final class Builder implements Gateway.Builder {
        private static final Signer UNDEFINED_SIGNER = (digest) -> {
            throw new UnsupportedOperationException("No signing implementation supplied");
        };

        private Channel grpcChannel;
        private Identity identity;
        private Signer signer = UNDEFINED_SIGNER; // No signer implementation is required if only offline signing is used
        private Function<byte[], byte[]> hash = Hash::sha256;
        private final CallOptions.Builder optionsBuilder = CallOptions.newBuiler();

        @Override
        public Builder connection(final Channel grpcChannel) {
            Objects.requireNonNull(grpcChannel, "connection");
            this.grpcChannel = grpcChannel;
            return this;
        }

        @Override
        public Builder identity(final Identity identity) {
            Objects.requireNonNull(identity, "identity");
            this.identity = identity;
            return this;
        }

        @Override
        public Builder signer(final Signer signer) {
            Objects.requireNonNull(signer, "signer");
            this.signer = signer;
            return this;
        }

        @Override
        public Builder hash(final Function<byte[], byte[]> hash) {
            Objects.requireNonNull(hash, "hash");
            this.hash = hash;
            return this;
        }

        @Override
        public Builder evaluateOptions(final CallOption... options) {
            optionsBuilder.evaluate(Arrays.asList(options));
            return this;
        }

        @Override
        public Builder endorseOptions(final CallOption... options) {
            optionsBuilder.endorse(Arrays.asList(options));
            return this;
        }

        @Override
        public Builder submitOptions(final CallOption... options) {
            optionsBuilder.submit(Arrays.asList(options));
            return this;
        }

        @Override
        public Builder commitStatusOptions(final CallOption... options) {
            optionsBuilder.commitStatus(Arrays.asList(options));
            return this;
        }

        @Override
        public Builder chaincodeEventsOptions(final CallOption... options) {
            optionsBuilder.chaincodeEvents(Arrays.asList(options));
            return this;
        }

        @Override
        public Gateway.Builder blockEventsOptions(final CallOption... options) {
            optionsBuilder.blockEvents(Arrays.asList(options));
            return this;
        }

        @Override
        public Gateway.Builder filteredBlockEventsOptions(final CallOption... options) {
            optionsBuilder.filteredBlockEvents(Arrays.asList(options));
            return this;
        }

        @Override
        public Gateway.Builder blockAndPrivateDataEventsOptions(final CallOption... options) {
            optionsBuilder.blockAndPrivateDataEvents(Arrays.asList(options));
            return this;
        }

        @Override
        public GatewayImpl connect() {
            return new GatewayImpl(this);
        }
    }

    private final GatewayClient client;
    private final SigningIdentity signingIdentity;

    private GatewayImpl(final Builder builder) {
        signingIdentity = new SigningIdentity(builder.identity, builder.hash, builder.signer);
        client = new GatewayClient(builder.grpcChannel, builder.optionsBuilder.build());
    }

    @Override
    public Identity getIdentity() {
        return this.signingIdentity.getIdentity();
    }

    @Override
    public void close() {
    }

    @Override
    public Network getNetwork(final String networkName) {
        return new NetworkImpl(client, signingIdentity, networkName);
    }

    @Override
    public Proposal newSignedProposal(final byte[] bytes, final byte[] signature) {
        try {
            ProposedTransaction proposedTransaction = ProposedTransaction.parseFrom(bytes);
            org.hyperledger.fabric.protos.peer.Proposal proposal = org.hyperledger.fabric.protos.peer.Proposal
                    .parseFrom(proposedTransaction.getProposal().getProposalBytes());
            Header header = Header.parseFrom(proposal.getHeader());
            ChannelHeader channelHeader = ChannelHeader.parseFrom(header.getChannelHeader());

            ProposalImpl result = new ProposalImpl(client, signingIdentity, channelHeader.getChannelId(),
                    proposedTransaction);
            result.setSignature(signature);
            return result;
        } catch (InvalidProtocolBufferException e) {
            throw new IllegalArgumentException(e);
        }
    }

    @Override
    public Transaction newSignedTransaction(final byte[] bytes, final byte[] signature) {
        try {
            PreparedTransaction preparedTransaction = PreparedTransaction.parseFrom(bytes);
            TransactionImpl transaction = new TransactionImpl(client, signingIdentity, preparedTransaction);
            transaction.setSignature(signature);
            return transaction;
        } catch (InvalidProtocolBufferException e) {
            throw new IllegalArgumentException(e);
        }
    }

    @Override
    public Commit newSignedCommit(final byte[] bytes, final byte[] signature) {
        try {
            SignedCommitStatusRequest signedRequest = SignedCommitStatusRequest.parseFrom(bytes);
            CommitStatusRequest request = CommitStatusRequest.parseFrom(signedRequest.getRequest());

            CommitImpl commit = new CommitImpl(client, signingIdentity, request.getTransactionId(), signedRequest);
            commit.setSignature(signature);
            return commit;
        } catch (InvalidProtocolBufferException e) {
            throw new IllegalArgumentException(e);
        }
    }

    @Override
    public ChaincodeEventsRequest newSignedChaincodeEventsRequest(final byte[] bytes, final byte[] signature) {
        try {
            SignedChaincodeEventsRequest signedRequest = SignedChaincodeEventsRequest.parseFrom(bytes);

            ChaincodeEventsRequestImpl result = new ChaincodeEventsRequestImpl(client, signingIdentity, signedRequest);
            result.setSignature(signature);
            return result;
        } catch (InvalidProtocolBufferException e) {
            throw new IllegalArgumentException(e);
        }
    }

    @Override
    public BlockEventsRequest newSignedBlockEventsRequest(final byte[] bytes, final byte[] signature) {
        try {
            Envelope request = Envelope.parseFrom(bytes);

            BlockEventsRequestImpl result = new BlockEventsRequestImpl(client, signingIdentity, request);
            result.setSignature(signature);
            return result;
        } catch (InvalidProtocolBufferException e) {
            throw new IllegalArgumentException(e);
        }
    }

    @Override
    public FilteredBlockEventsRequest newSignedFilteredBlockEventsRequest(final byte[] bytes, final byte[] signature) {
        try {
            Envelope request = Envelope.parseFrom(bytes);

            FilteredBlockEventsRequestImpl result = new FilteredBlockEventsRequestImpl(client, signingIdentity,
                    request);
            result.setSignature(signature);
            return result;
        } catch (InvalidProtocolBufferException e) {
            throw new IllegalArgumentException(e);
        }
    }

    @Override
    public BlockAndPrivateDataEventsRequest newSignedBlockAndPrivateDataEventsRequest(final byte[] bytes,
            final byte[] signature) {
        try {
            Envelope request = Envelope.parseFrom(bytes);

            BlockAndPrivateDataEventsRequestImpl result = new BlockAndPrivateDataEventsRequestImpl(client,
                    signingIdentity, request);
            result.setSignature(signature);
            return result;
        } catch (InvalidProtocolBufferException e) {
            throw new IllegalArgumentException(e);
        }
    }
}
