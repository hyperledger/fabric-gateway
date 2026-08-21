// Copyright IBM Corp. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package scenario

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hyperledger/fabric-admin-sdk/pkg/chaincode"
	"github.com/hyperledger/fabric-admin-sdk/pkg/channel"
	adminidentity "github.com/hyperledger/fabric-admin-sdk/pkg/identity"
	"github.com/hyperledger/fabric-protos-go-apiv2/common"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/proto"
)

const (
	fixturesDir       = "../fixtures"
	dockerComposeFile = "docker-compose-tls.yaml"
	dockerComposeDir  = fixturesDir + "/docker-compose"

	chaincodeDir       = fixturesDir + "/chaincode"
	cryptoMaterialDir  = fixturesDir + "/crypto-material"
	configBlockFile    = cryptoMaterialDir + "/mychannel.block"
	ordererOrgDir      = cryptoMaterialDir + "/crypto-config/ordererOrganizations/example.com"
	ordererCACertFile  = ordererOrgDir + "/tlsca/tlsca.example.com-cert.pem"
	collectionFileName = "collections_config.json"

	// adminUser is the organization user whose credentials are used for Fabric administrative operations.
	adminUser = "Admin"

	installTimeout     = 5 * time.Minute
	transactionTimeout = 2 * time.Minute
)

type orgConfig struct {
	mspID string
	peers []string
}

type ordererConfig struct {
	name      string
	adminPort string
}

var orgs = []orgConfig{
	{
		mspID: "Org1MSP",
		peers: []string{"peer0.org1.example.com", "peer1.org1.example.com"},
	},
	{
		mspID: "Org2MSP",
		peers: []string{"peer0.org2.example.com", "peer1.org2.example.com"},
	},
	{
		mspID: "Org3MSP",
		peers: []string{"peer0.org3.example.com"},
	},
}

var orderers = []ordererConfig{
	{name: "orderer1.example.com", adminPort: "7053"},
	{name: "orderer2.example.com", adminPort: "8053"},
	{name: "orderer3.example.com", adminPort: "9053"},
}

type peerConnectionInfo struct {
	host               string
	port               uint16
	serverNameOverride string
	tlsRootCertPath    string
	running            bool
	healthzPort        uint16
}

var peerConnectionInfos = map[string]*peerConnectionInfo{
	"peer0.org1.example.com": {
		host:               "localhost",
		port:               7051,
		serverNameOverride: "peer0.org1.example.com",
		tlsRootCertPath:    fixturesDir + "/crypto-material/crypto-config/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt",
		running:            true,
		healthzPort:        7443,
	},
	"peer1.org1.example.com": {
		host:               "localhost",
		port:               9051,
		serverNameOverride: "peer1.org1.example.com",
		tlsRootCertPath:    fixturesDir + "/crypto-material/crypto-config/peerOrganizations/org1.example.com/peers/peer1.org1.example.com/tls/ca.crt",
		running:            true,
		healthzPort:        9443,
	},
	"peer0.org2.example.com": {
		host:               "localhost",
		port:               8051,
		serverNameOverride: "peer0.org2.example.com",
		tlsRootCertPath:    fixturesDir + "/crypto-material/crypto-config/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt",
		running:            true,
		healthzPort:        8443,
	},
	"peer1.org2.example.com": {
		host:               "localhost",
		port:               10051,
		serverNameOverride: "peer1.org2.example.com",
		tlsRootCertPath:    fixturesDir + "/crypto-material/crypto-config/peerOrganizations/org2.example.com/peers/peer1.org2.example.com/tls/ca.crt",
		running:            true,
		healthzPort:        10443,
	},
	"peer0.org3.example.com": {
		host:               "localhost",
		port:               11051,
		serverNameOverride: "peer0.org3.example.com",
		tlsRootCertPath:    fixturesDir + "/crypto-material/crypto-config/peerOrganizations/org3.example.com/peers/peer0.org3.example.com/tls/ca.crt",
		running:            true,
		healthzPort:        11443,
	},
}

func GetOrgForMSP(mspID string) string {
	switch mspID {
	case "Org1MSP":
		return "org1.example.com"
	case "Org2MSP":
		return "org2.example.com"
	case "Org3MSP":
		return "org3.example.com"
	default:
		return ""
	}
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

// adminIdentity is the signing identity used to carry out administrative operations on behalf of an organization.
func (org orgConfig) adminIdentity() (adminidentity.SigningIdentity, error) {
	certificate, err := adminidentity.ReadCertificate(certificatePath(adminUser, org.mspID))
	if err != nil {
		return nil, err
	}

	privateKey, err := adminidentity.ReadPrivateKey(PrivateKeyPath(adminUser, org.mspID))
	if err != nil {
		return nil, err
	}

	return adminidentity.NewPrivateKeySigningIdentity(org.mspID, certificate, privateKey)
}

func (orderer ordererConfig) adminURL() string {
	return "https://localhost:" + orderer.adminPort
}

func (orderer ordererConfig) tlsDir() string {
	return ordererOrgDir + "/orderers/" + orderer.name + "/tls"
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
		for peer := range peerConnectionInfos {
			waitForHealthzOK(peer)
		}
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
	cmd := exec.Command("./generate-hsm-user.sh", hsmUserid) //#nosec G204
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

// chaincodeDeployment describes a chaincode to be deployed to a channel.
type chaincodeDeployment struct {
	language        string
	name            string
	version         string
	channelName     string
	signaturePolicy string
}

func deployChaincode(ccType string, ccName string, version string, channelName string, signaturePolicy string) error {
	deployment := &chaincodeDeployment{
		language:        ccType,
		name:            ccName,
		version:         version,
		channelName:     channelName,
		signaturePolicy: signaturePolicy,
	}
	return deployment.deploy()
}

func (d *chaincodeDeployment) deploy() error {
	currentPolicy, deployed := runningChaincodes.policy(d.name, d.version, d.channelName)
	if deployed && currentPolicy == d.signaturePolicy {
		// Nothing to do as already deployed with correct signature policy.
		return nil
	}

	fmt.Printf("Deploy %s chaincode named %s at version %s on channel %s\n", d.language, d.name, d.version, d.channelName)

	chaincodePackage, err := newChaincodePackage(d.language, d.sourceDir(), d.label())
	if err != nil {
		return err
	}

	// A chaincode that has already been deployed only needs a new definition with an incremented sequence number to be
	// approved and committed; its chaincode package is already installed.
	sequence := int64(1)
	if deployed {
		if sequence, err = d.nextSequenceNumber(); err != nil {
			return err
		}
	} else if err := installChaincode(chaincodePackage); err != nil {
		return err
	}

	definition, err := d.newDefinition(chaincodePackage, sequence)
	if err != nil {
		return err
	}

	if err := approveChaincode(definition); err != nil {
		return err
	}

	if err := commitChaincode(definition); err != nil {
		return err
	}

	runningChaincodes.add(d.name, d.version, d.channelName, d.signaturePolicy)

	return nil
}

func (d *chaincodeDeployment) newDefinition(chaincodePackage []byte, sequence int64) (*chaincode.Definition, error) {
	packageID, err := chaincode.PackageID(bytes.NewReader(chaincodePackage))
	if err != nil {
		return nil, err
	}

	applicationPolicy, err := chaincode.NewApplicationPolicy(d.signaturePolicy, "")
	if err != nil {
		return nil, err
	}

	collections, err := readCollectionConfig(d.sourceDir() + "/" + collectionFileName)
	if err != nil {
		return nil, err
	}

	return &chaincode.Definition{
		ChannelName:       d.channelName,
		PackageID:         packageID,
		Name:              d.name,
		Version:           d.version,
		Sequence:          sequence,
		ApplicationPolicy: applicationPolicy,
		Collections:       collections,
	}, nil
}

func (d *chaincodeDeployment) sourceDir() string {
	return chaincodeDir + "/" + d.language + "/" + d.name
}

func (d *chaincodeDeployment) label() string {
	return d.name + "v" + d.version
}

func (d *chaincodeDeployment) nextSequenceNumber() (int64, error) {
	var sequence int64

	err := withOrgGateway(orgs[0], func(gateway *chaincode.Gateway) error {
		ctx, cancel := context.WithTimeout(context.Background(), transactionTimeout)
		defer cancel()

		result, err := gateway.QueryCommittedWithName(ctx, d.channelName, d.name)
		if err != nil {
			return err
		}

		sequence = result.GetSequence() + 1
		return nil
	})

	return sequence, err
}

func installChaincode(chaincodePackage []byte) error {
	fmt.Println("Install chaincode on all peers")

	wg := new(errgroup.Group)

	for _, org := range orgs {
		for _, peer := range org.peers {
			wg.Go(func() error {
				return installChaincodeToPeer(org, peer, chaincodePackage)
			})
		}
	}

	return wg.Wait()
}

func installChaincodeToPeer(org orgConfig, peer string, chaincodePackage []byte) error {
	id, err := org.adminIdentity()
	if err != nil {
		return err
	}

	return withPeerConnection(peer, func(connection *grpc.ClientConn) error {
		ctx, cancel := context.WithTimeout(context.Background(), installTimeout)
		defer cancel()

		if _, err := chaincode.NewPeer(connection, id).Install(ctx, bytes.NewReader(chaincodePackage)); err != nil {
			return fmt.Errorf("failed to install chaincode on peer %s: %w", peer, err)
		}

		return nil
	})
}

func approveChaincode(definition *chaincode.Definition) error {
	fmt.Printf("Approve chaincode named %s at version %s and sequence %d on channel %s\n",
		definition.Name, definition.Version, definition.Sequence, definition.ChannelName)

	wg := new(errgroup.Group)

	for _, org := range orgs {
		wg.Go(func() error {
			return approveChaincodeForOrg(org, definition)
		})
	}

	return wg.Wait()
}

func approveChaincodeForOrg(org orgConfig, definition *chaincode.Definition) error {
	return withOrgGateway(org, func(gateway *chaincode.Gateway) error {
		ctx, cancel := context.WithTimeout(context.Background(), transactionTimeout)
		defer cancel()

		if err := gateway.Approve(ctx, definition); err != nil {
			return fmt.Errorf("failed to approve chaincode for %s: %w", org.mspID, err)
		}

		return nil
	})
}

func commitChaincode(definition *chaincode.Definition) error {
	fmt.Printf("Commit chaincode named %s at version %s and sequence %d on channel %s\n",
		definition.Name, definition.Version, definition.Sequence, definition.ChannelName)

	return withOrgGateway(orgs[0], func(gateway *chaincode.Gateway) error {
		ctx, cancel := context.WithTimeout(context.Background(), transactionTimeout)
		defer cancel()

		return gateway.Commit(ctx, definition)
	})
}

func createAndJoinChannels() error {
	fmt.Println("createAndJoinChannels")

	if channelsJoined {
		return restartAllPeers()
	}

	if _, err := startAllPeers(); err != nil {
		return err
	}

	configBlock, err := readConfigBlock()
	if err != nil {
		return err
	}

	if err := joinOrderers(configBlock); err != nil {
		return err
	}

	if err := joinPeers(configBlock); err != nil {
		return err
	}

	channelsJoined = true

	return nil
}

// readConfigBlock reads the channel configuration block created with the network's crypto material.
func readConfigBlock() (*common.Block, error) {
	blockBytes, err := os.ReadFile(configBlockFile)
	if err != nil {
		return nil, err
	}

	block := &common.Block{}
	if err := proto.Unmarshal(blockBytes, block); err != nil {
		return nil, fmt.Errorf("failed to parse config block %s: %w", configBlockFile, err)
	}

	return block, nil
}

func joinOrderers(configBlock *common.Block) error {
	certPool, err := newCertPool(ordererCACertFile)
	if err != nil {
		return err
	}

	wg := new(errgroup.Group)

	for _, orderer := range orderers {
		wg.Go(func() error {
			return joinOrderer(orderer, configBlock, certPool)
		})
	}

	return wg.Wait()
}

func joinOrderer(orderer ordererConfig, configBlock *common.Block, certPool *x509.CertPool) error {
	clientCertificate, err := tls.LoadX509KeyPair(orderer.tlsDir()+"/server.crt", orderer.tlsDir()+"/server.key")
	if err != nil {
		return err
	}

	response, err := channel.CreateChannel(orderer.adminURL(), configBlock, certPool, clientCertificate)
	if err != nil {
		return fmt.Errorf("failed to join orderer %s to channel: %w", orderer.name, err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(response.Body)
		return fmt.Errorf("failed to join orderer %s to channel: %s: %s", orderer.name, response.Status, body)
	}

	return nil
}

func joinPeers(configBlock *common.Block) error {
	wg := new(errgroup.Group)

	for _, org := range orgs {
		for _, peer := range org.peers {
			wg.Go(func() error {
				return joinPeer(org, peer, configBlock)
			})
		}
	}

	return wg.Wait()
}

func joinPeer(org orgConfig, peer string, configBlock *common.Block) error {
	id, err := org.adminIdentity()
	if err != nil {
		return err
	}

	return withPeerConnection(peer, func(connection *grpc.ClientConn) error {
		ctx, cancel := context.WithTimeout(context.Background(), transactionTimeout)
		defer cancel()

		if err := channel.JoinChannel(ctx, connection, id, configBlock); err != nil {
			return fmt.Errorf("failed to join peer %s to channel: %w", peer, err)
		}

		return nil
	})
}

// withOrgGateway invokes a function with a chaincode gateway that transacts as an organization's admin user, using
// one of the organization's peers as the gateway.
func withOrgGateway(org orgConfig, f func(*chaincode.Gateway) error) error {
	id, err := org.adminIdentity()
	if err != nil {
		return err
	}

	return withPeerConnection(org.peers[0], func(connection *grpc.ClientConn) error {
		return f(chaincode.NewGateway(connection, id))
	})
}

// withPeerConnection invokes a function with a gRPC connection to a specific peer, closing the connection afterwards.
func withPeerConnection(peer string, f func(*grpc.ClientConn) error) error {
	connection, err := newPeerConnection(peer)
	if err != nil {
		return err
	}
	defer func() { _ = connection.Close() }()

	return f(connection)
}

func newPeerConnection(peer string) (*grpc.ClientConn, error) {
	peerInfo, ok := peerConnectionInfos[peer]
	if !ok {
		return nil, fmt.Errorf("no connection info found for peer: %s", peer)
	}

	certPool, err := newCertPool(peerInfo.tlsRootCertPath)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("dns:///%s:%d", peerInfo.host, peerInfo.port)
	transportCredentials := credentials.NewClientTLSFromCert(certPool, peerInfo.serverNameOverride)

	return grpc.NewClient(url, grpc.WithTransportCredentials(transportCredentials))
}

func newCertPool(certificateFile string) (*x509.CertPool, error) {
	certificatePEM, err := os.ReadFile(certificateFile) // #nosec G304
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certificatePEM) {
		return nil, fmt.Errorf("failed to parse TLS certificate: %s", certificateFile)
	}

	return certPool, nil
}

func stopPeer(peer string) error {
	if _, err := dockerCommand("stop", peer); err != nil {
		return err
	}

	peerConnectionInfos[peer].running = false
	return nil
}

func startPeer(peer string) error {
	if _, err := dockerCommand("start", peer); err != nil {
		return err
	}

	peerConnectionInfos[peer].running = true
	waitForHealthzOK(peer)
	return nil
}

func waitForHealthzOK(peer string) {
	peerInfo := peerConnectionInfos[peer]
	healthzURL := fmt.Sprintf("http://%s:%d/healthz", peerInfo.host, peerInfo.healthzPort)

	var health *HealthStatus
	for !health.IsOK() {
		if health != nil {
			time.Sleep(time.Second)
		}

		if current, err := getHealth(healthzURL); err != nil {
			health = new(HealthStatus)
		} else {
			if !current.IsOK() {
				log.Printf("Bad health for peer %s: %v\n", peer, current)
			}
			health = current
		}
	}
}

func getHealth(url string) (*HealthStatus, error) {
	response, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	result := new(HealthStatus)
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}

	return result, nil
}

// FailedCheck represents a failed status check for a component.
type FailedCheck struct {
	Component string `json:"component"`
	Reason    string `json:"reason"`
}

// HealthStatus represents the current health status of all registered components.
type HealthStatus struct {
	Status       string        `json:"status"`
	Time         time.Time     `json:"time"`
	FailedChecks []FailedCheck `json:"failed_checks,omitempty"`
}

func (h *HealthStatus) IsOK() bool {
	return h != nil && h.Status == "OK"
}

func restartAllPeers() error {
	startedPeers, err := startAllPeers()
	if err != nil {
		return err
	}

	if len(startedPeers) > 0 {
		// Give service discovery time to sync after restarting peers
		time.Sleep(5 * time.Second)
	}

	return nil
}

func startAllPeers() ([]string, error) {
	fmt.Println("startAllPeers")

	var startedPeers []string

	for peer, info := range peerConnectionInfos {
		if !info.running {
			if _, err := dockerCommand("start", peer); err != nil {
				return startedPeers, err
			}
			peerConnectionInfos[peer].running = true
			startedPeers = append(startedPeers, peer)
		}
	}

	for _, peer := range startedPeers {
		waitForHealthzOK(peer)
	}

	return startedPeers, nil
}

func dockerCommand(args ...string) (string, error) {
	fmt.Println("\033[1m", ">", "docker", strings.Join(args, " "), "\033[0m")
	cmd := exec.Command("docker", args...) //#nosec G204
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
