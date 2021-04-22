package scenario;

import java.io.*;
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
import javax.json.Json;
import javax.json.JsonArray;
import javax.json.JsonObject;
import javax.json.JsonReader;
import javax.json.JsonString;

import io.cucumber.datatable.DataTable;
import io.cucumber.docstring.DocString;
import io.cucumber.java.After;
import io.cucumber.java.en.Given;
import io.cucumber.java.en.Then;
import io.cucumber.java.en.When;
import io.grpc.ManagedChannel;
import io.grpc.ManagedChannelBuilder;
import io.grpc.netty.shaded.io.grpc.netty.GrpcSslContexts;
import io.grpc.netty.shaded.io.grpc.netty.NettyChannelBuilder;
import org.hyperledger.fabric.client.Contract;
import org.hyperledger.fabric.client.Gateway;
import org.hyperledger.fabric.client.Network;
import org.hyperledger.fabric.client.identity.Identities;
import org.hyperledger.fabric.client.identity.Identity;
import org.hyperledger.fabric.client.identity.Signer;
import org.hyperledger.fabric.client.identity.Signers;
import org.hyperledger.fabric.client.identity.X509Identity;

import static org.assertj.core.api.Assertions.assertThat;

public class ScenarioSteps {
    private static final Set<String> runningChaincodes = new HashSet<>();
    private static boolean channelsJoined = false;
    private static final String DOCKER_COMPOSE_TLS_FILE = "docker-compose-tls.yaml";
    private static final Path FIXTURES_DIR = Paths.get("..", "scenario", "fixtures").toAbsolutePath();
    private static final Path DOCKER_COMPOSE_DIR = Paths.get(FIXTURES_DIR.toString(), "docker-compose")
            .toAbsolutePath();

    private static final Map<String, String> MSP_ID_TO_ORG_MAP;

    static {
        Map<String, String> mspIdToOrgMap = new HashMap<>();
        mspIdToOrgMap.put("Org1MSP", "org1.example.com");
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
                new OrgConfig("org1_cli", "/etc/hyperledger/configtx/Org1MSPanchors.tx", "peer0.org1.example.com:7051", "peer1.org1.example.com:9051"),
                new OrgConfig("org2_cli", "/etc/hyperledger/configtx/Org2MSPanchors.tx", "peer0.org2.example.com:8051", "peer1.org2.example.com:10051"),
                new OrgConfig("org3_cli", "/etc/hyperledger/configtx/Org3MSPanchors.tx", "peer0.org3.example.com:11051")
        );
        ORG_CONFIGS = Collections.unmodifiableCollection(orgConfigs);
    }

    private Gateway.Builder gatewayBuilder;
    private Gateway gateway;
    private Network network;
    private Contract contract;
    private TransactionInvocation transactionInvocation;

    private static final class OrgConfig {
        final String cli;
        final String anchortx;
        final Set<String> peers;

        OrgConfig(String cli, String anchortx, String... peers) {
            this.cli = cli;
            this.anchortx = anchortx;
            Set<String> peerSet = new HashSet<>(Arrays.asList(peers));
            this.peers = Collections.unmodifiableSet(peerSet);
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
        if (gateway != null) {
            gateway.close();
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

            List<String> createChannelCommand = new ArrayList<>();
            Collections.addAll(createChannelCommand, "docker", "exec", "org1_cli", "peer", "channel", "create",
                    "-o", "orderer.example.com:7050", "-c", "mychannel", "-f",
                    "/etc/hyperledger/configtx/channel.tx", "--outputBlock",
                    "/etc/hyperledger/configtx/mychannel.block");
            createChannelCommand.addAll(tlsOptions);
            exec(createChannelCommand);

            for (OrgConfig org : ORG_CONFIGS) {
                for (String peer : org.peers) {
                    String env = "CORE_PEER_ADDRESS=" + peer;
                    List<String> joinChannelCommand = new ArrayList<>();
                    Collections.addAll(joinChannelCommand, "docker", "exec", "-e", env, org.cli, "peer", "channel", "join",
                            "-b", "/etc/hyperledger/configtx/mychannel.block");
                    joinChannelCommand.addAll(tlsOptions);
                    exec(joinChannelCommand);
                }

                List<String> anchorPeersCommand = new ArrayList<>();
                Collections.addAll(anchorPeersCommand, "docker", "exec", org.cli, "peer", "channel", "update",
                        "-o", "orderer.example.com:7050", "-c", "mychannel", "-f", org.anchortx);
                anchorPeersCommand.addAll(tlsOptions);
                exec(anchorPeersCommand);
            }

            channelsJoined = true;
        }
    }

    @Given("I deploy {word} chaincode named {word} at version {word} for all organizations on channel {word} with endorsement policy {}")
    public void deployChaincode(String ccType, String ccName, String version, String channelName, String signaturePolicy) throws IOException, InterruptedException {
        String mangledName = ccName + version + channelName;
        if (runningChaincodes.contains(mangledName)) {
            return;
        }

        final List<String> tlsOptions = Arrays.asList("--tls", "true", "--cafile",
                "/etc/hyperledger/configtx/crypto-config/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem");

        String ccPath = Paths.get(FileSystems.getDefault().getSeparator(), "opt", "gopath", "src",
                "github.com", "chaincode", ccType, ccName).toString();
        String ccLabel = ccName + "v" + version;
        String ccPackage = ccName + ".tar.gz";

        for (OrgConfig org : ORG_CONFIGS) {
            exec("docker", "exec", org.cli, "peer", "lifecycle", "chaincode", "package", ccPackage, "--lang",
                    ccType, "--label", ccLabel, "--path", ccPath);

            for (String peer : org.peers) {
                String env = "CORE_PEER_ADDRESS=" + peer;
                exec("docker", "exec", "-e", env, org.cli, "peer", "lifecycle", "chaincode", "install", ccPackage);
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
                    "--sequence", "1", "--waitForEvent");
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
                "--sequence", "1",
                "--waitForEvent",
                "--peerAddresses", "peer0.org1.example.com:7051",
                "--peerAddresses",
                "peer0.org2.example.com:8051",
                "--tlsRootCertFiles",
                "/etc/hyperledger/configtx/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt",
                "--tlsRootCertFiles",
                "/etc/hyperledger/configtx/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt");
        commitCommand.addAll(tlsOptions);
        exec(commitCommand);

        runningChaincodes.add(mangledName);
        Thread.sleep(60000);
    }

    @Given("I create a gateway for user {word} in MSP {word}")
    public void createGateway(String user, String mspId) throws CertificateException, InvalidKeyException, IOException {
        gatewayBuilder = Gateway.newInstance()
                .identity(newIdentity(user, mspId))
                .signer(newSigner(user, mspId));
    }

    @Given("I create a gateway without signer for user {word} in MSP {word}")
    public void createGatewayWithoutSigner(String user, String mspId) throws CertificateException, IOException {
        gatewayBuilder = Gateway.newInstance()
                .identity(newIdentity(user, mspId));
    }

    @Given("I connect the gateway to {word}")
    public void connectGateway(String name) throws Exception {
        ConnectionInfo info = peerConnectionInfo.get(name);
        ManagedChannel channel = NettyChannelBuilder.forTarget(info.url)
                .sslContext(GrpcSslContexts.forClient().trustManager(new FileInputStream(info.tlsRootCertPath)).build())
                .overrideAuthority(info.serverNameOverride)
                .build();

        gateway = gatewayBuilder.connection(channel).connect();
    }

    @Given("I use the {word} network")
    public void useNetwork(String networkName) {
        network = gateway.getNetwork(networkName);
    }

    @Given("I use the {word} contract")
    public void useContract(String contractName) {
        contract = network.getContract(contractName);
    }

    @When("^I prepare to (evaluate|submit) an? ([^ ]+) transaction$")
    public void prepareTransaction(String action, String transactionName) {
        if (action.equals("submit")) {
            transactionInvocation = TransactionInvocation.prepareToSubmit(contract, transactionName);
        } else {
            transactionInvocation = TransactionInvocation.prepareToEvaluate(contract, transactionName);
        }
    }

    @When("^I set the transaction arguments? to (.+)$")
    public void setTransactionArguments(String argsJson) {
        String[] args = newStringArray(parseJsonArray(argsJson));
        transactionInvocation.setArguments(args);
    }

    @When("I do off-line signing as user {word} in MSP {word}")
    public void offlineSign(String user, String mspId) throws InvalidKeyException, IOException {
        transactionInvocation.setOfflineSigner(newSigner(user, mspId));
    }

    @When("I invoke the transaction")
    public void invokeTransaction() {
        transactionInvocation.invoke();
        transactionInvocation.getResponse();
    }

    @When("I set transient data on the transaction to")
    public void setTransientData(DataTable data) {
        Map<String, String> table = data.asMap(String.class, String.class);
        Map<String, byte[]> transientMap = table.entrySet().stream()
                .collect(Collectors.toMap(Map.Entry::getKey, entry -> entry.getValue().getBytes(StandardCharsets.UTF_8)));
        transactionInvocation.setTransient(transientMap);
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

    @Then("the transaction invocation should fail")
    public void invokeFailingTransaction() {
        transactionInvocation.invoke();
        transactionInvocation.getError();
    }

    @Then("the response should be JSON matching")
    public void assertJsonResponse(DocString expected) {
        try (JsonReader expectedReader = createJsonReader(expected.getContent());
             JsonReader actualReader = createJsonReader(transactionInvocation.getResponse())) {
            JsonObject expectedObject = expectedReader.readObject();
            JsonObject actualObject = actualReader.readObject();
            assertThat(actualObject).isEqualTo(expectedObject);
        }
    }

    @Then("the response should be {string}")
    public void assertResponse(String expected) {
        assertThat(transactionInvocation.getResponse()).isEqualTo(expected);
    }

    @Then("the error message should contain {string}")
    public void assertErrorMessageContains(String expected) {
        assertThat(transactionInvocation.getError()).hasMessageContaining(expected);
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

    private static JsonArray parseJsonArray(String jsonString) {
        try (JsonReader reader = createJsonReader(jsonString)) {
            return reader.readArray();
        }
    }

    private static JsonReader createJsonReader(String jsonString) {
        return Json.createReader(new StringReader(jsonString));
    }

    private static String[] newStringArray(JsonArray jsonArray) {
        return jsonArray.getValuesAs(JsonString.class).stream().map(JsonString::getString).toArray(String[]::new);
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
            return Signers.newPrivateKeySigner((ECPrivateKey) privateKey);
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
