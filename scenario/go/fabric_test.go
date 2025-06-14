// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package scenario

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	fixturesDir       = "../fixtures"
	dockerComposeFile = "docker-compose-tls.yaml"
	dockerComposeDir  = fixturesDir + "/docker-compose"
)

type orgConfig struct {
	cli   string
	peers []string
}

type ordererConfig struct {
	address string
	port    string
}

var orgs = []orgConfig{
	{
		cli:   "org1_cli",
		peers: []string{"peer0.org1.example.com:7051", "peer1.org1.example.com:9051"},
	},
	{
		cli:   "org2_cli",
		peers: []string{"peer0.org2.example.com:8051", "peer1.org2.example.com:10051"},
	},
	{
		cli:   "org3_cli",
		peers: []string{"peer0.org3.example.com:11051"},
	},
}

var orderers = []ordererConfig{
	{address: "orderer1.example.com", port: "7053"},
	{address: "orderer2.example.com", port: "8053"},
	{address: "orderer3.example.com", port: "9053"},
}

type peerConnectionInfo struct {
	host               string
	port               uint16
	serverNameOverride string
	tlsRootCertPath    string
	running            bool
}

var peerConnectionInfos = map[string]*peerConnectionInfo{
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

var (
	fabricRunning     = false
	channelsJoined    = false
	runningChaincodes = make(ChaincodeSet)
)

type ChaincodeSet map[string]string

func (set ChaincodeSet) policy(chaincodeName string, version string, channelName string) (policy string, exists bool) {
	key := chaincodeKey(chaincodeName, version, channelName)
	policy, exists = set[key]
	return
}

func chaincodeKey(chaincodeName string, version string, channelName string) string {
	return chaincodeName + version + channelName
}

func (set ChaincodeSet) add(chaincodeName string, version string, channelName string, signaturePolicy string) {
	key := chaincodeKey(chaincodeName, version, channelName)
	set[key] = signaturePolicy
}

func startFabric() error {
	if !fabricRunning {
		fmt.Println("startFabric")
		err := createCryptoMaterial()
		if err != nil {
			return err
		}
		cmd := exec.Command("docker", "compose", "-f", dockerComposeFile, "-p", "node", "up", "-d")
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
		cmd := exec.Command("docker", "compose", "-f", dockerComposeFile, "-p", "node", "down")
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

func generateHSMUser(hsmUserid string) error {
	fmt.Println("generateHSMUser")
	cmd := exec.Command("./generate-hsm-user.sh", hsmUserid)
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

	sequence := "1"

	if policy, exists := runningChaincodes.policy(ccName, version, channelName); exists {
		if policy == signaturePolicy {
			// Nothing to do as already deployed with correct signature policy.
			return nil
		}

		// Already exists but different signature policy...
		// No need to re-install, just increment the sequence number and approve/commit new signature policy.
		currentSequence, err := committedSequenceNumber(ccName, channelName)
		if err != nil {
			return err
		}
		sequence = strconv.Itoa(currentSequence + 1)
	} else {
		if err := installChaincode(ccName, ccType, version); err != nil {
			return err
		}
	}

	// is there a collections_config.json file?
	var collectionConfig []string
	collectionFile := "/chaincode/" + ccType + "/" + ccName + "/collections_config.json"
	if _, err := os.Stat(fixturesDir + collectionFile); err == nil {
		collectionConfig = []string{
			"--collections-config",
			"/opt/gopath/src/github.com" + collectionFile,
		}
	}

	if err := approveChaincode(ccName, version, sequence, channelName, signaturePolicy, collectionConfig); err != nil {
		return err
	}

	if err := commitChaincode(ccName, version, sequence, channelName, signaturePolicy, collectionConfig); err != nil {
		return err
	}

	runningChaincodes.add(ccName, version, channelName, signaturePolicy)
	time.Sleep(10 * time.Second)

	return nil
}

func committedSequenceNumber(chaincodeName string, channelName string) (int, error) {
	out, err := dockerCommandWithTLS(
		"exec", "org1_cli", "peer", "lifecycle", "chaincode", "querycommitted",
		"-o", "orderer1.example.com:7050", "--channelID", channelName, "--name", chaincodeName,
	)
	if err != nil {
		return 0, err
	}

	pattern := regexp.MustCompile(".*Sequence: ([0-9]+),.*")
	match := pattern.FindStringSubmatch(out)
	if len(match) != 2 {
		return 0, errors.New("cannot find installed chaincode for Org1")
	}
	i, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, err
	}

	return i, nil
}

func installChaincode(name string, language string, version string) error {
	path := "/opt/gopath/src/github.com/chaincode/" + language + "/" + name
	pkg := name + ".tar.gz"

	for _, org := range orgs {
		_, err := dockerCommand(
			"exec", org.cli, "peer", "lifecycle", "chaincode", "package", pkg,
			"--lang", language,
			"--label", chaincodeLabel(name, version),
			"--path", path,
		)
		if err != nil {
			return err
		}

		for _, peer := range org.peers {
			env := "CORE_PEER_ADDRESS=" + peer
			_, err := dockerCommand(
				"exec", "-e", env, org.cli, "peer", "lifecycle", "chaincode", "install", pkg,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func chaincodeLabel(name string, version string) string {
	return name + "v" + version
}

func approveChaincode(name string, version string, sequence string, channelName string, signaturePolicy string, collectionConfig []string) error {
	for _, org := range orgs {
		out, err := dockerCommand(
			"exec", org.cli, "peer", "lifecycle", "chaincode", "queryinstalled",
		)
		if err != nil {
			return err
		}

		label := chaincodeLabel(name, version)
		pattern := regexp.MustCompile(".*Package ID: (.*), Label: " + label + ".*")
		match := pattern.FindStringSubmatch(out)
		if len(match) != 2 {
			return errors.New("cannot find installed chaincode for Org1")
		}
		packageID := match[1]

		approveCommand := []string{
			"exec", org.cli, "peer", "lifecycle", "chaincode", "approveformyorg",
			"--package-id", packageID,
			"--channelID", channelName,
			"--orderer", "orderer1.example.com:7050",
			"--name", name,
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

	return nil
}

func commitChaincode(name string, version string, sequence string, channelName string, signaturePolicy string, collectionConfig []string) error {
	commitCommand := []string{
		"exec", "org1_cli", "peer", "lifecycle", "chaincode", "commit",
		"--channelID", channelName,
		"--orderer", "orderer1.example.com:7050",
		"--name", name,
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

	return nil
}

func createAndJoinChannels() error {
	fmt.Println("createAndJoinChannels")

	if err := startAllPeers(); err != nil {
		return err
	}

	if channelsJoined {
		return nil
	}

	for _, orderer := range orderers {
		tlsdir := fmt.Sprintf("/etc/hyperledger/configtx/crypto-config/ordererOrganizations/example.com/orderers/%s/tls", orderer.address)
		if _, err := dockerCommand(
			"exec", "org1_cli", "osnadmin", "channel", "join",
			"--channelID", "mychannel",
			"--config-block", "/etc/hyperledger/configtx/mychannel.block",
			"-o", fmt.Sprintf("%s:%s", orderer.address, orderer.port),
			"--ca-file", "/etc/hyperledger/configtx/crypto-config/ordererOrganizations/example.com/tlsca/tlsca.example.com-cert.pem",
			"--client-cert", tlsdir+"/server.crt",
			"--client-key", tlsdir+"/server.key",
		); err != nil {
			return err
		}
	}

	for _, org := range orgs {
		for _, peer := range org.peers {
			env := "CORE_PEER_ADDRESS=" + peer
			if _, err := dockerCommandWithTLS(
				"exec", "-e", env, org.cli, "peer", "channel", "join",
				"-b", "/etc/hyperledger/configtx/mychannel.block",
			); err != nil {
				return err
			}
		}
	}

	channelsJoined = true
	time.Sleep(10 * time.Second)

	return nil
}

func stopPeer(peer string) error {
	_, err := dockerCommand(
		"stop", peer,
	)
	if err != nil {
		return err
	}
	peerConnectionInfos[peer].running = false
	return nil
}

func startPeer(peer string) error {
	_, err := dockerCommand(
		"start", peer,
	)
	if err != nil {
		return err
	}
	peerConnectionInfos[peer].running = true
	time.Sleep(20 * time.Second)
	return nil
}

func startAllPeers() error {
	fmt.Println("startAllPeers")
	startedPeers := false
	for peer, info := range peerConnectionInfos {
		if !info.running {
			if _, err := dockerCommand("start", peer); err != nil {
				return err
			}
			peerConnectionInfos[peer].running = true
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
