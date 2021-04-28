/*
 * Copyright IBM Corp. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package com.example;

import java.io.FileReader;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.security.PrivateKey;
import java.security.cert.X509Certificate;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.function.Supplier;

import io.grpc.Channel;
import io.grpc.netty.shaded.io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.shaded.io.grpc.netty.NettyChannelBuilder;
import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.Gateway;
import org.hyperledger.fabric.client.Network;
import org.hyperledger.fabric.client.Transaction;
import org.hyperledger.fabric.client.identity.*;
import org.hyperledger.fabric.protos.peer.TransactionPackage;

public class Sample
{
    private static final String mspID     = "Org1MSP";
    private static final Path cryptoPath  = Paths.get("..","..", "scenario", "fixtures", "crypto-material", "crypto-config", "peerOrganizations", "org1.example.com");
    private static final Path certPath    = cryptoPath.resolve(Paths.get("users", "User1@org1.example.com", "msp", "signcerts", "User1@org1.example.com-cert.pem"));
    private static final Path keyPath     = cryptoPath.resolve(Paths.get("users", "User1@org1.example.com", "msp", "keystore", "key.pem"));
    private static final Path tlsCertPath = cryptoPath.resolve(Paths.get("peers", "peer0.org1.example.com", "tls", "ca.crt"));

    public static void main( String[] args ) throws Exception
    {
        // make a gRPC connection to the gateway peer
        X509Certificate tlsCert = Identities.readX509Certificate(new FileReader(tlsCertPath.toFile()));
        Channel channel = NettyChannelBuilder.forTarget("localhost:7051")
                .sslContext(GrpcSslContexts.forClient().trustManager(tlsCert).build())
                .overrideAuthority("peer0.org1.example.com")
                .build();

        X509Certificate certificate = Identities.readX509Certificate(new FileReader(certPath.toFile()));
        PrivateKey privateKey = Identities.readPrivateKey(new FileReader(keyPath.toFile()));
        Identity identity = new X509Identity(mspID, certificate);
        Signer signer = Signers.newPrivateKeySigner(privateKey);

        try (Gateway gateway = Gateway.newInstance()
                .identity(identity)
                .signer(signer)
                .connection(channel)
                .connect()) {
            Network network = gateway.getNetwork("mychannel");
            Contract contract = network.getContract("basic");

            String time = LocalDateTime.now().toString();

            // Submit a transaction, blocking until the transaction has been committed on the ledger.
            System.out.println("Submitting transaction to basic chaincode with value " + time + "...");
            byte[] result = contract.submitTransaction("put", "time", time);
            System.out.println("Submit result = " + new String(result));

            System.out.println("Evaluating query...");
            result = contract.evaluateTransaction("get", "time");
            System.out.println("Query result = " + new String(result));

            // Submit transaction asynchronously, allowing this thread to process the chaincode response (e.g. update a UI)
            // without waiting for the commit notification
            System.out.println("Submitting transaction asynchronously to basic chaincode with value " + time + "...");
            Transaction txn = contract.newProposal("put")
                    .addArguments("async", time)
                    .build()
                    .endorse();
            Supplier<TransactionPackage.TxValidationCode> commit = txn.submitAsync();
            System.out.println("Proposal result = " + new String(txn.getResult()));

            // wait for transactions to commit before querying the value
            TransactionPackage.TxValidationCode status = commit.get();
            if (status != TransactionPackage.TxValidationCode.VALID) {
                throw new Exception("Failed to commit transaction");
            }
            // Committed.  Check the value:
            result = contract.evaluateTransaction("get", "async");
            System.out.println("Transaction committed. Query result = " + new String(result));
        }
    }
}
