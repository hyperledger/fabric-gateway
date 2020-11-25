/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package org.hyperledger.fabric.client.impl;

import java.nio.charset.StandardCharsets;
import java.security.GeneralSecurityException;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.security.SecureRandom;
import java.time.Instant;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import com.google.protobuf.ByteString;
import com.google.protobuf.Timestamp;
import org.bouncycastle.util.encoders.Hex;
import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.GatewayRuntimeException;
import org.hyperledger.fabric.client.Proposal;
import org.hyperledger.fabric.client.Transaction;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.gateway.PreparedTransaction;
import org.hyperledger.fabric.gateway.ProposedTransaction;
import org.hyperledger.fabric.gateway.Result;
import org.hyperledger.fabric.protos.common.Common;
import org.hyperledger.fabric.protos.msp.Identities;
import org.hyperledger.fabric.protos.peer.Chaincode;
import org.hyperledger.fabric.protos.peer.ProposalPackage;

public class ProposalImpl implements Proposal {
    private final Contract contract;
    private final NetworkImpl network;
    private final GatewayImpl gateway;
    private final String transactionName;
    private final List<byte[]> arguments = new ArrayList<>();
    private Map<String, byte[]> transientData;
    private static final int NONCE_LENGTH = 24;
    private static final SecureRandom RANDOM = new SecureRandom();

    ProposalImpl(ContractImpl contract, String transactionName) {
        this.contract = contract;
        this.network = contract.getNetwork();
        this.gateway = network.getGateway();
        this.transactionName = transactionName;
    }

    private String getQualifiedTransactionName() {
        return contract.getContractName()
                .map(contractName -> contractName + ":" + transactionName)
                .orElse(transactionName);
    }

    @Override
    public String getTransactionId() {
        return null;
    }

    @Override
    public byte[] getBytes() {
        return new byte[0];
    }

    @Override
    public byte[] getHash() {
        return new byte[0];
    }

    @Override
    public Proposal addArguments(final byte[]... args) {
        arguments.addAll(Arrays.asList(args));
        return this;
    }

    @Override
    public Proposal addArguments(final String... args) {
        if(args != null) {
            List<byte[]> byteArgs = Arrays.stream(args)
                    .map(arg -> arg.getBytes(StandardCharsets.UTF_8))
                    .collect(Collectors.toList());
            arguments.addAll(byteArgs);
        }
        return this;
    }

    @Override
    public Proposal setTransient(final Map<String, byte[]> transientData) {
        this.transientData = new HashMap<>(transientData);
        return this;
    }

    @Override
    public byte[] evaluate() {
        try {
            ProposedTransaction proposedTransaction = createProposedTransaction();

            Result result = gateway.getStub().evaluate(proposedTransaction);
            if (result != null && result.getValue() != null) {
                return result.getValue().toByteArray();
            }
            return new byte[0];
        } catch (GeneralSecurityException e) {
            throw new GatewayRuntimeException(e);
        }
    }

    @Override
    public Transaction endorse() {
        try {
            ProposedTransaction proposedTransaction = createProposedTransaction();
            PreparedTransaction preparedTransaction = gateway.getStub().endorse(proposedTransaction);
            return new TransactionImpl(gateway, preparedTransaction);
        } catch (GeneralSecurityException e) {
            throw new GatewayRuntimeException(e);
        }
    }

    private ProposedTransaction createProposedTransaction() throws GeneralSecurityException {
        ProposalPackage.Proposal proposal = createProposal();
        ProposalPackage.SignedProposal signedProposal = signProposal(proposal);
        ProposedTransaction pt = ProposedTransaction.newBuilder()
                .setProposal(signedProposal)
                .build();
        return pt;
    }

    public static byte[] generateNonce() {
        byte[] values = new byte[NONCE_LENGTH];
        RANDOM.nextBytes(values);
        return values;
    }

    private ProposalPackage.Proposal createProposal() throws NoSuchAlgorithmException {
        Identity id = gateway.getIdentity();
        Identities.SerializedIdentity creator = Identities.SerializedIdentity.newBuilder()
                .setMspid(id.getMspId())
                .setIdBytes(ByteString.copyFrom(id.getCredentials()))
                .build();

        byte[] nonce = generateNonce();
        MessageDigest digest = MessageDigest.getInstance("SHA-256");
        digest.update(nonce);
        digest.update(creator.toByteArray());
        byte[] hash = digest.digest();
        String txID = new String(Hex.encode(hash));

        Chaincode.ChaincodeID chaincodeID = Chaincode.ChaincodeID.newBuilder().setName(contract.getChaincodeId()).build();
        ProposalPackage.ChaincodeHeaderExtension chaincodeHeaderExtension = ProposalPackage.ChaincodeHeaderExtension.newBuilder()
                .setChaincodeId(chaincodeID)
                .build();
        Common.ChannelHeader channelHeader = Common.ChannelHeader.newBuilder()
                .setType(Common.HeaderType.ENDORSER_TRANSACTION.getNumber())
                .setTxId(txID)
                .setTimestamp(Timestamp.newBuilder().setSeconds(Instant.now().getEpochSecond()).build())
                .setChannelId(network.getName())
                .setExtension(chaincodeHeaderExtension.toByteString())
                .setEpoch(0)
                .build();
        Common.SignatureHeader signatureHeader = Common.SignatureHeader.newBuilder()
                .setCreator(creator.toByteString())
                .setNonce(ByteString.copyFrom(nonce))
                .build();
        Common.Header header = Common.Header.newBuilder()
                .setChannelHeader(channelHeader.toByteString())
                .setSignatureHeader(signatureHeader.toByteString())
                .build();
        Chaincode.ChaincodeInput.Builder cib = Chaincode.ChaincodeInput.newBuilder();
        cib.addArgs(ByteString.copyFrom(transactionName.getBytes()));
        for (byte[] arg : arguments) {
            cib.addArgs(ByteString.copyFrom(arg));
        }
        Chaincode.ChaincodeSpec chaincodeSpec = Chaincode.ChaincodeSpec.newBuilder()
                .setType(Chaincode.ChaincodeSpec.Type.NODE)
                .setChaincodeId(chaincodeID)
                .setInput(cib.build())
                .build();
        Chaincode.ChaincodeInvocationSpec chaincodeInvocationSpec = Chaincode.ChaincodeInvocationSpec.newBuilder()
                .setChaincodeSpec(chaincodeSpec)
                .build();
        ProposalPackage.ChaincodeProposalPayload.Builder cppb = ProposalPackage.ChaincodeProposalPayload.newBuilder()
                .setInput(chaincodeInvocationSpec.toByteString());
        if(transientData != null) {
            transientData.forEach((String key, byte[] value) -> {
                cppb.putTransientMap(key, ByteString.copyFrom(value));
            });
        }
        ProposalPackage.Proposal proposal = ProposalPackage.Proposal.newBuilder()
                .setHeader(header.toByteString())
                .setPayload(cppb.build().toByteString())
                .build();

        return proposal;
    }

    private ProposalPackage.SignedProposal signProposal(ProposalPackage.Proposal proposal) throws GeneralSecurityException {
        ByteString payload = proposal.toByteString();
        byte[] hash = Hash.sha256(payload.toByteArray());
        byte[] signature = gateway.getSigner().sign(hash);
        ProposalPackage.SignedProposal signedProposal = ProposalPackage.SignedProposal.newBuilder()
                .setProposalBytes(payload)
                .setSignature(ByteString.copyFrom(signature))
                .build();
        return signedProposal;
    }
}
