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
import java.util.Iterator;
import java.util.concurrent.TimeUnit;

import io.grpc.ManagedChannel;
import io.grpc.netty.shaded.io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.shaded.io.grpc.netty.NettyChannelBuilder;
import org.hyperledger.fabric.client.ChaincodeEvent;
import org.hyperledger.fabric.client.CloseableIterator;
import org.hyperledger.fabric.client.CommitException;
import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.Gateway;
import org.hyperledger.fabric.client.Network;
import org.hyperledger.fabric.client.Status;
import org.hyperledger.fabric.client.SubmittedTransaction;
import org.hyperledger.fabric.client.identity.Identities;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.client.identity.Signers;
import org.hyperledger.fabric.client.identity.X509Identity;

public class Sample {
    private static final String mspID     = "Org1MSP";
    private static final Path cryptoPath  = Paths.get("..","..", "scenario", "fixtures", "crypto-material", "crypto-config", "peerOrganizations", "org1.example.com");
    private static final Path certPath    = cryptoPath.resolve(Paths.get("users", "User1@org1.example.com", "msp", "signcerts", "User1@org1.example.com-cert.pem"));
    private static final Path keyPath     = cryptoPath.resolve(Paths.get("users", "User1@org1.example.com", "msp", "keystore", "key.pem"));
    private static final Path tlsCertPath = cryptoPath.resolve(Paths.get("peers", "peer0.org1.example.com", "tls", "ca.crt"));

    public static void main( String[] args ) throws Exception {
        // The gRPC client connection should be shared by all Gateway connections to this endpoint
        ManagedChannel channel = newGrpcConnection();

        Gateway.Builder builder = Gateway.newInstance()
                .identity(newIdentity())
                .signer(newSigner())
                .connection(channel);

        try (Gateway gateway = builder.connect()) {
            System.out.println("exampleSubmit:");
            exampleSubmit(gateway);
            System.out.println();

            System.out.println("exampleSubmitAsync:");
            exampleSubmitAsync(gateway);
            System.out.println();

            System.out.println("exampleSubmitPrivateData:");
            exampleSubmitPrivateData(gateway);
            System.out.println();

            System.out.println("exampleSubmitPrivateData2:");
            exampleSubmitPrivateData2(gateway);
            System.out.println();

            System.out.println("exampleStateBasedEndorsement:");
            exampleStateBasedEndorsement(gateway);
            System.out.println();

            System.out.println("exampleChaincodeEvents:");
            exampleChaincodeEvents(gateway);
            System.out.println();
        } catch (Throwable e) {
            e.printStackTrace();
        } finally {
            channel.shutdownNow();
            try {
                channel.awaitTermination(5, TimeUnit.SECONDS);
            } catch (InterruptedException e) {
                // Ignore
            }
        }

        System.exit(0);
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

    private static void exampleSubmitAsync(Gateway gateway) throws CommitException {
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

        Status status = commit.getStatus();
        if (!status.isSuccessful()) {
            throw new CommitException(commit.getTransactionId(), status.getCode());
        }

        System.out.println("Transaction committed successfully");
        System.out.println("Evaluating \"get\" query with arguments: async");

        byte[] evaluateResult = contract.evaluateTransaction("get", "async");
        System.out.println("Query result: " + new String(evaluateResult, StandardCharsets.UTF_8));
    }

    private static void exampleSubmitPrivateData(Gateway gateway) throws CommitException {
        Network network = gateway.getNetwork("mychannel");
        Contract contract = network.getContract("private");

        String timestamp = LocalDateTime.now().toString();
        System.out.println("Submitting \"WritePrivateData\" transaction with private data: " + timestamp);

        // Submit transaction, blocking until the transaction has been committed on the ledger.
        // The 'transient' data will not get written to the ledger, and is used to send sensitive data to the trusted endorsing peers.
        // The gateway will only send this to peers that are included in the ownership policy of all collections accessed by the chaincode function.
        // It is assumed that the gateway's organization is trusted and will invoke the chaincode to work out if extra endorsements are required from other orgs.
        // In this example, it will also seek endorsement from Org3, which is included in the ownership policy of both collections.
        contract.newProposal("WritePrivateData")
                .putTransient("collection", "SharedCollection,Org3Collection") // SharedCollection owned by Org1 & Org3, Org3Collection owned by Org3.
                .putTransient("key", "my-private-key")
                .putTransient("value", timestamp)
                .build()
                .endorse()
                .submit();

        System.out.println("Evaluating \"ReadPrivateData\" query with arguments: \"SharedCollection\", \"my-private-key\"");

        byte[] evaluateResult = contract.evaluateTransaction("ReadPrivateData", "SharedCollection", "my-private-key");
        System.out.println("Query result: " + new String(evaluateResult, StandardCharsets.UTF_8));
    }

    private static void exampleSubmitPrivateData2(Gateway gateway) throws CommitException {
        Network network = gateway.getNetwork("mychannel");
        Contract contract = network.getContract("private");

        String timestamp = LocalDateTime.now().toString();
        System.out.println("Submitting \"WritePrivateData\" transaction with private data: " + timestamp);

        // This example is similar to the previous private data example.
        // The difference here is that the gateway cannot assume that Org3 is trusted to receive transient data
        // that might be destined for storage in Org1Collection, since Org3 is not in its ownership policy.
        // The client application must explicitly specify which organizations must endorse using the setEndorsingOrganizations() function.
        contract.newProposal("WritePrivateData")
                .putTransient("collection", "Org1Collection,Org3Collection") // Org1Collection owned by Org1, Org3Collection owned by Org3.
                .putTransient("key", "my-private-key2")
                .putTransient("value", timestamp)
                .setEndorsingOrganizations("Org1MSP", "Org3MSP")
                .build()
                .endorse()
                .submit();

        System.out.println("Evaluating \"ReadPrivateData\" query with arguments: \"Org1Collection\", \"my-private-key2\"");

        byte[] evaluateResult = contract.evaluateTransaction("ReadPrivateData", "Org1Collection", "my-private-key2");
        System.out.println("Query result: " + new String(evaluateResult, StandardCharsets.UTF_8));
    }

    private static void exampleStateBasedEndorsement(Gateway gateway) throws CommitException {
        Network network = gateway.getNetwork("mychannel");
        Contract contract = network.getContract("private");

        System.out.println("Submitting \"SetStateWithEndorser\" transaction with arguments:  \"sbe-key\", \"value1\", \"Org1MSP\"");
        // Submit a transaction, blocking until the transaction has been committed on the ledger.
        contract.submitTransaction("SetStateWithEndorser", "sbe-key", "value1", "Org1MSP");

        // Query the current state
        System.out.println("Evaluating \"GetState\" query with arguments: \"sbe-key\"");
        byte[] evaluateResult = contract.evaluateTransaction("GetState", "sbe-key");
        System.out.println("Query result: " + new String(evaluateResult, StandardCharsets.UTF_8));

        // Submit transaction to modify the state.
        System.out.println("Submitting \"ChangeState\" transaction with arguments:  \"sbe-key\", \"value2\"");
        // Submit a transaction, blocking until the transaction has been committed on the ledger.
        contract.submitTransaction("ChangeState", "sbe-key", "value2");

        // Verify the current state
        System.out.println("Evaluating \"GetState\" query with arguments: \"sbe-key\"");
        evaluateResult = contract.evaluateTransaction("GetState", "sbe-key");
        System.out.println("Query result: " + new String(evaluateResult, StandardCharsets.UTF_8));

        // Now change the state-based endorsement policy for this state.
        System.out.println("Submitting \"SetStateEndorsers\" transaction with arguments:  \"sbe-key\", \"Org2MSP\", \"Org3MSP\"");
        // Submit a transaction, blocking until the transaction has been committed on the ledger.
        contract.submitTransaction("SetStateEndorsers", "sbe-key", "Org2MSP", "Org3MSP");

        // Modify the state.  It will now require endorsement from Org2 and Org3 for this transaction to succeed.
        // The gateway will endorse this transaction proposal on one of its organization's peers and will determine if
        // extra endorsements are required to satisfy any state changes.
        // In this example, it will seek endorsements from Org2 and Org3 in order to satisfy the SBE policy.
        System.out.println("Submitting \"ChangeState\" transaction with arguments:  \"sbe-key\", \"value3\"");
        // Submit a transaction, blocking until the transaction has been committed on the ledger.
        contract.submitTransaction("ChangeState", "sbe-key", "value3");

        // Verify the new state
        System.out.println("Evaluating \"GetState\" query with arguments: \"sbe-key\"");
        evaluateResult = contract.evaluateTransaction("GetState", "sbe-key");
        System.out.println("Query result: " + new String(evaluateResult, StandardCharsets.UTF_8));
    }

    private static void exampleChaincodeEvents(Gateway gateway) throws CommitException {
        Network network = gateway.getNetwork("mychannel");
        Contract contract = network.getContract("basic");

        // Submit a transaction that generates a chaincode event
        System.out.println("Submitting \"event\" transaction with arguments:  \"my-event-name\", \"my-event-payload\"");
        Status status = contract.newProposal("event")
                .addArguments("my-event-name", "my-event-payload")
                .build()
                .endorse()
                .submitAsync()
                .getStatus();
        if (!status.isSuccessful()) {
            throw new CommitException(status.getTransactionId(), status.getCode());
        }

        long blockNumber = status.getBlockNumber();

        System.out.println("Read chaincode events starting at block number " + blockNumber);
        try (CloseableIterator<ChaincodeEvent> events = network.newChaincodeEventsRequest("basic")
                .startBlock(blockNumber)
                .build()
                .getEvents()) {
            ChaincodeEvent event = events.next();
            System.out.println("Received event name: " + event.getEventName() +
                    ", payload: " + new String(event.getPayload(), StandardCharsets.UTF_8) +
                    ", txId: " + event.getTransactionId());
        }

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
