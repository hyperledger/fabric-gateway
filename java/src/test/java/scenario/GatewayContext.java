/*
 * Copyright 2021 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scenario;

import java.io.IOException;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.TimeUnit;

import io.grpc.ManagedChannel;
import org.hyperledger.fabric.client.ChaincodeEvent;
import org.hyperledger.fabric.client.Checkpointer;
import org.hyperledger.fabric.client.CloseableIterator;
import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.Gateway;
import org.hyperledger.fabric.client.InMemoryCheckpointer;
import org.hyperledger.fabric.client.Network;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.protos.common.Block;
import org.hyperledger.fabric.protos.peer.BlockAndPrivateData;
import org.hyperledger.fabric.protos.peer.FilteredBlock;

public class GatewayContext {
    private final Gateway.Builder gatewayBuilder;
    private final Map<String, EventListener<ChaincodeEvent>> chaincodeEventListeners = new HashMap<>();
    private final Map<String, EventListener<Block>> blockEventListeners = new HashMap<>();
    private final Map<String, EventListener<FilteredBlock>> filteredBlockEventListeners = new HashMap<>();
    private final Map<String, EventListener<BlockAndPrivateData>> blockAndPrivateDataEventListeners = new HashMap<>();
    private Checkpointer checkpointer;
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

    public void createCheckpointer() {
        checkpointer = new InMemoryCheckpointer();
    }

    public TransactionInvocation newTransaction(String action, String transactionName) {
        if (action.equals("submit")) {
            return TransactionInvocation.prepareToSubmit(gateway, contract, transactionName);
        } else {
            return TransactionInvocation.prepareToEvaluate(gateway, contract, transactionName);
        }
    }

    public void listenForChaincodeEvents(String listenerName, String chaincodeName) {
        receiveChaincodeEvents(listenerName, network.getChaincodeEvents(chaincodeName));
    }

    public void listenForChaincodeEventsUsingCheckpointer(String listenerName, String chaincodeName) {
        CloseableIterator<ChaincodeEvent> iter = network.newChaincodeEventsRequest(chaincodeName)
                .checkpoint(checkpointer)
                .build()
                .getEvents();
        receiveChaincodeEventsUsingCheckpointer(listenerName, iter);
    }

    public void replayChaincodeEvents(String listenerName, String chaincodeName, long startBlock) {
        CloseableIterator<ChaincodeEvent> iter = network.newChaincodeEventsRequest(chaincodeName)
                .startBlock(startBlock)
                .build()
                .getEvents();
        receiveChaincodeEvents(listenerName, iter);
    }

    private void receiveChaincodeEventsUsingCheckpointer(final String listenerName, final CloseableIterator<ChaincodeEvent> iter) {
        closeChaincodeEvents(listenerName);
        EventListener<ChaincodeEvent> e = new CheckpointEventListener<>(iter, checkpointer::checkpointChaincodeEvent);
        chaincodeEventListeners.put(listenerName, e);
    }

    private void receiveChaincodeEvents(final String listenerName, final CloseableIterator<ChaincodeEvent> iter) {
        closeChaincodeEvents(listenerName);
        chaincodeEventListeners.put(listenerName, new BasicEventListener<>(iter));
    }

    public ChaincodeEvent nextChaincodeEvent(String listenerName) throws InterruptedException, IOException {
        return chaincodeEventListeners.get(listenerName).next();
    }

    public void listenForBlockEvents(String listenerName) {
        receiveBlockEvents(listenerName, network.getBlockEvents());
    }

    public void listenForBlockEventsUsingCheckpointer(String listenerName) {
        CloseableIterator<Block> iter = network.newBlockEventsRequest()
                .checkpoint(checkpointer)
                .build()
                .getEvents();
        receiveBlockEventsUsingCheckpointer(listenerName, iter);
    }

    private void receiveBlockEventsUsingCheckpointer(final String listenerName, final CloseableIterator<Block> iter) {
        closeBlockEvents(listenerName);
        EventListener<Block> e = new CheckpointEventListener<>(iter, event-> checkpointer.checkpointBlock(event.getHeader().getNumber()));
        blockEventListeners.put(listenerName, e);
    }

    public void listenForFilteredBlockEventsUsingCheckpointer(String listenerName) {
        CloseableIterator<FilteredBlock> iter = network.newFilteredBlockEventsRequest()
                .checkpoint(checkpointer)
                .build()
                .getEvents();
        receiveFilteredBlockEventsUsingCheckpointer(listenerName, iter);
    }

    private void receiveFilteredBlockEventsUsingCheckpointer(final String listenerName, final CloseableIterator<FilteredBlock> iter) {
        closeFilteredBlockEvents(listenerName);
        EventListener<FilteredBlock> e = new CheckpointEventListener<>(iter, event-> checkpointer.checkpointBlock(event.getNumber()));
        filteredBlockEventListeners.put(listenerName, e);
    }

    public void  listenForBlockAndPrivateDataUsingCheckpointer(String listenerName) {
        CloseableIterator<BlockAndPrivateData> iter = network.newBlockAndPrivateDataEventsRequest()
                .checkpoint(checkpointer)
                .build()
                .getEvents();
        receiveBlockAndPrivateDataEventsUsingCheckpointer(listenerName, iter);
    }

    private void receiveBlockAndPrivateDataEventsUsingCheckpointer(final String listenerName, final CloseableIterator<BlockAndPrivateData> iter) {
        closeBlockAndPrivateDataEvents(listenerName);
        EventListener<BlockAndPrivateData> e = new CheckpointEventListener<>(iter, event-> checkpointer.checkpointBlock(event.getBlock().getHeader().getNumber()));
        blockAndPrivateDataEventListeners.put(listenerName, e);
    }

    public void replayBlockEvents(String listenerName, long startBlock) {
        CloseableIterator<Block> iter = network.newBlockEventsRequest()
                .startBlock(startBlock)
                .build()
                .getEvents();
        receiveBlockEvents(listenerName, iter);
    }

    private void receiveBlockEvents(final String listenerName, final CloseableIterator<Block> iter) {
        closeBlockEvents(listenerName);
        blockEventListeners.put(listenerName, new BasicEventListener<>(iter));
    }

    public Block nextBlockEvent(String listenerName) throws InterruptedException, IOException {
        return blockEventListeners.get(listenerName).next();
    }

    public void listenForFilteredBlockEvents(String listenerName) {
        receiveFilteredBlockEvents(listenerName, network.getFilteredBlockEvents());
    }

    public void replayFilteredBlockEvents(String listenerName, long startBlock) {
        CloseableIterator<FilteredBlock> iter = network.newFilteredBlockEventsRequest()
                .startBlock(startBlock)
                .build()
                .getEvents();
        receiveFilteredBlockEvents(listenerName, iter);
    }

    private void receiveFilteredBlockEvents(final String listenerName, final CloseableIterator<FilteredBlock> iter) {
        closeFilteredBlockEvents(listenerName);
        filteredBlockEventListeners.put(listenerName, new BasicEventListener<>(iter));
    }

    public FilteredBlock nextFilteredBlockEvent(String listenerName) throws InterruptedException, IOException {
        return filteredBlockEventListeners.get(listenerName).next();
    }

    public void listenForBlockAndPrivateDataEvents(String listenerName) {
        receiveBlockAndPrivateDataEvents(listenerName, network.getBlockAndPrivateDataEvents());
    }

    public void replayBlockAndPrivateDataEvents(String listenerName, long startBlock) {
        CloseableIterator<BlockAndPrivateData> iter = network.newBlockAndPrivateDataEventsRequest()
                .startBlock(startBlock)
                .build()
                .getEvents();
        receiveBlockAndPrivateDataEvents(listenerName, iter);
    }

    private void receiveBlockAndPrivateDataEvents(final String listenerName, final CloseableIterator<BlockAndPrivateData> iter) {
        closeBlockAndPrivateDataEvents(listenerName);
        blockAndPrivateDataEventListeners.put(listenerName, new BasicEventListener<>(iter));
    }

    public BlockAndPrivateData nextBlockAndPrivateDataEvent(String listenerName) throws InterruptedException, IOException {
        return blockAndPrivateDataEventListeners.get(listenerName).next();
    }

    public void close() {
        chaincodeEventListeners.values().forEach(EventListener::close);
        blockEventListeners.values().forEach(EventListener::close);
        filteredBlockEventListeners.values().forEach(EventListener::close);
        blockAndPrivateDataEventListeners.values().forEach(EventListener::close);

        if (gateway != null) {
            gateway.close();
        }

        closeChannel();
    }

    public void closeChaincodeEvents(String listenerName) {
        closeEventListener(chaincodeEventListeners, listenerName);
    }

    public void closeBlockEvents(String listenerName) {
        closeEventListener(blockEventListeners, listenerName);
    }

    public void closeFilteredBlockEvents(String listenerName) {
        closeEventListener(filteredBlockEventListeners, listenerName);
    }

    public void closeBlockAndPrivateDataEvents(String listenerName) {
        closeEventListener(blockAndPrivateDataEventListeners, listenerName);
    }

    private <T> void closeEventListener(final Map<String, EventListener<T>> listeners, final String listenerName) {
        listeners.computeIfPresent(listenerName, (name, listener) -> {
            listener.close();
            return null;
        });
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
