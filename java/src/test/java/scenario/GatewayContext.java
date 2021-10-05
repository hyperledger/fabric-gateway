/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scenario;

import java.util.concurrent.CompletableFuture;
import java.util.concurrent.LinkedTransferQueue;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.TransferQueue;

import io.grpc.ManagedChannel;
import org.hyperledger.fabric.client.ChaincodeEvent;
import org.hyperledger.fabric.client.CloseableIterator;
import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.Gateway;
import org.hyperledger.fabric.client.Network;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;

public class GatewayContext {
    private final Gateway.Builder gatewayBuilder;
    private ManagedChannel channel;
    private Gateway gateway;
    private Network network;
    private Contract contract;
    private TransferQueue<ChaincodeEvent> eventQueue;
    private CompletableFuture<Void> eventJob;
    private CloseableIterator<ChaincodeEvent> eventIter;

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
        CloseableIterator<ChaincodeEvent> iter = network.getChaincodeEvents(chaincodeId);
        receiveChaincodeEvents(iter);
    }

    public void replayChaincodeEvents(String chaincodeId, long startBlock) {
        CloseableIterator<ChaincodeEvent> iter = network.newChaincodeEventsRequest(chaincodeId)
                .startBlock(startBlock)
                .build()
                .getEvents();
        receiveChaincodeEvents(iter);
    }

    private void receiveChaincodeEvents(final CloseableIterator<ChaincodeEvent> iter) {
        closeChaincodeEvents();

        this.eventIter = iter;

        // Java gRPC implementation doesn't request events until the first read from iterator, so start reading
        // asynchronously immediately
        final TransferQueue<ChaincodeEvent> queue = new LinkedTransferQueue<>();
        eventQueue = queue;
        eventJob = CompletableFuture.runAsync(() -> iter.forEachRemaining(event -> {
            try {
                queue.transfer(event);
            } catch (InterruptedException e) {
                e.printStackTrace();
            }
        }));
    }

    public ChaincodeEvent nextChaincodeEvent() throws InterruptedException {
        return eventQueue.poll(30, TimeUnit.SECONDS);
    }

    public void close() {
        closeChaincodeEvents();

        if (gateway != null) {
            gateway.close();
        }

        closeChannel();
    }

    private void closeChaincodeEvents() {
        if (eventIter != null) {
            eventIter.close();
            eventIter = null;
        }

        if (eventJob != null) {
            eventJob.cancel(true);
            eventJob = null;
        }
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
}
