/*
 * Copyright 2019 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client;

import com.google.protobuf.ByteString;
import com.google.protobuf.InvalidProtocolBufferException;
import io.grpc.CallOptions;
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

import java.util.Objects;
import java.util.function.Function;
import java.util.function.UnaryOperator;

final class GatewayImpl implements Gateway {
    public static final class Builder implements Gateway.Builder {
        private static final Signer UNDEFINED_SIGNER = (digest) -> {
            throw new UnsupportedOperationException("No signing implementation supplied");
        };

        private Channel grpcChannel;
        private Identity identity;
        private Signer signer = UNDEFINED_SIGNER; // No signer implementation is required if only offline signing is used
        private Function<byte[], byte[]> hash = Hash.SHA256;
        private ByteString tlsCertificateHash = ByteString.empty();
        private final DefaultCallOptions.Builder optionsBuilder = DefaultCallOptions.newBuiler();

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
        public Builder tlsClientCertificateHash(final byte[] certificateHash) {
            Objects.requireNonNull(certificateHash, "certificateHash");
            tlsCertificateHash = ByteString.copyFrom(certificateHash);
            return this;
        }

        @Override
        public Builder evaluateOptions(final UnaryOperator<CallOptions> options) {
            Objects.requireNonNull(options, "evaluateOptions");
            optionsBuilder.evaluate(options);
            return this;
        }

        @Override
        public Builder endorseOptions(final UnaryOperator<CallOptions> options) {
            Objects.requireNonNull(options, "endorseOptions");
            optionsBuilder.endorse(options);
            return this;
        }

        @Override
        public Gateway.Builder submitOptions(final UnaryOperator<CallOptions> options) {
            Objects.requireNonNull(options, "submitOptions");
            optionsBuilder.submit(options);
            return this;
        }

        @Override
        public Gateway.Builder commitStatusOptions(final UnaryOperator<CallOptions> options) {
            Objects.requireNonNull(options, "commitStatusOptions");
            optionsBuilder.commitStatus(options);
            return this;
        }

        @Override
        public Gateway.Builder chaincodeEventsOptions(final UnaryOperator<CallOptions> options) {
            Objects.requireNonNull(options, "chaincodeEventsOptions");
            optionsBuilder.chaincodeEvents(options);
            return this;
        }

        @Override
        public Gateway.Builder blockEventsOptions(final UnaryOperator<CallOptions> options) {
            Objects.requireNonNull(options, "blockEventsOptions");
            optionsBuilder.blockEvents(options);
            return this;
        }

        @Override
        public Gateway.Builder filteredBlockEventsOptions(final UnaryOperator<CallOptions> options) {
            Objects.requireNonNull(options, "filteredBlockEventsOptions");
            optionsBuilder.filteredBlockEvents(options);
            return this;
        }

        @Override
        public Gateway.Builder blockAndPrivateDataEventsOptions(final UnaryOperator<CallOptions> options) {
            Objects.requireNonNull(options, "blockAndPrivateDataEventsOptions");
            optionsBuilder.blockAndPrivateDataEvents(options);
            return this;
        }

        @Override
        public GatewayImpl connect() {
            return new GatewayImpl(this);
        }
    }

    private final GatewayClient client;
    private final SigningIdentity signingIdentity;
    private final ByteString tlsCertificateHash;

    private GatewayImpl(final Builder builder) {
        signingIdentity = new SigningIdentity(builder.identity, builder.hash, builder.signer);
        client = new GatewayClient(builder.grpcChannel, builder.optionsBuilder.build());
        tlsCertificateHash = builder.tlsCertificateHash;
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
        return new NetworkImpl(client, signingIdentity, networkName, tlsCertificateHash);
    }

    @Override
    public Proposal newSignedProposal(final byte[] bytes, final byte[] signature) {
        ProposalImpl result = newProposal(bytes);
        result.setSignature(signature);
        return result;
    }

    @Override
    public ProposalImpl newProposal(final byte[] bytes) {
        try {
            ProposedTransaction proposedTransaction = ProposedTransaction.parseFrom(bytes);
            org.hyperledger.fabric.protos.peer.Proposal proposal =
                    org.hyperledger.fabric.protos.peer.Proposal.parseFrom(proposedTransaction.getProposal().getProposalBytes());
            Header header = Header.parseFrom(proposal.getHeader());
            ChannelHeader channelHeader = ChannelHeader.parseFrom(header.getChannelHeader());

            return new ProposalImpl(client, signingIdentity, channelHeader.getChannelId(), proposedTransaction);

        } catch (InvalidProtocolBufferException e) {
            throw new IllegalArgumentException(e);
        }
    }

    @Override
    public Transaction newSignedTransaction(final byte[] bytes, final byte[] signature) {
        TransactionImpl transaction = newTransaction(bytes);
        transaction.setSignature(signature);
        return transaction;
    }

    @Override
    public TransactionImpl newTransaction(final byte[] bytes) {
        try {
            PreparedTransaction preparedTransaction = PreparedTransaction.parseFrom(bytes);
            return new TransactionImpl(client, signingIdentity, preparedTransaction);
        } catch (InvalidProtocolBufferException e) {
            throw new IllegalArgumentException(e);
        }
    }

    @Override
    public Commit newSignedCommit(final byte[] bytes, final byte[] signature) {
        CommitImpl commit = newCommit(bytes);
        commit.setSignature(signature);
        return commit;
    }

    @Override
    public CommitImpl newCommit(final byte[] bytes) {
        try {
            SignedCommitStatusRequest signedRequest = SignedCommitStatusRequest.parseFrom(bytes);
            CommitStatusRequest request = CommitStatusRequest.parseFrom(signedRequest.getRequest());

            return new CommitImpl(client, signingIdentity, request.getTransactionId(), signedRequest);
        } catch (InvalidProtocolBufferException e) {
            throw new IllegalArgumentException(e);
        }
    }

    @Override
    public ChaincodeEventsRequest newSignedChaincodeEventsRequest(final byte[] bytes, final byte[] signature) {
        ChaincodeEventsRequestImpl result = newChaincodeEventsRequest(bytes);
        result.setSignature(signature);
        return result;
    }

    @Override
    public ChaincodeEventsRequestImpl newChaincodeEventsRequest(final byte[] bytes) {
        try {
            SignedChaincodeEventsRequest signedRequest = SignedChaincodeEventsRequest.parseFrom(bytes);

            return new ChaincodeEventsRequestImpl(client, signingIdentity, signedRequest);
        } catch (InvalidProtocolBufferException e) {
            throw new IllegalArgumentException(e);
        }
    }

    @Override
    public BlockEventsRequest newSignedBlockEventsRequest(final byte[] bytes, final byte[] signature) {
        BlockEventsRequestImpl result = newBlockEventsRequest(bytes);
        result.setSignature(signature);
        return result;
    }

    @Override
    public BlockEventsRequestImpl newBlockEventsRequest(final byte[] bytes) {
        try {
            Envelope request = Envelope.parseFrom(bytes);

            return new BlockEventsRequestImpl(client, signingIdentity, request);
        } catch (InvalidProtocolBufferException e) {
            throw new IllegalArgumentException(e);
        }
    }

    @Override
    public FilteredBlockEventsRequest newSignedFilteredBlockEventsRequest(final byte[] bytes, final byte[] signature) {
        FilteredBlockEventsRequestImpl result = newFilteredBlockEventsRequest(bytes);
        result.setSignature(signature);
        return result;
    }

    @Override
    public FilteredBlockEventsRequestImpl newFilteredBlockEventsRequest(final byte[] bytes) {
        try {
            Envelope request = Envelope.parseFrom(bytes);

            return new FilteredBlockEventsRequestImpl(client, signingIdentity, request);
        } catch (InvalidProtocolBufferException e) {
            throw new IllegalArgumentException(e);
        }
    }

    @Override
    public BlockAndPrivateDataEventsRequest newSignedBlockAndPrivateDataEventsRequest(final byte[] bytes, final byte[] signature) {
        BlockAndPrivateDataEventsRequestImpl result = newBlockAndPrivateDataEventsRequest(bytes);
        result.setSignature(signature);
        return result;
    }

    @Override
    public BlockAndPrivateDataEventsRequestImpl newBlockAndPrivateDataEventsRequest(final byte[] bytes) {
        try {
            Envelope request = Envelope.parseFrom(bytes);

            return new BlockAndPrivateDataEventsRequestImpl(client, signingIdentity, request);
        } catch (InvalidProtocolBufferException e) {
            throw new IllegalArgumentException(e);
        }
    }
}
