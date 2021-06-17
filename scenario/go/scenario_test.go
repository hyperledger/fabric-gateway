/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package scenario

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages-go/v10"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"github.com/hyperledger/fabric-protos-go/peer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

const (
	fixturesDir       = "../fixtures"
	dockerComposeFile = "docker-compose-tls.yaml"
	dockerComposeDir  = fixturesDir + "/docker-compose"
)

type TransactionType int

type orgConfig struct {
	cli      string
	anchortx string
	peers    []string
}

var orgs = []orgConfig{
	{
		cli:      "org1_cli",
		anchortx: "/etc/hyperledger/configtx/Org1MSPanchors.tx",
		peers:    []string{"peer0.org1.example.com:7051", "peer1.org1.example.com:9051"},
	},
	{
		cli:      "org2_cli",
		anchortx: "/etc/hyperledger/configtx/Org2MSPanchors.tx",
		peers:    []string{"peer0.org2.example.com:8051", "peer1.org2.example.com:10051"},
	},
	{
		cli:      "org3_cli",
		anchortx: "/etc/hyperledger/configtx/Org3MSPanchors.tx",
		peers:    []string{"peer0.org3.example.com:11051"},
	},
}

type connectionInfo struct {
	host               string
	port               uint16
	serverNameOverride string
	tlsRootCertPath    string
	running            bool
}

var peerConnectionInfo = map[string]*connectionInfo{
	"peer0.org1.example.com": {
		host:               "localhost",
		port:               7051,
		serverNameOverride: "peer0.org1.example.com",
		tlsRootCertPath:    fixturesDir + "/crypto-material/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt",
		running:            true,
	},
	"peer1.org1.example.com": {
		host:               "localhost",
		port:               9051,
		serverNameOverride: "peer1.org1.example.com",
		tlsRootCertPath:    fixturesDir + "/crypto-material/crypto-config/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/ca.crt",
		running:            true,
	},
	"peer0.org2.example.com": {
		host:               "localhost",
		port:               8051,
		serverNameOverride: "peer0.org2.example.com",
		tlsRootCertPath:    fixturesDir + "/crypto-material/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt",
		running:            true,
	},
	"peer1.org2.example.com": {
		host:               "localhost",
		port:               10051,
		serverNameOverride: "peer1.org2.example.com",
		tlsRootCertPath:    fixturesDir + "/crypto-material/crypto-config/peerOrganizations/org2.example.com/peers/peer1.org2.example.com/tls/ca.crt",
		running:            true,
	},
	"peer0.org3.example.com": {
		host:               "localhost",
		port:               11051,
		serverNameOverride: "peer0.org3.example.com",
		tlsRootCertPath:    fixturesDir + "/crypto-material/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt",
		running:            true,
	},
}

const (
	Evaluate TransactionType = iota
	Submit
)

var _mspToOrgMap map[string]string

func GetOrgForMSP(mspID string) string {
	if nil == _mspToOrgMap {
		_mspToOrgMap = make(map[string]string)
		_mspToOrgMap["Org1MSP"] = "org1.example.com"
		_mspToOrgMap["Org2MSP"] = "org2.example.com"
		_mspToOrgMap["Org3MSP"] = "org3.example.com"
	}

	return _mspToOrgMap[mspID]
}

type GatewayConnection struct {
	ID          identity.Identity
	options     []client.ConnectOption
	gateway     *client.Gateway
	network     *client.Network
	contract    *client.Contract
	transaction *Transaction
}

func NewGatewayConnection(user string, mspID string) (*GatewayConnection, error) {
	id, err := newIdentity(mspID, CertificatePath(user, mspID))
	if err != nil {
		return nil, err
	}

	connection := GatewayConnection{
		ID: id,
	}
	return &connection, nil
}

func NewGatewayConnectionWithSigner(user string, mspID string) (*GatewayConnection, error) {
	connection, err := NewGatewayConnection(user, mspID)
	if err != nil {
		return nil, err
	}

	sign, err := NewSign(PrivateKeyPath(user, mspID))
	if err != nil {
		return nil, err
	}

	connection.AddOptions(client.WithSign(sign))

	return connection, nil
}

func CertificatePath(user string, mspID string) string {
	org := GetOrgForMSP(mspID)
	return credentialsDirectory(user, org) + "/signcerts/" + user + "@" + org + "-cert.pem"
}

func PrivateKeyPath(user string, mspID string) string {
	return credentialsDirectory(user, GetOrgForMSP(mspID)) + "/keystore/key.pem"
}

func credentialsDirectory(user string, org string) string {
	return fixturesDir + "/crypto-material/crypto-config/peerOrganizations/" + org + "/users/" +
		user + "@" + org + "/msp"
}

func (connection *GatewayConnection) AddOptions(options ...client.ConnectOption) {
	connection.options = append(connection.options, options...)
}

func (connection *GatewayConnection) Connect() error {
	var err error
	connection.gateway, err = client.Connect(connection.ID, connection.options...)
	return err
}

func newIdentity(mspID string, certPath string) (*identity.X509Identity, error) {
	certificatePEM, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, err
	}

	certificate, err := identity.CertificateFromPEM(certificatePEM)
	if err != nil {
		return nil, err
	}

	return identity.NewX509Identity(mspID, certificate)
}

func NewSign(keyPath string) (identity.Sign, error) {
	privateKeyPEM, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	privateKey, err := identity.PrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	return identity.NewPrivateKeySign(privateKey)
}

type Transaction struct {
	txType      TransactionType
	name        string
	options     []client.ProposalOption
	offlineSign identity.Sign
	result      []byte
}

func (transaction *Transaction) AddOptions(options ...client.ProposalOption) {
	transaction.options = append(transaction.options, options...)
}

var (
	fabricRunning     = false
	channelsJoined    = false
	runningChaincodes = make(map[string]string)
	connectedGateways = make(map[string]*GatewayConnection)
	currentGateway    *GatewayConnection
)

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.AfterSuite(func() {
		if os.Getenv("GATEWAY_NO_SHUTDOWN") != "TRUE" {
			stopFabric()
		}
	})
}

func InitializeScenario(s *godog.ScenarioContext) {
	s.Step(`^I create a gateway named (\S+) for user (\S+) in MSP (\S+)$`, createGateway)
	s.Step(`^I create a gateway named (\S+) without signer for user (\S+) in MSP (\S+)$`, createGatewayWithoutSigner)
	s.Step(`^I connect the gateway to (\S+)$`, connectGateway)
	s.Step(`^I use the gateway named (\S+)$`, useGateway)
	s.Step(`^I deploy (\S+) chaincode named (\S+) at version (\S+) for all organizations on channel (\S+) with endorsement policy (.+)$`, deployChaincode)
	s.Step(`^I have created and joined all channels$`, createAndJoinChannels)
	s.Step(`^I have deployed a Fabric network$`, haveFabricNetwork)
	s.Step(`^I prepare to (submit|evaluate) an? (\S+) transaction$`, prepareTransaction)
	s.Step(`^I set the transaction arguments? to (.+)$`, setArguments)
	s.Step(`^I set transient data on the transaction to$`, setTransientData)
	s.Step(`^I set the endorsing organizations? to (.+)$`, setEndorsingOrgs)
	s.Step(`^I do off-line signing as user (\S+) in MSP (\S+)$`, useOfflineSigner)
	s.Step(`^I invoke the transaction$`, invokeTransaction)
	s.Step(`^I use the (\S+) contract$`, useContract)
	s.Step(`^I use the (\S+) network$`, useNetwork)
	s.Step(`^the response should be JSON matching$`, theResponseShouldBeJSONMatching)
	s.Step(`^I stop the peer named (\S+)$`, stopPeer)
	s.Step(`^I start the peer named (\S+)$`, startPeer)
	s.Step(`^the response should be "(.*)"$`, theResponseShouldBe)
	s.Step(`^the transaction invocation should fail$`, theTransactionShouldFail)
}

func startFabric() error {
	if !fabricRunning {
		fmt.Println("startFabric")
		err := createCryptoMaterial()
		if err != nil {
			return err
		}
		cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "-p", "node", "up", "-d")
		cmd.Dir = dockerComposeDir
		out, err := cmd.CombinedOutput()
		if out != nil {
			fmt.Println(string(out))
		}
		if err != nil {
			return err
		}
		fabricRunning = true
		time.Sleep(20 * time.Second)
	} else {
		fmt.Println("Fabric already running")
	}

	return nil
}

func stopFabric() error {
	if fabricRunning {
		fmt.Println("stopFabric")
		cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "-p", "node", "down")
		cmd.Dir = dockerComposeDir
		out, err := cmd.CombinedOutput()
		if out != nil {
			fmt.Println(string(out))
		}
		if err != nil {
			return err
		}
		fabricRunning = false
	}
	return nil
}

func createCryptoMaterial() error {
	fmt.Println("createCryptoMaterial")
	cmd := exec.Command("./generate.sh")
	cmd.Dir = fixturesDir
	out, err := cmd.CombinedOutput()
	if out != nil {
		fmt.Println(string(out))
	}
	if err != nil {
		return err
	}
	return nil
}

func deployChaincode(ccType, ccName, version, channelName, signaturePolicy string) error {
	fmt.Println("deployChaincode")
	exists := false
	sequence := "1"
	mangledName := ccName + version + channelName
	if policy, ok := runningChaincodes[mangledName]; ok {
		if policy == signaturePolicy {
			return nil
		}
		// Already exists but different signature policy...
		// No need to re-install, just increment the sequence number and approve/commit new signature policy
		exists = true
		out, err := dockerCommandWithTLS(
			"exec", "org1_cli", "peer", "lifecycle", "chaincode", "querycommitted",
			"-o", "orderer.example.com:7050", "--channelID", channelName, "--name", ccName,
		)
		if err != nil {
			return err
		}

		pattern := regexp.MustCompile(".*Sequence: ([0-9]+),.*")
		match := pattern.FindStringSubmatch(out)
		if len(match) != 2 {
			return fmt.Errorf("cannot find installed chaincode for Org1")
		}
		i, err := strconv.Atoi(match[1])
		if err != nil {
			return err
		}
		sequence = fmt.Sprintf("%d", i+1)
	}

	ccPath := "github.com/chaincode/" + ccType + "/" + ccName
	if ccType != "golang" {
		ccPath = "/opt/gopath/src/" + ccPath
	}
	ccLabel := ccName + "v" + version
	ccPackage := ccName + ".tar.gz"

	// is there a collections_config.json file?
	var collectionConfig []string
	collectionFile := "/chaincode/" + ccType + "/" + ccName + "/collections_config.json"
	if _, err := os.Stat(fixturesDir + collectionFile); err == nil {
		collectionConfig = []string{
			"--collections-config",
			"/opt/gopath/src/github.com" + collectionFile,
		}
	}

	for _, org := range orgs {
		if !exists {
			_, err := dockerCommand(
				"exec", org.cli, "peer", "lifecycle", "chaincode", "package", ccPackage,
				"--lang", ccType,
				"--label", ccLabel,
				"--path", ccPath,
			)
			if err != nil {
				return err
			}

			for _, peer := range org.peers {
				env := "CORE_PEER_ADDRESS=" + peer
				_, err := dockerCommand(
					"exec", "-e", env, org.cli, "peer", "lifecycle", "chaincode", "install", ccPackage,
				)
				if err != nil {
					return err
				}
			}
		}

		out, err := dockerCommand(
			"exec", org.cli, "peer", "lifecycle", "chaincode", "queryinstalled",
		)
		if err != nil {
			return err
		}

		pattern := regexp.MustCompile(".*Package ID: (.*), Label: " + ccLabel + ".*")
		match := pattern.FindStringSubmatch(out)
		if len(match) != 2 {
			return fmt.Errorf("cannot find installed chaincode for Org1")
		}
		packageID := match[1]

		approveCommand := []string{
			"exec", org.cli, "peer", "lifecycle", "chaincode", "approveformyorg",
			"--package-id", packageID,
			"--channelID", channelName,
			"--orderer", "orderer.example.com:7050",
			"--name", ccName,
			"--version", version,
			"--signature-policy", signaturePolicy,
			"--sequence", sequence,
			"--waitForEvent",
		}
		approveCommand = append(approveCommand, collectionConfig...)
		_, err = dockerCommandWithTLS(approveCommand...)
		if err != nil {
			return err
		}
	}

	// commit
	commitCommand := []string{
		"exec", "org1_cli", "peer", "lifecycle", "chaincode", "commit",
		"--channelID", channelName,
		"--orderer", "orderer.example.com:7050",
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
		"/etc/hyperledger/configtx/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt",
	}
	commitCommand = append(commitCommand, collectionConfig...)
	_, err := dockerCommandWithTLS(commitCommand...)
	if err != nil {
		return err
	}

	runningChaincodes[mangledName] = signaturePolicy
	time.Sleep(10 * time.Second)

	return nil
}

func createGateway(name string, user string, mspID string) error {
	connection, err := NewGatewayConnectionWithSigner(user, mspID)
	if err != nil {
		return err
	}

	currentGateway = connection
	connectedGateways[name] = connection
	return nil
}

func createGatewayWithoutSigner(name string, user string, mspID string) error {
	connection, err := NewGatewayConnection(user, mspID)
	if err != nil {
		return err
	}

	currentGateway = connection
	connectedGateways[name] = connection
	return nil
}

func connectGateway(peer string) error {
	conn, ok := peerConnectionInfo[peer]
	if !ok {
		return fmt.Errorf("no connection info found for peer: %s", peer)
	}

	certificate, err := loadX509Cert(conn.tlsRootCertPath)
	if err != nil {
		return err
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(certificate)

	url := conn.host + ":" + strconv.FormatUint(uint64(conn.port), 10)

	transportCredentials := credentials.NewClientTLSFromCert(certPool, conn.serverNameOverride)
	clientConn, err := grpc.Dial(url, grpc.WithTransportCredentials(transportCredentials))
	if err != nil {
		return err
	}

	currentGateway.AddOptions(client.WithClientConnection(clientConn))

	return currentGateway.Connect()
}

func useGateway(name string) error {
	var ok bool
	if currentGateway, ok = connectedGateways[name]; !ok {
		return fmt.Errorf("no gateway found: %s", name)
	}
	return nil
}

func loadX509Cert(certFile string) (*x509.Certificate, error) {
	cf, e := ioutil.ReadFile(certFile)
	if e != nil {
		return nil, e
	}

	cpb, _ := pem.Decode(cf)
	crt, e := x509.ParseCertificate(cpb.Bytes)

	if e != nil {
		return nil, e
	}
	return crt, nil
}

func createAndJoinChannels() error {
	fmt.Println("createAndJoinChannels")
	startAllPeers()
	if !channelsJoined {
		_, err := dockerCommandWithTLS(
			"exec", "org1_cli", "peer", "channel", "create",
			"-o", "orderer.example.com:7050",
			"-c", "mychannel",
			"-f", "/etc/hyperledger/configtx/channel.tx",
			"--outputBlock", "/etc/hyperledger/configtx/mychannel.block",
		)
		if err != nil {
			return err
		}

		for _, org := range orgs {
			for _, peer := range org.peers {
				env := "CORE_PEER_ADDRESS=" + peer
				_, err = dockerCommandWithTLS(
					"exec", "-e", env, org.cli, "peer", "channel", "join",
					"-b", "/etc/hyperledger/configtx/mychannel.block",
				)
				if err != nil {
					return err
				}
			}
			_, err = dockerCommandWithTLS(
				"exec", org.cli, "peer", "channel", "update",
				"-o", "orderer.example.com:7050",
				"-c", "mychannel",
				"-f", org.anchortx,
			)
			if err != nil {
				return err
			}
		}

		channelsJoined = true
		time.Sleep(10 * time.Second)

	}
	return nil
}

func stopPeer(peer string) error {
	_, err := dockerCommand(
		"stop", peer,
	)
	if err != nil {
		return err
	}
	peerConnectionInfo[peer].running = false
	return nil
}

func startPeer(peer string) error {
	_, err := dockerCommand(
		"start", peer,
	)
	if err != nil {
		return err
	}
	peerConnectionInfo[peer].running = true
	time.Sleep(20 * time.Second)
	return nil
}

func startAllPeers() error {
	fmt.Println("startAllPeers")
	startedPeers := false
	for peer, info := range peerConnectionInfo {
		if !info.running {
			if _, err := dockerCommand("start", peer); err != nil {
				return err
			}
			peerConnectionInfo[peer].running = true
			startedPeers = true
		}
	}

	if startedPeers {
		time.Sleep(20 * time.Second)
	}
	return nil
}

func dockerCommandWithTLS(args ...string) (string, error) {
	tlsOptions := []string{
		"--tls",
		//"true",
		"--cafile",
		"/etc/hyperledger/configtx/crypto-config/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem",
	}

	fullArgs := append(args, tlsOptions...)
	return dockerCommand(fullArgs...)
}

func dockerCommand(args ...string) (string, error) {
	fmt.Println("\033[1m", ">", "docker", strings.Join(args, " "), "\033[0m")
	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	if out != nil {
		fmt.Println(string(out))
	}
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, string(out))
	}

	return string(out), nil
}

func haveFabricNetwork() error {
	if !fabricRunning {
		return startFabric()
	}
	return nil
}

func prepareTransaction(action string, txnName string) error {
	txnType := Submit
	if action == "evaluate" {
		txnType = Evaluate
	}

	currentGateway.transaction = &Transaction{
		txType: txnType,
		name:   txnName,
	}
	return nil
}

func setArguments(argsJSON string) error {
	args, err := unmarshalArgs(argsJSON)
	if err != nil {
		return err
	}

	currentGateway.transaction.AddOptions(client.WithArguments(args...))

	return nil
}

func useOfflineSigner(user string, mspID string) error {
	keyPath := PrivateKeyPath(user, mspID)
	offlineSign, err := NewSign(keyPath)
	if err != nil {
		return err
	}

	currentGateway.transaction.offlineSign = offlineSign

	return nil
}

func unmarshalArgs(argsJSON string) ([]string, error) {
	var args []string
	err := json.Unmarshal([]byte(argsJSON), &args)
	if err != nil {
		return nil, err
	}

	return args, nil
}

func setTransientData(table *messages.PickleStepArgument_PickleTable) error {
	transient := make(map[string][]byte)
	for _, row := range table.Rows {
		transient[row.Cells[0].Value] = []byte(row.Cells[1].Value)
	}

	currentGateway.transaction.AddOptions(client.WithTransient(transient))
	return nil
}

func setEndorsingOrgs(argsJSON string) error {
	args, err := unmarshalArgs(argsJSON)
	if err != nil {
		return err
	}

	currentGateway.transaction.AddOptions(client.WithEndorsingOrganizations(args...))
	return nil
}

func invokeTransaction() error {
	var err error
	currentGateway.transaction.result, err = transactionInvokeFn(currentGateway.transaction.txType)()
	if err != nil {
		if s, ok := status.FromError(err); ok {
			fmt.Printf("Error details: %+v\n", s.Details())
		}
		return err
	}
	return nil
}

func theTransactionShouldFail() error {
	var err error
	currentGateway.transaction.result, err = transactionInvokeFn(currentGateway.transaction.txType)()
	if nil == err {
		return fmt.Errorf("transaction invocation was expected to fail, but it returned: %s", currentGateway.transaction.result)
	}
	if s, ok := status.FromError(err); ok {
		fmt.Printf("Error details: %+v\n", s.Details())
	}
	return nil
}

func transactionInvokeFn(txType TransactionType) func() ([]byte, error) {
	switch txType {
	case Submit:
		return invokeSubmit
	case Evaluate:
		return invokeEvaluate
	default:
		panic(fmt.Sprintf("unknown transaction type: %v", txType))
	}
}

type Signable interface {
	Digest() []byte
	Bytes() ([]byte, error)
}

func invokeEvaluate() ([]byte, error) {
	proposal, err := currentGateway.contract.NewProposal(currentGateway.transaction.name, currentGateway.transaction.options...)
	if err != nil {
		return nil, err
	}

	proposal, err = offlineSignProposal(proposal)
	if err != nil {
		return nil, err
	}

	return proposal.Evaluate()
}

func invokeSubmit() ([]byte, error) {
	proposal, err := currentGateway.contract.NewProposal(currentGateway.transaction.name, currentGateway.transaction.options...)
	if err != nil {
		return nil, err
	}

	proposal, err = offlineSignProposal(proposal)
	if err != nil {
		return nil, err
	}

	unsignedTransaction, err := proposal.Endorse()
	if err != nil {
		return nil, err
	}

	signedTransaction, err := offlineSignTransaction(unsignedTransaction)
	if err != nil {
		return nil, err
	}

	result := signedTransaction.Result()

	unsignedCommit, err := signedTransaction.Submit()
	if err != nil {
		return result, err
	}

	signedCommit, err := offlineSignCommit(unsignedCommit)
	if err != nil {
		return result, err
	}

	status, err := signedCommit.Status()
	if err != nil {
		return result, err
	}
	if status != peer.TxValidationCode_VALID {
		return result, fmt.Errorf("commit failed with status %v (%v)", status, peer.TxValidationCode_name[int32(status)])
	}

	return result, nil
}

func offlineSignProposal(proposal *client.Proposal) (*client.Proposal, error) {
	if nil == currentGateway.transaction.offlineSign {
		return proposal, nil
	}

	bytes, signature, err := bytesAndSignature(proposal)
	if err != nil {
		return nil, err
	}

	proposal, err = currentGateway.contract.NewSignedProposal(bytes, signature)
	if err != nil {
		return nil, err
	}

	return proposal, nil
}

func offlineSignTransaction(clientTransaction *client.Transaction) (*client.Transaction, error) {
	if nil == currentGateway.transaction.offlineSign {
		return clientTransaction, nil
	}

	bytes, signature, err := bytesAndSignature(clientTransaction)
	if err != nil {
		return nil, err
	}

	clientTransaction, err = currentGateway.contract.NewSignedTransaction(bytes, signature)
	if err != nil {
		return nil, err
	}

	return clientTransaction, nil
}

func offlineSignCommit(commit *client.Commit) (*client.Commit, error) {
	if nil == currentGateway.transaction.offlineSign {
		return commit, nil
	}

	bytes, signature, err := bytesAndSignature(commit)
	if err != nil {
		return nil, err
	}

	commit, err = currentGateway.network.NewSignedCommit(bytes, signature)
	if err != nil {
		return nil, err
	}

	return commit, nil
}

func bytesAndSignature(signable Signable) ([]byte, []byte, error) {
	digest := signable.Digest()
	bytes, err := signable.Bytes()
	if err != nil {
		return nil, nil, err
	}

	signature, err := currentGateway.transaction.offlineSign(digest)
	if err != nil {
		return nil, nil, err
	}

	return bytes, signature, nil
}

func useContract(contractName string) error {
	currentGateway.contract = currentGateway.network.GetContract(contractName)
	return nil
}

func useNetwork(channelName string) error {
	currentGateway.network = currentGateway.gateway.GetNetwork(channelName)
	return nil
}

func theResponseShouldBeJSONMatching(arg *messages.PickleStepArgument_PickleDocString) error {
	same, err := jsonEqual([]byte(arg.GetContent()), currentGateway.transaction.result)
	if err != nil {
		return err
	}
	if !same {
		return fmt.Errorf("transaction response doesn't match expected value. Got: %s", string(currentGateway.transaction.result))
	}
	return nil
}

func jsonEqual(a, b []byte) (bool, error) {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(j2, j), nil
}

func theResponseShouldBe(expected string) error {
	actual := string(currentGateway.transaction.result)
	if actual != expected {
		return fmt.Errorf("transaction response \"%s\" does not match expected value \"%s\"", actual, expected)
	}
	return nil
}
