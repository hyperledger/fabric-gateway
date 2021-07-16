/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scenario;

import java.util.Iterator;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.LinkedTransferQueue;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TransferQueue;

import org.hyperledger.fabric.client.ChaincodeEvent;
import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.Gateway;
import org.hyperledger.fabric.client.Network;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;

import io.grpc.ManagedChannel;

public class GatewayContext {
    private final Gateway.Builder gatewayBuilder;
    private ManagedChannel channel;
    private Gateway gateway;
    private Network network;
    private Contract contract;
    private TransferQueue<ChaincodeEvent> chaincodeEventQueue;
    private CompletableFuture<Void> chaincodeEventJob;

    public GatewayContext(Identity identity) {
        gatewayBuilder = Gateway.newInstance()
                .identity(identity);
    }

    public GatewayContext(Identity identity, Signer signer) {
        gatewayBuilder = Gateway.newInstance()
                .identity(identity)
                .signer(signer);
    }

    public void connect(ManagedChannel channel) {
        this.channel = channel;
        gateway = gatewayBuilder.connection(channel).connect();
    }

    public void useNetwork(String networkName) {
        network = gateway.getNetwork(networkName);
    }

    public void useContract(String contractName) {
        contract = network.getContract(contractName);
    }

    public TransactionInvocation newTransaction(String action, String transactionName) {
        if (action.equals("submit")) {
            return TransactionInvocation.prepareToSubmit(network, contract, transactionName);
        } else {
            return TransactionInvocation.prepareToEvaluate(network, contract, transactionName);
        }
    }

    public void listenForChaincodeEvents(String chaincodeId) {
        cancelChaincodeEvents();

        Iterator<ChaincodeEvent> eventIter = network.getChaincodeEvents(chaincodeId);

        // The gRPC implementation in Java appears not to start reading events until hasNext() is first called on the
        // Iterator so immediately start asynchronously reading from it to avoid missing events
        chaincodeEventQueue = new LinkedTransferQueue<ChaincodeEvent>();
        chaincodeEventJob = CompletableFuture.runAsync(() -> {
            eventIter.forEachRemaining(event -> {
                try {
                    chaincodeEventQueue.transfer(event);
                } catch (InterruptedException e) {
                    throw new RuntimeException(e);
                }
            });
        });
    }

    public ChaincodeEvent nextChaincodeEvent() throws InterruptedException {
        return chaincodeEventQueue.poll(30, TimeUnit.SECONDS);
    }

    public void close() {
        if (gateway != null) {
            gateway.close();
        }
        cancelChaincodeEvents();
        closeChannel();
    }

    private void closeChannel() {
        if (channel.isShutdown()) {
            return;
        }

        channel.shutdownNow();
        try {
            channel.awaitTermination(5, TimeUnit.SECONDS);
        } catch (InterruptedException e) {
            // Ignore
        }
    }

    private void cancelChaincodeEvents() {
        if (chaincodeEventJob != null) {
            chaincodeEventJob.cancel(true);
            chaincodeEventJob = null;
        }
    }
}
