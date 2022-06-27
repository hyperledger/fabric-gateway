/*
 * Copyright 2020 IBM All Rights Reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scenario;

import com.google.gson.Gson;
import com.google.gson.JsonElement;
import com.google.gson.JsonParser;
import io.cucumber.datatable.DataTable;
import io.cucumber.docstring.DocString;
import io.cucumber.java.After;
import io.cucumber.java.en.Given;
import io.cucumber.java.en.Then;
import io.cucumber.java.en.When;
import io.grpc.ManagedChannel;
import io.grpc.Status;
import io.grpc.netty.shaded.io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.shaded.io.grpc.netty.NettyChannelBuilder;
import io.grpc.netty.shaded.io.netty.handler.ssl.SslContext;
import org.hyperledger.fabric.client.ChaincodeEvent;
import org.hyperledger.fabric.client.GatewayException;
import org.hyperledger.fabric.client.identity.Identities;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.client.identity.Signers;
import org.hyperledger.fabric.client.identity.X509Identity;
import org.hyperledger.fabric.protos.common.Block;
import org.hyperledger.fabric.protos.gateway.ErrorDetail;
import org.hyperledger.fabric.protos.peer.BlockAndPrivateData;
import org.hyperledger.fabric.protos.peer.FilteredBlock;

import java.io.BufferedReader;
import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.Reader;
import java.nio.charset.StandardCharsets;
import java.nio.file.FileSystems;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.security.InvalidKeyException;
import java.security.PrivateKey;
import java.security.cert.CertificateException;
import java.security.cert.X509Certificate;
import java.security.interfaces.ECPrivateKey;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Collectors;

import static org.assertj.core.api.Assertions.assertThat;

public class ScenarioSteps {
    private static final Map<String, String> runningChaincodes = new HashMap<>();
    private static boolean channelsJoined = false;
    private static final String DOCKER_COMPOSE_TLS_FILE = "docker-compose-tls.yaml";
    private static final Path FIXTURES_DIR = Paths.get("..", "scenario", "fixtures").toAbsolutePath();
    private static final Path DOCKER_COMPOSE_DIR = Paths.get(FIXTURES_DIR.toString(), "docker-compose")
            .toAbsolutePath();
    private static final String DEFAULT_LISTENER_NAME = "";

    private static final Map<String, String> MSP_ID_TO_ORG_MAP;
    static {
        Map<String, String> mspIdToOrgMap = new HashMap<>();
        mspIdToOrgMap.put("Org1MSP", "org1.example.com");
        mspIdToOrgMap.put("Org2MSP", "org2.example.com");
        mspIdToOrgMap.put("Org3MSP", "org3.example.com");
        MSP_ID_TO_ORG_MAP = Collections.unmodifiableMap(mspIdToOrgMap);
    }

    private static final Map<String, ConnectionInfo> peerConnectionInfo = new HashMap<>();
    static {
        String certPathTemplate = FIXTURES_DIR + "/crypto-material/crypto-config/peerOrganizations/org$O.example.com/peers/peer$P.org$O.example.com/tls/ca.crt";
        peerConnectionInfo.put("peer0.org1.example.com",
                new ConnectionInfo("localhost:7051", "peer0.org1.example.com", certPathTemplate.replace("$P", "0").replace("$O", "1")));
        peerConnectionInfo.put("peer1.org1.example.com",
                new ConnectionInfo("localhost:9051", "peer1.org1.example.com", certPathTemplate.replace("$P", "1").replace("$O", "1")));
        peerConnectionInfo.put("peer0.org2.example.com",
                new ConnectionInfo("localhost:8051", "peer0.org2.example.com", certPathTemplate.replace("$P", "0").replace("$O", "2")));
        peerConnectionInfo.put("peer1.org2.example.com",
                new ConnectionInfo("localhost:10051", "peer1.org2.example.com", certPathTemplate.replace("$P", "1").replace("$O", "2")));
        peerConnectionInfo.put("peer0.org3.example.com",
                new ConnectionInfo("localhost:11051", "peer0.org3.example.com", certPathTemplate.replace("$P", "0").replace("$O", "3")));
    }

    private static final Collection<OrgConfig> ORG_CONFIGS;
    static {
        List<OrgConfig> orgConfigs = Arrays.asList(
                new OrgConfig("org1_cli", "peer0.org1.example.com:7051", "peer1.org1.example.com:9051"),
                new OrgConfig("org2_cli", "peer0.org2.example.com:8051", "peer1.org2.example.com:10051"),
                new OrgConfig("org3_cli", "peer0.org3.example.com:11051")
        );
        ORG_CONFIGS = Collections.unmodifiableCollection(orgConfigs);
    }

    private GatewayContext currentGateway;
    private final Map<String, GatewayContext> gateways = new HashMap<>();
    private TransactionInvocation transactionInvocation;
    private long lastCommittedBlockNumber;
    private final Gson gson = new Gson();

    private static final class OrgConfig {
        final String cli;
        final Set<String> peers;

        OrgConfig(String cli, String... peers) {
            this.cli = cli;
            Set<String> peerSet = new HashSet<>(Arrays.asList(peers));
            this.peers = Collections.unmodifiableSet(peerSet);
        }
    }

    private static final Collection<OrdererConfig> ORDERER_CONFIGS;
    static {
        List<OrdererConfig> ordererConfigs = Arrays.asList(
                new OrdererConfig("orderer1.example.com", "7053"),
                new OrdererConfig("orderer2.example.com", "8053"),
                new OrdererConfig("orderer3.example.com", "9053")
        );
        ORDERER_CONFIGS = Collections.unmodifiableCollection(ordererConfigs);
    }

    private static final class OrdererConfig {
        final String address;
        final String port;

        OrdererConfig(String address, String port) {
            this.address = address;
            this.port = port;
        }
    }

    private static final class ConnectionInfo {
        final String url;
        final String serverNameOverride;
        final String tlsRootCertPath;
        boolean running = true;

        ConnectionInfo(String url, String serverNameOverride, String tlsRootCertPath) {
            this.url = url;
            this.serverNameOverride = serverNameOverride;
            this.tlsRootCertPath = tlsRootCertPath;
        }

        void start() {
            running = true;
        }

        void stop() {
            running = false;
        }
    }

    @After
    public void afterEach() {
        for (GatewayContext context : gateways.values()) {
            context.close();
        }
    }

    @Given("I have deployed a Fabric network")
    public void deployFabricNetwork() {
    }

    @Given("I have created and joined all channels")
    public void createAndJoinAllChannels() throws IOException, InterruptedException {
        // TODO this only does mychannel
        startAllPeers();
        if (!channelsJoined) {
            final List<String> tlsOptions = Arrays.asList("--tls", "true", "--cafile",
                    "/etc/hyperledger/configtx/crypto-config/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem");

            for (OrdererConfig orderer : ORDERER_CONFIGS) {
                String ordDir = "/etc/hyperledger/configtx/crypto-config/ordererOrganizations/example.com/orderers/" + orderer.address;
                List<String> createChannelCommand = new ArrayList<>();
                Collections.addAll(createChannelCommand, "docker", "exec", "org1_cli", "osnadmin", "channel", "join",
                        "--channelID", "mychannel",
                        "--config-block", "/etc/hyperledger/configtx/mychannel.block",
                        "-o", orderer.address + ":" + orderer.port,
                        "--ca-file", "/etc/hyperledger/configtx/crypto-config/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem",
                        "--client-cert", ordDir + "/tls/server.crt",
                        "--client-key", ordDir + "/tls/server.key");
                exec(createChannelCommand);
            }

            for (OrgConfig org : ORG_CONFIGS) {
                for (String peer : org.peers) {
                    String env = "CORE_PEER_ADDRESS=" + peer;
                    List<String> joinChannelCommand = new ArrayList<>();
                    Collections.addAll(joinChannelCommand, "docker", "exec", "-e", env, org.cli, "peer", "channel", "join",
                            "-b", "/etc/hyperledger/configtx/mychannel.block");
                    joinChannelCommand.addAll(tlsOptions);
                    exec(joinChannelCommand);
                }

            }

            channelsJoined = true;
        }
    }

    @Given("I deploy {word} chaincode named {word} at version {word} for all organizations on channel {word} with endorsement policy {}")
    public void deployChaincode(String ccType, String ccName, String version, String channelName, String signaturePolicy) throws IOException, InterruptedException {
        final List<String> tlsOptions = Arrays.asList("--tls", "true", "--cafile",
                "/etc/hyperledger/configtx/crypto-config/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem");

        boolean exists = false;
        String sequence = "1";
        String mangledName = ccName + version + channelName;
        if (runningChaincodes.containsKey(mangledName)) {
            if (runningChaincodes.get(mangledName).equals(signaturePolicy)) {
                return;
            }
            // Already exists but different signature policy...
            // No need to re-install, just increment the sequence number and approve/commit new signature policy
            exists = true;
            List<String> queryCommand = new ArrayList<>();
            Collections.addAll(queryCommand,"docker", "exec", "org1_cli", "peer", "lifecycle", "chaincode", "querycommitted",
                    "-o", "orderer1.example.com:7050", "--channelID", channelName, "--name", ccName);
            queryCommand.addAll(tlsOptions);
            String out = exec(queryCommand);
            Pattern regex = Pattern.compile(".*Sequence: (\\d+),.*");
            Matcher matcher = regex.matcher(out);
            if (!matcher.matches()) {
                System.out.println(out);
                throw new IllegalStateException("Cannot find installed chaincode for Org1: " + ccName);
            }
            String seqStr = matcher.group(1);
            int seqInt = Integer.parseInt(seqStr);
            sequence = String.valueOf(seqInt + 1);
        }

        String ccPath = Paths.get(FileSystems.getDefault().getSeparator(), "opt", "gopath", "src",
                "github.com", "chaincode", ccType, ccName).toString();
        String ccLabel = ccName + "v" + version;
        String ccPackage = ccName + ".tar.gz";

        String collectionsConfig = null;
        String collectionsFile = Paths.get("chaincode", ccType, ccName, "collections_config.json").toString();
        if (Paths.get(FIXTURES_DIR.toString(), collectionsFile).toAbsolutePath().toFile().exists()) {
            collectionsConfig = Paths.get("/opt/gopath/src/github.com", collectionsFile).toString();
        }

        for (OrgConfig org : ORG_CONFIGS) {
            if (!exists) {
                exec("docker", "exec", org.cli, "peer", "lifecycle", "chaincode", "package", ccPackage, "--lang",
                        ccType, "--label", ccLabel, "--path", ccPath);

                for (String peer : org.peers) {
                    String env = "CORE_PEER_ADDRESS=" + peer;
                    exec("docker", "exec", "-e", env, org.cli, "peer", "lifecycle", "chaincode", "install", ccPackage);
                }
            }

            String installed = exec("docker", "exec", org.cli, "peer", "lifecycle", "chaincode",
                    "queryinstalled");
            Pattern regex = Pattern.compile(".*Package ID: (.*), Label: " + ccLabel + ".*");
            Matcher matcher = regex.matcher(installed);
            if (!matcher.matches()) {
                System.out.println(installed);
                throw new IllegalStateException("Cannot find installed chaincode for Org1: " + ccLabel);
            }
            String packageId = matcher.group(1);

            List<String> approveCommand = new ArrayList<>();
            Collections.addAll(approveCommand, "docker", "exec", org.cli, "peer", "lifecycle", "chaincode",
                    "approveformyorg", "--package-id", packageId, "--channelID", channelName, "--name", ccName,
                    "--version", version, "--signature-policy", signaturePolicy,
                    "--sequence", sequence, "--waitForEvent");
            if(collectionsConfig != null) {
                Collections.addAll(approveCommand, "--collections-config", collectionsConfig);
            }
            approveCommand.addAll(tlsOptions);
            exec(approveCommand);
        }

        // commit
        List<String> commitCommand = new ArrayList<>();
        Collections.addAll(commitCommand, "docker", "exec", "org1_cli", "peer", "lifecycle", "chaincode", "commit",
                "--channelID", channelName,
                "--name", ccName,
                "--version", version,
                "--signature-policy", signaturePolicy,
                "--sequence", sequence,
                "--waitForEvent",
                "--peerAddresses", "peer0.org1.example.com:7051",
                "--peerAddresses", "peer0.org2.example.com:8051",
                "--tlsRootCertFiles",
                "/etc/hyperledger/configtx/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt",
                "--tlsRootCertFiles",
                "/etc/hyperledger/configtx/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt");
        if(collectionsConfig != null) {
            Collections.addAll(commitCommand, "--collections-config", collectionsConfig);
        }
        commitCommand.addAll(tlsOptions);
        exec(commitCommand);

        runningChaincodes.put(mangledName, signaturePolicy);
        Thread.sleep(10000);
    }

    @Given("I create a gateway named {word} for user {word} in MSP {word}")
    public void createGateway(String name, String user, String mspId) throws CertificateException, InvalidKeyException, IOException {
        Identity identity = newIdentity(user, mspId);
        Signer signer = newSigner(user, mspId);
        currentGateway = new GatewayContext(identity, signer);
        gateways.put(name, currentGateway);
    }

    @Given("I create a gateway named {word} without signer for user {word} in MSP {word}")
    public void createGatewayWithoutSigner(String name, String user, String mspId) throws CertificateException, IOException {
        Identity identity = newIdentity(user, mspId);
        currentGateway = new GatewayContext(identity);
        gateways.put(name, currentGateway);
    }

    @Given("I connect the gateway to {word}")
    public void connectGateway(String name) throws Exception {
        ConnectionInfo info = peerConnectionInfo.get(name);
        SslContext sslContext = GrpcSslContexts.forClient()
                .trustManager(Files.newInputStream(Paths.get(info.tlsRootCertPath)))
                .build();
        ManagedChannel channel = NettyChannelBuilder.forTarget(info.url)
                .sslContext(sslContext)
                .overrideAuthority(info.serverNameOverride)
                .build();
        currentGateway.connect(channel);
    }

    @Given("I use the gateway named {word}")
    public void useGateway(String name) {
        currentGateway = gateways.get(name);
    }

    @Given("I use the {word} network")
    public void useNetwork(String networkName) {
        currentGateway.useNetwork(networkName);
    }

    @Given("I use the {word} contract")
    public void useContract(String contractName) {
        currentGateway.useContract(contractName);
    }

    @Given("I create a checkpointer")
    public void createCheckpointer() {
        currentGateway.createCheckpointer();
    }

    @When("^I prepare to (evaluate|submit) an? ([^ ]+) transaction$")
    public void prepareTransaction(String action, String transactionName) {
        transactionInvocation = currentGateway.newTransaction(action, transactionName);
    }

    @When("^I set the transaction arguments? to (.+)$")
    public void setTransactionArguments(String argsJson) {
        String[] args = gson.fromJson(argsJson, String[].class);
        transactionInvocation.setArguments(args);
    }

    @When("I do off-line signing as user {word} in MSP {word}")
    public void offlineSign(String user, String mspId) throws InvalidKeyException, IOException {
        transactionInvocation.setOfflineSigner(newSigner(user, mspId));
    }

    @When("I invoke the transaction")
    public void invokeSuccessfulTransaction() {
        invokeTransaction();
        transactionInvocation.getResponse();
    }

    @When("I set transient data on the transaction to")
    public void setTransientData(DataTable data) {
        Map<String, String> transientMap = data.asMap(String.class, String.class);
        transactionInvocation.setTransient(transientMap);
    }

    @When("^I set the endorsing organizations? to (.+)$")
    public void setEndorsingOrgs(String orgsJson) {
        String[] orgs = gson.fromJson(orgsJson, String[].class);
        transactionInvocation.setEndorsingOrgs(orgs);
    }

    @When("I stop the peer named {}")
    public void stopPeer(String peer) throws IOException, InterruptedException {
        exec("docker", "stop", peer);
        peerConnectionInfo.get(peer).stop();
    }

    @When("I start the peer named {}")
    public void startPeer(String peer) throws IOException, InterruptedException {
        exec("docker", "start", peer);
        peerConnectionInfo.get(peer).stop();
        Thread.sleep(20000);
    }

    @When("I listen for chaincode events from {word}")
    public void listenForChaincodeEvents(String chaincodeName) {
        listenForChaincodeEventsOnListener(chaincodeName, DEFAULT_LISTENER_NAME);
    }

    @When("I use the checkpointer to listen for chaincode events from {word}")
    public void listenForChaincodeEventsUsingCheckpointer(String chaincodeName){
        currentGateway.listenForChaincodeEventsUsingCheckpointer(DEFAULT_LISTENER_NAME, chaincodeName);
    }

    @When("I listen for chaincode events from {word} on a listener named {string}")
    public void listenForChaincodeEventsOnListener(String chaincodeName, String listenerName) {
        currentGateway.listenForChaincodeEvents(listenerName, chaincodeName);
    }

    @When("I replay chaincode events from {word} starting at last committed block")
    public void replayChaincodeEventsFromLastBlock(String chaincodeName) {
        currentGateway.replayChaincodeEvents(DEFAULT_LISTENER_NAME, chaincodeName, lastCommittedBlockNumber);
    }

    @When("I stop listening for chaincode events")
    public void stopChaincodeEventListening() {
        stopChaincodeEventListeningOnListener(DEFAULT_LISTENER_NAME);
    }

    @When("I stop listening for chaincode events on {string}")
    public void stopChaincodeEventListeningOnListener(String listenerName) {
        currentGateway.closeChaincodeEvents(listenerName);
    }

    @When("I listen for block events")
    public void listenForBlockEvents() {
        listenForBlockEventsOnListener(DEFAULT_LISTENER_NAME);
    }

    @When("I listen for block events on a listener named {string}")
    public void listenForBlockEventsOnListener(String listenerName) {
        currentGateway.listenForBlockEvents(listenerName);
    }

    @When("I use the checkpointer to listen for block events")
    public void listenForBlockEventsUsingCheckpointer() {
        currentGateway.listenForBlockEventsUsingCheckpointer(DEFAULT_LISTENER_NAME);
    }

    @When("I use the checkpointer to listen for filtered block events")
    public void listenForFilteredBlockEventsUsingCheckpointer() {
        currentGateway.listenForFilteredBlockEventsUsingCheckpointer(DEFAULT_LISTENER_NAME);
    }

    @When("I use the checkpointer to listen for block and private data events")
    public void listenForBlockAndPrivateDataUsingCheckpointer(){
        currentGateway.listenForBlockAndPrivateDataUsingCheckpointer(DEFAULT_LISTENER_NAME);
    }

    @When("I replay block events starting at last committed block")
    public void replayBlockEventsFromLastBlock() {
        currentGateway.replayBlockEvents(DEFAULT_LISTENER_NAME, lastCommittedBlockNumber);
    }

    @When("I stop listening for block events")
    public void stopBlockEventListening() {
        stopBlockEventListeningOnListener(DEFAULT_LISTENER_NAME);
    }

    @When("I stop listening for block events on {string}")
    public void stopBlockEventListeningOnListener(String listenerName) {
        currentGateway.closeBlockEvents(listenerName);
    }

    @When("I listen for filtered block events")
    public void listenForFilteredBlockEvents() {
        listenForFilteredBlockEventsOnListener(DEFAULT_LISTENER_NAME);
    }

    @When("I listen for filtered block events on a listener named {string}")
    public void listenForFilteredBlockEventsOnListener(String listenerName) {
        currentGateway.listenForFilteredBlockEvents(listenerName);
    }

    @When("I replay filtered block events starting at last committed block")
    public void replayFilteredBlockEventsFromLastBlock() {
        currentGateway.replayFilteredBlockEvents(DEFAULT_LISTENER_NAME, lastCommittedBlockNumber);
    }

    @When("I stop listening for filtered block events")
    public void stopFilteredBlockEventListening() {
        stopBlockEventListeningOnListener(DEFAULT_LISTENER_NAME);
    }

    @When("I stop listening for filtered block events on {string}")
    public void stopFilteredBlockEventListeningOnListener(String listenerName) {
        currentGateway.closeFilteredBlockEvents(listenerName);
    }

    @When("I listen for block and private data events")
    public void listenForBlockAndPrivateDataEvents() {
        listenForBlockAndPrivateDataEventsOnListener(DEFAULT_LISTENER_NAME);
    }

    @When("I listen for block and private data events on a listener named {string}")
    public void listenForBlockAndPrivateDataEventsOnListener(String listenerName) {
        currentGateway.listenForBlockAndPrivateDataEvents(listenerName);
    }

    @When("I replay block and private data events starting at last committed block")
    public void replayBlockAndPrivateDataEventsFromLastBlock() {
        currentGateway.replayBlockAndPrivateDataEvents(DEFAULT_LISTENER_NAME, lastCommittedBlockNumber);
    }

    @When("I stop listening for block and private data events")
    public void stopBlockAndPrivateDataEventListening() {
        stopBlockEventListeningOnListener(DEFAULT_LISTENER_NAME);
    }

    @When("I stop listening for block and private data events on {string}")
    public void stopBlockAndPrivateDataEventListeningOnListener(String listenerName) {
        currentGateway.closeBlockAndPrivateDataEvents(listenerName);
    }

    @Then("the transaction invocation should fail")
    public void assertTransactionFails() {
        invokeTransaction();
        transactionInvocation.getError();
    }

    @Then("the response should be JSON matching")
    public void assertJsonResponse(DocString expected) {
        JsonElement expectedElement = JsonParser.parseString(expected.getContent());
        JsonElement actualElement = JsonParser.parseString(transactionInvocation.getResponse());
        assertThat(actualElement).isEqualTo(expectedElement);
    }

    @Then("the response should be {string}")
    public void assertResponse(String expected) {
        assertThat(transactionInvocation.getResponse()).isEqualTo(expected);
    }

    @Then("the error message should contain {string}")
    public void assertErrorMessageContains(String expected) {
        assertThat(transactionInvocation.getError()).hasMessageContaining(expected);
    }

    @Then("the error details should be")
    public void assertErrorDetails(Map<String, List<String>> table) {
        Throwable t = transactionInvocation.getError();
        assertThat(t).isInstanceOf(GatewayException.class);

        Map<String, List<String>> expected = new HashMap<>(table);

        for (ErrorDetail detail : ((GatewayException) t).getDetails()) {
            String address = detail.getAddress();
            List<String> row = expected.get(address);
            assertThat(row).isNotNull();
            assertThat(detail.getMessage()).contains(row.get(1));
            expected.remove(address);
        }

        assertThat(expected).isEmpty();
    }

    @Then("the error status should be {word}")
    public void assertErrorStatus(String expected) {
        Status.Code expectedCode = Status.Code.valueOf(expected);

        Throwable t = transactionInvocation.getError();
        assertThat(t).isInstanceOf(GatewayException.class);

        GatewayException e = (GatewayException) t;
        Status.Code actual = e.getStatus().getCode();

        assertThat(actual).isEqualTo(expectedCode);
    }

    @Then("I should receive a chaincode event named {string} with payload {string}")
    public void assertReceiveChaincodeEvent(String eventName, String payload) throws InterruptedException, IOException {
        assertReceiveChaincodeEventOnListener(eventName, payload, DEFAULT_LISTENER_NAME);
    }

    @Then("I should receive a chaincode event named {string} with payload {string} on {string}")
    public void assertReceiveChaincodeEventOnListener(String eventName, String payload, String listenerName) throws InterruptedException, IOException {
        ChaincodeEvent event = currentGateway.nextChaincodeEvent(listenerName);
        assertThat(event.getEventName()).isEqualTo(eventName);
        assertThat(new String(event.getPayload(), StandardCharsets.UTF_8)).isEqualTo(payload);
    }

    @Then("I should receive a block event")
    public void assertReceiveBlockEvent() throws InterruptedException, IOException {
        assertReceiveBlockEventOnListener(DEFAULT_LISTENER_NAME);
    }

    @Then("I should receive a block event on {string}")
    public void assertReceiveBlockEventOnListener(String listenerName) throws InterruptedException, IOException {
        Block event = currentGateway.nextBlockEvent(listenerName);
        assertThat(event).isNotNull();
    }

    @Then("I should receive a filtered block event")
    public void assertReceiveFilteredBlockEvent() throws InterruptedException, IOException {
        assertReceiveFilteredBlockEventOnListener(DEFAULT_LISTENER_NAME);
    }

    @Then("I should receive a filtered block event on {string}")
    public void assertReceiveFilteredBlockEventOnListener(String listenerName) throws InterruptedException, IOException {
        FilteredBlock event = currentGateway.nextFilteredBlockEvent(listenerName);
        assertThat(event).isNotNull();
    }

    @Then("I should receive a block and private data event")
    public void assertReceiveBlockAndPrivateDataEvent() throws InterruptedException, IOException {
        assertReceiveBlockAndPrivateDataEventOnListener(DEFAULT_LISTENER_NAME);
    }

    @Then("I should receive a block and private data event on {string}")
    public void assertReceiveBlockAndPrivateDataEventOnListener(String listenerName) throws InterruptedException, IOException {
        BlockAndPrivateData event = currentGateway.nextBlockAndPrivateDataEvent(listenerName);
        assertThat(event).isNotNull();
    }

    private void invokeTransaction() {
        transactionInvocation.invoke();
        lastCommittedBlockNumber = transactionInvocation.getBlockNumber();
    }

    private static void startAllPeers() throws InterruptedException, IOException {
        Set<String> stoppedPeers = peerConnectionInfo.entrySet().stream()
                .filter(entry -> !entry.getValue().running)
                .map(Map.Entry::getKey)
                .collect(Collectors.toSet());

        if (stoppedPeers.isEmpty()) {
            return;
        }

        for (String peer : stoppedPeers) {
            exec("docker", "start", peer);
            peerConnectionInfo.get(peer).start();
        }
        Thread.sleep(20000);
    }

    public static String newString(byte[] bytes) {
        return new String(bytes, StandardCharsets.UTF_8);
    }

    private static String exec(List<String> commandArgs) throws IOException, InterruptedException {
        return exec(null, commandArgs);
    }

    private static String exec(Path dir, List<String> commandArgs) throws IOException, InterruptedException {
        return exec(dir, commandArgs.toArray(new String[0]));
    }

    private static String exec(String... commandArgs) throws IOException, InterruptedException {
        return exec(null, commandArgs);
    }

    private static String exec(Path dir, String... commandArgs) throws IOException, InterruptedException {
        String commandString = String.join(" ", commandArgs);
        System.err.println(commandString);
        StringBuilder sb = new StringBuilder();

        File dirFile = dir != null ? dir.toFile() : null;
        Process process = Runtime.getRuntime().exec(commandArgs, null, dirFile);
        int exitCode = process.waitFor();

        // get STDERR for the process and print it
        try (InputStream errorStream = process.getErrorStream();
             BufferedReader reader = new BufferedReader(new InputStreamReader(errorStream))) {
            for (String line; (line = reader.readLine()) != null; ) {
                System.err.println(line);
                sb.append(line);
            }
        }

        // get STDOUT for the process and print it
        try (InputStream inputStream = process.getInputStream();
             BufferedReader reader = new BufferedReader(new InputStreamReader(inputStream))) {
            for (String line; (line = reader.readLine()) != null; ) {
                System.out.println(line);
                sb.append(line);
            }
        }

        assertThat(exitCode).withFailMessage("Failed to execute command: %s", commandString).isEqualTo(0);
        return sb.toString();
    }

    static void startFabric() throws Exception {
        createCryptoMaterial();
        exec(DOCKER_COMPOSE_DIR, "docker-compose", "-f", DOCKER_COMPOSE_TLS_FILE, "-p", "node", "up", "-d");
        Thread.sleep(10000);
    }

    static void stopFabric() throws Exception {
        exec(DOCKER_COMPOSE_DIR, "docker-compose", "-f", DOCKER_COMPOSE_TLS_FILE, "-p", "node", "down");
    }

    private static void createCryptoMaterial() throws Exception {
        Path fixtures = Paths.get("..", "scenario", "fixtures");
        exec(fixtures, "./generate.sh");
    }

    private static Identity newIdentity(String user, String mspId) throws IOException, CertificateException {
        String org = getOrgForMspId(mspId);
        Path credentialPath = getCredentialPath(user, org);
        Path certificatePath = credentialPath.resolve(Paths.get("signcerts", user + "@" + org + "-cert.pem"));
        X509Certificate certificate = readX509Certificate(certificatePath);

        return new X509Identity(mspId, certificate);
    }

    private static Signer newSigner(String user, String mspId) throws IOException, InvalidKeyException {
        String org = getOrgForMspId(mspId);
        Path credentialPath = getCredentialPath(user, org);
        Path privateKeyPath = credentialPath.resolve(Paths.get("keystore", "key.pem"));
        PrivateKey privateKey = getPrivateKey(privateKeyPath);

        if (privateKey instanceof ECPrivateKey) {
            return Signers.newPrivateKeySigner(privateKey);
        }
        throw new RuntimeException("Unexpected private key type: " + privateKey.getClass().getSimpleName());
    }

    private static String getOrgForMspId(String mspId) {
        String org = MSP_ID_TO_ORG_MAP.get(mspId);
        if (null == org) {
            throw new IllegalArgumentException("Unknown MSP ID: " + mspId);
        }
        return org;
    }

    private static Path getCredentialPath(String user, String org) {
        return Paths.get("..", "scenario", "fixtures", "crypto-material", "crypto-config",
                "peerOrganizations", org, "users", user + "@" + org, "msp");
    }

    private static X509Certificate readX509Certificate(final Path certificatePath)
            throws IOException, CertificateException {
        try (Reader certificateReader = Files.newBufferedReader(certificatePath, StandardCharsets.UTF_8)) {
            return Identities.readX509Certificate(certificateReader);
        }
    }

    private static PrivateKey getPrivateKey(final Path privateKeyPath) throws IOException, InvalidKeyException {
        try (Reader privateKeyReader = Files.newBufferedReader(privateKeyPath, StandardCharsets.UTF_8)) {
            return Identities.readPrivateKey(privateKeyReader);
        }
    }
}
