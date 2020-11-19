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
	"github.com/pkg/errors"
)

const (
	fixturesDir       = "../fixtures"
	dockerComposeFile = "docker-compose-tls.yaml"
	dockerComposeDir  = fixturesDir + "/docker-compose"
	gatewayDir        = "../../bin"
)

type TransactionType int

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
	s.Step(`^I deploy (\S+) chaincode named (\S+) at version (\S+) for all organizations on channel (\S+) with endorsement policy (\S+) and arguments (.+)$`, deployChaincode)
	s.Step(`^I have created and joined all channels from the tls connection profile$`, createAndJoinChannels)
	s.Step(`^I have deployed a (\S+) Fabric network$`, haveFabricNetwork)
	s.Step(`^I prepare to submit an? (\S+) transaction$`, prepareSubmit)
	s.Step(`^I prepare to evaluate an? (\S+) transaction$`, prepareEvaluate)
	s.Step(`^I set the transaction arguments? to (.+)$`, setArguments)
	s.Step(`^I set transient data on the transaction to$`, setTransientData)
	s.Step(`^I invoke the transaction$`, invokeTransaction)
	s.Step(`^I use the (\S+) contract$`, useContract)
	s.Step(`^I use the (\S+) network$`, useNetwork)
	s.Step(`^the response should be JSON matching$`, theResponseShouldBeJSONMatching)
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

func deployChaincode(ccType, ccName, version, channelName, policyType, argsJSON string) error {
	mangledName := ccName + version + channelName
	if _, ok := runningChaincodes[mangledName]; ok {
		// already exists
		return nil
	}

	// var args []string
	// err := json.Unmarshal([]byte(argsJSON), &args)
	// if err != nil {
	// 	return err
	// }
	// init := map[string]interface{}{
	// 	"function": args[0],
	// 	"Args":     args[1:],
	// }
	// initArg, err := json.Marshal(init)

	ccPath := "/opt/gopath/src/github.com/chaincode/" + ccType + "/" + ccName
	ccLabel := ccName + "v" + version
	ccPackage := ccName + ".tar.gz"

	// org1
	_, err := dockerCommand(
		"exec", "org1_cli", "peer", "lifecycle", "chaincode", "package", ccPackage,
		"--lang", ccType,
		"--label", ccLabel,
		"--path", ccPath,
	)
	if err != nil {
		return err
	}

	_, err = dockerCommand(
		"exec", "org1_cli", "peer", "lifecycle", "chaincode", "install", ccPackage,
	)
	if err != nil {
		return err
	}

	out, err := dockerCommand(
		"exec", "org1_cli", "peer", "lifecycle", "chaincode", "queryinstalled",
	)
	if err != nil {
		return err
	}

	pattern := regexp.MustCompile(".*Package ID: (.*), Label: " + ccLabel + ".*")
	match := pattern.FindStringSubmatch(out)
	if len(match) != 2 {
		return errors.New("Cannot find installed chaincode for Org1")
	}
	packageID := match[1]

	_, err = dockerCommandWithTLS(
		"exec", "org1_cli", "peer", "lifecycle", "chaincode",
		"approveformyorg", "--package-id", packageID, "--channelID", channelName, "--name", ccName,
		"--version", version, "--signature-policy", `AND("Org1MSP.member","Org2MSP.member")`,
		"--sequence", "1", "--waitForEvent",
	)
	if err != nil {
		return err
	}

	// org2
	_, err = dockerCommand(
		"exec", "org2_cli", "peer", "lifecycle", "chaincode", "package", ccPackage,
		"--lang", ccType,
		"--label", ccLabel,
		"--path", ccPath,
	)
	if err != nil {
		return err
	}

	_, err = dockerCommand(
		"exec", "org2_cli", "peer", "lifecycle", "chaincode", "install", ccPackage,
	)
	if err != nil {
		return err
	}

	out, err = dockerCommand(
		"exec", "org2_cli", "peer", "lifecycle", "chaincode", "queryinstalled",
	)
	if err != nil {
		return err
	}

	pattern = regexp.MustCompile(".*Package ID: (.*), Label: " + ccLabel + ".*")
	match = pattern.FindStringSubmatch(out)
	if len(match) != 2 {
		return errors.New("Cannot find installed chaincode for Org2")
	}
	packageID = match[1]

	_, err = dockerCommandWithTLS(
		"exec", "org2_cli", "peer", "lifecycle", "chaincode",
		"approveformyorg", "--package-id", packageID, "--channelID", channelName, "--name", ccName,
		"--version", version, "--signature-policy", `AND("Org1MSP.member","Org2MSP.member")`,
		"--sequence", "1", "--waitForEvent",
	)
	if err != nil {
		return err
	}

	// commit
	_, err = dockerCommandWithTLS(
		"exec", "org1_cli", "peer", "lifecycle", "chaincode",
		"commit", "--channelID", channelName, "--name", ccName, "--version", version,
		"--signature-policy", "AND(\"Org1MSP.member\",\"Org2MSP.member\")", "--sequence", "1",
		"--waitForEvent", "--peerAddresses", "peer0.org1.example.com:7051", "--peerAddresses",
		"peer0.org2.example.com:8051",
		"--tlsRootCertFiles",
		"/etc/hyperledger/configtx/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt",
		"--tlsRootCertFiles",
		"/etc/hyperledger/configtx/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt",
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
		return errors.Errorf("Invalid endpoint: %s", address)
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

		_, err = dockerCommandWithTLS(
			"exec", "org1_cli", "peer", "channel", "join",
			"-b", "/etc/hyperledger/configtx/mychannel.block",
		)
		if err != nil {
			return err
		}

		_, err = dockerCommandWithTLS(
			"exec", "org2_cli", "peer", "channel", "join",
			"-b", "/etc/hyperledger/configtx/mychannel.block",
		)
		if err != nil {
			return err
		}

		_, err = dockerCommandWithTLS(
			"exec", "org1_cli", "peer", "channel", "update",
			"-o", "orderer.example.com:7050",
			"-c", "mychannel",
			"-f", "/etc/hyperledger/configtx/Org1MSPanchors.tx",
		)
		if err != nil {
			return err
		}

		_, err = dockerCommandWithTLS(
			"exec", "org2_cli", "peer", "channel", "update",
			"-o", "orderer.example.com:7050",
			"-c", "mychannel",
			"-f", "/etc/hyperledger/configtx/Org2MSPanchors.tx",
		)
		if err != nil {
			return err
		}

		channelsJoined = true
		time.Sleep(10 * time.Second)

	}
	return nil
}

func dockerCommandWithTLS(args ...string) (string, error) {
	tlsOptions := []string{
		"--tls",
		"true",
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
		return "", errors.Wrap(err, string(out))
	}

	return string(out), nil
}

func haveFabricNetwork(tlsType string) error {
	if !fabricRunning {
		return startFabric()
	}
	return nil
}

func prepareSubmit(txnName string) error {
	transaction = &Transaction{
		txType: Submit,
		name:   txnName,
	}
	return nil
}

func prepareEvaluate(txnName string) error {
	transaction = &Transaction{
		txType: Evaluate,
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

func transactionInvokeFn(txType TransactionType) (func(string, ...client.ProposalOption) ([]byte, error), error) {
	switch txType {
	case Submit:
		return contract.SubmitSync, nil
	case Evaluate:
		return contract.Evaluate, nil
	default:
		return nil, errors.Errorf("Unknown transaction type: %v", txType)
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

func theResponseShouldBeJSONMatching(arg *messages.PickleStepArgument_PickleDocString) error {
	same, err := jsonEqual([]byte(arg.GetContent()), transactionResult)
	if err != nil {
		return err
	}
	if !same {
		return errors.New("Transaction response doesn't match expected value")
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
