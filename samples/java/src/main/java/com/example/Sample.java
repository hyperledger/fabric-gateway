/*
 * Copyright IBM Corp. All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package com.example;

import java.io.IOException;
import java.io.Reader;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.security.InvalidKeyException;
import java.security.PrivateKey;
import java.security.cert.CertificateException;
import java.security.cert.X509Certificate;
import java.time.LocalDateTime;
import java.util.Arrays;

import io.grpc.ManagedChannel;
import io.grpc.netty.shaded.io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.shaded.io.grpc.netty.NettyChannelBuilder;
import org.hyperledger.fabric.client.CommitException;
import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.Gateway;
import org.hyperledger.fabric.client.Network;
import org.hyperledger.fabric.client.SubmittedTransaction;
import org.hyperledger.fabric.client.identity.Identities;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.client.identity.Signers;
import org.hyperledger.fabric.client.identity.X509Identity;
import org.hyperledger.fabric.protos.peer.TransactionPackage;

public class Sample {
    private static final String mspID     = "Org1MSP";
    private static final Path cryptoPath  = Paths.get("..","..", "scenario", "fixtures", "crypto-material", "crypto-config", "peerOrganizations", "org1.example.com");
    private static final Path certPath    = cryptoPath.resolve(Paths.get("users", "User1@org1.example.com", "msp", "signcerts", "User1@org1.example.com-cert.pem"));
    private static final Path keyPath     = cryptoPath.resolve(Paths.get("users", "User1@org1.example.com", "msp", "keystore", "key.pem"));
    private static final Path tlsCertPath = cryptoPath.resolve(Paths.get("peers", "peer0.org1.example.com", "tls", "ca.crt"));

    public static void main( String[] args ) throws Exception {
        // The gRPC client connection should be shared by all Gateway connections to this endpoint
        ManagedChannel channel = newGrpcConnection();

        Identity identity = newIdentity();
        Signer signer = newSigner();

        try (Gateway gateway = Gateway.newInstance()
                .identity(identity)
                .signer(signer)
                .connection(channel)
                .connect()) {
            exampleSubmit(gateway);
            System.out.println();

            exampleSubmitAsync(gateway);
            System.out.println();
        } finally {
            channel.shutdownNow();
        }
    }

    private static void exampleSubmit(Gateway gateway) throws CommitException {
        Network network = gateway.getNetwork("mychannel");
        Contract contract = network.getContract("basic");

        String timestamp = LocalDateTime.now().toString();
        System.out.println("Submitting \"put\" transaction with arguments: time, " + timestamp);

        // Submit a transaction, blocking until the transaction has been committed on the ledger.
        byte[] submitResult = contract.submitTransaction("put", "time", timestamp);

        System.out.println("Submit result: " + new String(submitResult, StandardCharsets.UTF_8));
        System.out.println("Evaluating \"get\" query with arguments: time");

        byte[] evaluateResult = contract.evaluateTransaction("get", "time");
        System.out.println("Query result: " + new String(evaluateResult, StandardCharsets.UTF_8));
    }

    private static void exampleSubmitAsync(Gateway gateway) {
        Network network = gateway.getNetwork("mychannel");
        Contract contract = network.getContract("basic");

        String timestamp = LocalDateTime.now().toString();
        System.out.println("Submitting \"put\" transaction asynchronously with arguments: async");

        // Submit transaction asynchronously, blocking until the transaction has been sent to the orderer, and allowing
        // this thread to process the chaincode response (e.g. update a UI) without waiting for the commit notification
        SubmittedTransaction commit = contract.newProposal("put")
                .addArguments("async", timestamp)
                .build()
                .endorse()
                .submitAsync();

        System.out.println("Submit result: " + new String(commit.getResult(), StandardCharsets.UTF_8));
        System.out.println("Waiting for transaction commit");

        if (!commit.isSuccessful()) {
            TransactionPackage.TxValidationCode status = commit.getStatus();
            throw new RuntimeException("Transaction " + commit.getTransactionId() +
                    " failed to commit with status code " + status.getNumber() + " (" + status.name() + ")");
        }

        System.out.println("Transaction committed successfully");
        System.out.println("Evaluating \"get\" query with arguments: async");

        byte[] evaluateResult = contract.evaluateTransaction("get", "async");
        System.out.println("Query result: " + new String(evaluateResult, StandardCharsets.UTF_8));
    }

    private static ManagedChannel newGrpcConnection() throws IOException, CertificateException {
        Reader tlsCertReader = Files.newBufferedReader(tlsCertPath);
        X509Certificate tlsCert = Identities.readX509Certificate(tlsCertReader);

        return NettyChannelBuilder.forTarget("localhost:7051")
                .sslContext(GrpcSslContexts.forClient().trustManager(tlsCert).build())
                .overrideAuthority("peer0.org1.example.com")
                .build();
    }

    private static Identity newIdentity() throws IOException, CertificateException {
        Reader certReader = Files.newBufferedReader(certPath);
        X509Certificate certificate = Identities.readX509Certificate(certReader);

        return new X509Identity(mspID, certificate);
    }

    private static Signer newSigner() throws IOException, InvalidKeyException {
        Reader keyReader = Files.newBufferedReader(keyPath);
        PrivateKey privateKey = Identities.readPrivateKey(keyReader);

        return Signers.newPrivateKeySigner(privateKey);
    }
}
