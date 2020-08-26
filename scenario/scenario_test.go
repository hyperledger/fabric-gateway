/*
Copyright 2020 IBM All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package scenario

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"syscall"
	"time"

	"github.com/cucumber/godog"
	messages "github.com/cucumber/messages-go/v10"
	"github.com/hyperledger/fabric-gateway/client/go/sdk"
	"github.com/hyperledger/fabric-gateway/pkg/gateway"
	"github.com/pkg/errors"
)

const (
	dockerComposeFile = "docker-compose-tls.yaml"
	dockerComposeDir  = "fixtures/docker-compose"
	gatewayDir        = "../prototype"
)

var fabricRunning bool = false
var channelsJoined bool = false
var runningChaincodes = make(map[string]bool)
var gatewayProcess *exec.Cmd
var gw *sdk.Gateway
var network *sdk.Network
var contract *sdk.Contract
var transaction *sdk.Transaction
var transactionResult []byte

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.AfterSuite(func() {
		stopGateway()
		stopFabric()
	})
}

func InitializeScenario(s *godog.ScenarioContext) {
	s.Step(`^I connect the gateway$`, iConnectTheGateway)
	s.Step(`^I deploy (\w+) chaincode named (\w+) at version ([^ ]+) for all organizations on channel (\w+) with endorsement policy ([^ ]+) and arguments(.+)$`, deployChaincode)
	s.Step(`^I have a gateway for (.+)$`, startGateway)
	s.Step(`^I have a gateway as user User(\d+) using the tls connection profile$`, haveGateway)
	s.Step(`^I have created and joined all channels from the tls connection profile$`, createAndJoinChannels)
	s.Step(`^I have deployed a (\w+) Fabric network$`, haveFabricNetwork)
	s.Step(`^I prepare an? (\w+) transaction$`, prepareTransaction)
	s.Step(`^I set transient data on the transaction to$`, setTransientData)
	s.Step(`^I (submit|evaluate) the transaction with arguments (.+)$`, actionTransaction)
	s.Step(`^I use the (\w+) contract$`, useContract)
	s.Step(`^I use the (\w+) network$`, useNetwork)
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

func startGateway(mspid string) error {
	if gatewayProcess == nil {
		gatewayProcess = exec.Command(
			"go", "run", "gateway.go",
			"-h", "peer0.org1.example.com",
			"-p", "7051",
			"-m", mspid,
			"-cert", "../scenario/fixtures/crypto-material/crypto-config/peerOrganizations/org1.example.com/users/User2@org1.example.com/msp/signcerts/User2@org1.example.com-cert.pem",
			"-key", "../scenario/fixtures/crypto-material/crypto-config/peerOrganizations/org1.example.com/users/User2@org1.example.com/msp/keystore/key.pem",
			"-tlscert", "../scenario/fixtures/crypto-material/crypto-config/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem",
		)
		gatewayProcess.Dir = gatewayDir
		gatewayProcess.Env = append(os.Environ(), "DISCOVERY_AS_LOCALHOST=TRUE")
		gatewayProcess.Stdout = os.Stdout
		gatewayProcess.Stderr = os.Stderr
		gatewayProcess.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		err := gatewayProcess.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

func stopGateway() error {
	if gatewayProcess != nil {
		pgid, err := syscall.Getpgid(gatewayProcess.Process.Pid)
		if err == nil {
			gatewayProcess = nil
			return syscall.Kill(-pgid, 15)
		}
		return err
	}
	return nil
}

func createCryptoMaterial() error {
	cmd := exec.Command("./generate.sh")
	cmd.Dir = "fixtures"
	out, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func iConnectTheGateway() error {
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

func haveGateway(arg1 int) error {
	if gw == nil {
		mspid := "Org1MSP"
		certPath := "fixtures/crypto-material/crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/signcerts/User1@org1.example.com-cert.pem"
		keyPath := "fixtures/crypto-material/crypto-config/peerOrganizations/org1.example.com/users/User1@org1.example.com/msp/keystore/key.pem"
		f, err := ioutil.ReadFile(certPath)
		if err != nil {
			return err
		}
		cert := string(f)
		f, err = ioutil.ReadFile(keyPath)
		if err != nil {
			return err
		}
		key := string(f)

		signer, err := gateway.CreateSigner(
			mspid,
			cert,
			key,
		)

		gw, err = sdk.Connect("localhost:1234", signer)
	}
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

func prepareTransaction(txnName string) error {
	transaction = contract.CreateTransaction(txnName)
	return nil
}

func setTransientData(table *messages.PickleStepArgument_PickleTable) error {
	transient := make(map[string][]byte)
	for _, row := range table.Rows {
		transient[row.Cells[0].Value] = []byte(row.Cells[1].Value)
	}
	transaction.SetTransient(transient)
	return nil
}

func actionTransaction(action, argsJSON string) error {
	var args []string
	err := json.Unmarshal([]byte(argsJSON), &args)
	if err != nil {
		return err
	}

	if action == "submit" {
		transactionResult, err = transaction.Submit(args...)
	} else {
		transactionResult, err = transaction.Evaluate(args...)
	}
	if err != nil {
		return err
	}
	return nil
}

func useContract(contractName string) error {
	contract = network.GetContract(contractName)
	return nil
}

func useNetwork(channelName string) error {
	network = gw.GetNetwork(channelName)
	return nil
}

func theResponseShouldBeJSONMatching(arg *messages.PickleStepArgument_PickleDocString) error {
	same, err := JSONEqual([]byte(arg.GetContent()), transactionResult)
	if err != nil {
		return err
	}
	if !same {
		return errors.New("Transaction response doesn't match expected value")
	}
	return nil
}

func JSONEqual(a, b []byte) (bool, error) {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(j2, j), nil
}
