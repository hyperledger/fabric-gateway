/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scenario;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.TimeUnit;

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
    private final Map<String, EventListener<ChaincodeEvent>> chaincodeEventListeners = new HashMap<>();
    private ManagedChannel channel;
    private Gateway gateway;
    private Network network;
    private Contract contract;

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

    public void listenForChaincodeEvents(String listenerName, String chaincodeName) {
        CloseableIterator<ChaincodeEvent> iter = network.getChaincodeEvents(chaincodeName);
        receiveChaincodeEvents(listenerName, iter);
    }

    public void replayChaincodeEvents(String listenerName, String chaincodeName, long startBlock) {
        CloseableIterator<ChaincodeEvent> iter = network.newChaincodeEventsRequest(chaincodeName)
                .startBlock(startBlock)
                .build()
                .getEvents();
        receiveChaincodeEvents(listenerName, iter);
    }

    private void receiveChaincodeEvents(final String listenerName, final CloseableIterator<ChaincodeEvent> iter) {
        closeChaincodeEvents(listenerName);
        EventListener<ChaincodeEvent> listener = new EventListener<>(iter);
        chaincodeEventListeners.put(listenerName, listener);
    }

    public ChaincodeEvent nextChaincodeEvent(String listenerName) throws InterruptedException {
        return chaincodeEventListeners.get(listenerName).next();
    }

    public void close() {
        chaincodeEventListeners.values().forEach(EventListener::close);

        if (gateway != null) {
            gateway.close();
        }

        closeChannel();
    }

    public void closeChaincodeEvents(String listenerName) {
        EventListener<?> listener = chaincodeEventListeners.get(listenerName);
        if (listener != null) {
            listener.close();
        }
    }

    private void closeChannel() {
        if (channel.isShutdown()) {
            return;
        }

        try {
            channel.shutdownNow().awaitTermination(5, TimeUnit.SECONDS);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }
}
