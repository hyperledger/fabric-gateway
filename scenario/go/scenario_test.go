/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package scenario

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages-go/v10"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/connection"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
)

const (
	fixturesDir       = "../fixtures"
	dockerComposeFile = "docker-compose-tls.yaml"
	dockerComposeDir  = fixturesDir + "/docker-compose"
	//signaturePolicy = `OR("Org3MSP.member",AND("Org1MSP.member","Org2MSP.member"))`
	//signaturePolicy = `OutOf(2, "Org1MSP.member", "Org2MSP.member", "Org3MSP.member")`
)

type TransactionType int

type orgConfig struct {
	cli string
	anchortx string
	peers []string
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

var runningPeers = map[string]bool{
	"peer0.org1.example.com": true,
	"peer1.org1.example.com": true,
	"peer0.org2.example.com": true,
	"peer1.org2.example.com": true,
	"peer0.org3.example.com": true,
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
	}

	return _mspToOrgMap[mspID]
}

type GatewayConnection struct {
	ID      identity.Identity
	options []client.ConnectOption
}

func NewGatewayConnection(user string, mspID string) (*GatewayConnection, error) {
	org := GetOrgForMSP(mspID)

	credentialsDir := fixturesDir + "/crypto-material/crypto-config/peerOrganizations/" + org + "/users/" +
		user + "@" + org + "/msp"
	certPath := credentialsDir + "/signcerts/" + user + "@" + org + "-cert.pem"
	keyPath := credentialsDir + "/keystore/key.pem"

	id, err := newIdentity(mspID, certPath)
	if err != nil {
		return nil, err
	}

	sign, err := newSign(keyPath)
	if err != nil {
		return nil, err
	}

	connection := &GatewayConnection{
		ID: id,
	}
	connection.AddOptions(client.WithSign(sign))

	return connection, nil
}

func (connection *GatewayConnection) AddOptions(options ...client.ConnectOption) {
	connection.options = append(connection.options, options...)
}

func (connection *GatewayConnection) Connect() (*client.Gateway, error) {
	return client.Connect(connection.ID, connection.options...)
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

func newSign(keyPath string) (identity.Sign, error) {
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
	txType  TransactionType
	name    string
	options []client.ProposalOption
}

func (transaction *Transaction) AddOptions(options ...client.ProposalOption) {
	transaction.options = append(transaction.options, options...)
}

var fabricRunning bool = false
var channelsJoined bool = false
var runningChaincodes = make(map[string]bool)
var gatewayProcess *exec.Cmd
var gatewayConnection *GatewayConnection
var gateway *client.Gateway
var network *client.Network
var contract *client.Contract
var transaction *Transaction
var transactionResult []byte

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.AfterSuite(func() {
		stopFabric()
	})
}

func InitializeScenario(s *godog.ScenarioContext) {
	s.Step(`^I create a gateway for user (\S+) in MSP (\S+)`, createGateway)
	s.Step(`^I connect the gateway to (\S+)$`, connectGateway)
	s.Step(`^I deploy (\S+) chaincode named (\S+) at version (\S+) for all organizations on channel (\S+) with endorsement policy (.+)$`, deployChaincode)
	s.Step(`^I have created and joined all channels from the tls connection profile$`, createAndJoinChannels)
	s.Step(`^I have deployed a (\S+) Fabric network$`, haveFabricNetwork)
	s.Step(`^I prepare to (submit|evaluate) an? (\S+) transaction$`, prepareSubmit)
	s.Step(`^I set the transaction arguments? to (.+)$`, setArguments)
	s.Step(`^I set transient data on the transaction to$`, setTransientData)
	s.Step(`^I invoke the transaction$`, invokeTransaction)
	s.Step(`^I invoke the transaction which I expect to fail$`, invokeTransactionFail)
	s.Step(`^I use the (\S+) contract$`, useContract)
	s.Step(`^I use the (\S+) network$`, useNetwork)
	s.Step(`^the response should equal (.+)$`, theResponseShouldEqual)
	s.Step(`^the response should be JSON matching$`, theResponseShouldBeJSONMatching)
	s.Step(`^I stop the peer named (\S+)$`, stopPeer)
	s.Step(`^I start the peer named (\S+)$`, startPeer)
}

func startFabric() error {
	if !fabricRunning {
		err := createCryptoMaterial()
		if err != nil {
			return err
		}
		cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "-p", "node", "up", "-d")
		cmd.Dir = dockerComposeDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		fabricRunning = true
		time.Sleep(20 * time.Second)
	} else {
		fmt.Println("Fabric already running")
	}

	return nil
}

func stopFabric() error {
	if fabricRunning {
		cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "-p", "node", "down")
		cmd.Dir = dockerComposeDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		fabricRunning = false
	}
	return nil
}

func createCryptoMaterial() error {
	cmd := exec.Command("./generate.sh")
	cmd.Dir = fixturesDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func deployChaincode(ccType, ccName, version, channelName, signaturePolicy string) error {
	mangledName := ccName + version + channelName
	if _, ok := runningChaincodes[mangledName]; ok {
		// already exists
		return nil
	}

	ccPath := "github.com/chaincode/" + ccType + "/" + ccName
	if ccType != "golang" {
		ccPath = "/opt/gopath/src/" + ccPath
	}
	ccLabel := ccName + "v" + version
	ccPackage := ccName + ".tar.gz"

	for _, org := range orgs {
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
			_, err = dockerCommand(
				"exec", "-e", env, org.cli, "peer", "lifecycle", "chaincode", "install", ccPackage,
			)
			if err != nil {
				return err
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

		_, err = dockerCommandWithTLS(
			"exec", org.cli, "peer", "lifecycle", "chaincode", "approveformyorg",
			"--package-id", packageID,
			"--channelID", channelName,
			"--name", ccName,
			"--version", version,
			"--signature-policy", signaturePolicy,
			"--sequence", "1",
			"--waitForEvent",
		)
		if err != nil {
			return err
		}
	}

	// commit
	_, err := dockerCommandWithTLS(
		"exec", "org1_cli", "peer", "lifecycle", "chaincode", "commit",
		"--channelID", channelName,
		"--name", ccName,
		"--version", version,
		"--signature-policy", signaturePolicy,
		"--sequence", "1",
		"--waitForEvent",
		"--peerAddresses", "peer0.org1.example.com:7051",
		"--peerAddresses", "peer0.org2.example.com:8051",
		//"--peerAddresses", "peer0.org3.example.com:11051",
		"--tlsRootCertFiles",
		"/etc/hyperledger/configtx/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt",
		"--tlsRootCertFiles",
		"/etc/hyperledger/configtx/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt",
		//"--tlsRootCertFiles",
		//"/etc/hyperledger/configtx/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt",
	)
	if err != nil {
		return err
	}

	runningChaincodes[mangledName] = true
	time.Sleep(10 * time.Second)

	return nil
}

func createGateway(user string, mspID string) error {
	connection, err := NewGatewayConnection(user, mspID)
	if err != nil {
		return err
	}

	gatewayConnection = connection
	gateway = nil
	return nil
}

func connectGateway(address string) error {
	hostPort := strings.Split(address, ":")
	if len(hostPort) != 2 {
		return fmt.Errorf("invalid endpoint: %s", address)
	}

	host := hostPort[0]
	port, err := strconv.ParseUint(hostPort[1], 10, 16)
	if err != nil {
		return err
	}

	endpoint := &connection.Endpoint{
		Host: host,
		Port: uint16(port),
	}
	gatewayConnection.AddOptions(client.WithEndpoint(endpoint))

	gw, err := gatewayConnection.Connect()
	if err != nil {
		return err
	}

	gateway = gw
	return nil
}

func createAndJoinChannels() error {
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
	runningPeers[peer] = false
	return nil
}

func startPeer(peer string) error {
	_, err := dockerCommand(
		"start", peer,
	)
	if err != nil {
		return err
	}
	runningPeers[peer] = true
	time.Sleep(20 * time.Second)
	return nil
}

func startAllPeers() error {
	for peer, running := range runningPeers {
		if !running {
			_, err := dockerCommand(
				"start", peer,
			)
			if err != nil {
				return err
			}
			runningPeers[peer] = true
		}
	}
	time.Sleep(20 * time.Second)
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
	cmd := exec.Command("docker", args...)
	out, err := cmd.CombinedOutput()
	fmt.Println(string(out))
	if err != nil {
		return "", fmt.Errorf("%s: %w", string(out), err)
	}

	return string(out), nil
}

func haveFabricNetwork(tlsType string) error {
	if !fabricRunning {
		return startFabric()
	}
	return nil
}

func prepareSubmit(action string, txnName string) error {
	txnType := Submit
	if action == "evaluate" {
		txnType = Evaluate
	}

	transaction = &Transaction{
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

	transaction.AddOptions(client.WithStringArguments(args...))

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

	transaction.AddOptions(client.WithTransient(transient))
	return nil
}

func invokeTransaction() error {
	invoke, err := transactionInvokeFn(transaction.txType)
	if err != nil {
		return err
	}

	transactionResult, err = invoke(transaction.name, transaction.options...)
	if err != nil {
		return err
	}
	return nil
}

func invokeTransactionFail() error {
	invoke, err := transactionInvokeFn(transaction.txType)
	if err != nil {
		return err
	}

	transactionResult, err = invoke(transaction.name, transaction.options...)
	if err == nil {
		return fmt.Errorf("transaction invocation was expected to fail, but it didn't")
	}
	return nil
}

func transactionInvokeFn(txType TransactionType) (func(string, ...client.ProposalOption) ([]byte, error), error) {
	switch txType {
	case Submit:
		return contract.SubmitSync, nil
	case Evaluate:
		return contract.Evaluate, nil
	default:
		return nil, fmt.Errorf("unknown transaction type: %v", txType)
	}
}

func useContract(contractName string) error {
	contract = network.GetContract(contractName)
	return nil
}

func useNetwork(channelName string) error {
	network = gateway.GetNetwork(channelName)
	return nil
}

func theResponseShouldEqual(expected string) error {
	if expected != string(transactionResult) {
		return fmt.Errorf("transaction response doesn't match expected value")
	}
	return nil
}

func theResponseShouldBeJSONMatching(arg *messages.PickleStepArgument_PickleDocString) error {
	same, err := jsonEqual([]byte(arg.GetContent()), transactionResult)
	if err != nil {
		return err
	}
	if !same {
		return fmt.Errorf("transaction response doesn't match expected value")
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
