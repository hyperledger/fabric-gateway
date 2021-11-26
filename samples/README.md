# Fabric Gateway Samples

The samples in this repository show how to create client applications that invoke transactions using the new embedded
Gateway in Fabric.

# Pre-reqs

Install the following pre-reqs to run the sample client applications:

- Go v1.16 (required for sample Fabric network and Go sample)
- Node 14.x or 16.x (required for Node sample)
- Java 11 (required for Java sample)
- Docker (required for sample Fabric network)
- Make (required to run Makefile targets, used to simplify set up and run of the samples)

In addition, you will need the `godog` tool to use the sample Fabric network, which can be installed with:

```sh
go install github.com/cucumber/godog/cmd/godog@v0.12.1
```

Make sure you can run `godog --version` after installing. If the command is not found, add the Go bin directory to your
path using: 

```sh
export PATH="${PATH}:$(go env GOPATH)/bin"
```

# Setup

The samples will only run against Fabric v2.4. The easiest way of setting up a gateway enabled Fabric network is to use
the scenario test framework that is part of this `fabric-gateway` repository. After cloning the `fabric-gateway`
repository, from the `fabric-gateway` directory run the following command:

```sh
make sample-network
```

This will create a local Docker network comprising five peers across three organisations and a single ordering node.
Each of the peers have been configured with the gateway enabled. Bringing the network up will take several minutes
because chaincodes will be installed and built on each of the peers, and deployed to the test channel.

A simple smart contract (named `basic`) will have been deployed on all the peers. Additionally, a smart contract with
private data (named `private`) will have been deployed on all the peers. The source code for the smart contracts can be
examined for the [basic](https://github.com/hyperledger/fabric-gateway/blob/main/scenario/fixtures/chaincode/golang/basic/main.go)
and [private](https://github.com/hyperledger/fabric-gateway/blob/main/scenario/fixtures/chaincode/golang/private/private.go)
smart contracts.

# Sample client applications

A sample client application is provided for each of the supported SDKs. Note that the SDKs implement the Fabric Gateway
programming model which has been in use since Fabric v1.4, but these are new implementations that target the embedded
peer gateway and they share no common code with existing Fabric SDKs.

In each of the language samples, the client application demonstrates the various transaction submit and evaluate
patterns supported by the gateway.

## Sample scenarios

In each sample a client from Org1MSP interacts with their organization's trusted peer `peer0.org1.example.com` to
gather endorsements for chaincode invocations. The gateway typically works by first executing the chaincode on the
local peer, and then based on the writeset in the execution results (writes to chaincode namespaces, private data
collections, and keys with state-based endorsement policies), the peer checks the endorsement plan from the discovery
service and determines which others organizations are required to endorse the transaction to ensure that the
transaction will ultimately get validated. The gateway then collects the required endorsements from peers that are
available from the required organizations, and finally submits the endorsed transaction for ordering and commit.

The gateway and discovery service work together to automatically collect the required endorsements so that client
applications can focus on application logic rather than Fabric transaction mechanics. The application can however
dictate endorsement logic to the gateway in certain scenarios. For example, it can dictate the set of organizations to
endorse a transaction when known.

Each of the language samples implement the following client application scenarios:

* **exampleSubmit** - Submits a transaction (`put`) to update the ledger followed by
  evaluating a transaction (`get`) to retrieve the value from the ledger (query).
  The value that is being updated and retrieved is the current timestamp to demonstrate that the update is working.
  Since the endorsement policy for the basic smart contract is `AND("Org1MSP.member","Org2MSP.member")`,
  the gateway collects an endorsement from the local Org1MSP peer, identifies the required endorsement
  from Org2MSP, and then collections an endorsement from an available Org2MSP peer.
  In this example the transaction submission blocks until the transaction ultimately gets validated.

* **exampleSubmitAsync** - Same as **exampleSubmit** however the transaction is submitted asynchronsly,
  meaning that control is passed back to the client application after the transaction is submitted for ordering,
  then the application listens for the eventual transaction commit event.

* **exampleSubmitPrivateData** - Submits a transaction that writes to one private data collection
  owned by Org1MSP and Org3MSP, and to another private data collection owned by Org3MSP.
  The gateway collects an endorsement from the local Org1MSP peer and
  then identifies the writes to the Org3MSP collection. Since Org3MSP collection writes must
  be endorsed by Org3MSP, and Org3MSP is a member of both collections, the gateway automatically gathers
  another endorsement from an Org3MSP peer.

* **exampleSubmitPrivateData2** - Similar to **exampleSubmitPrivateData** except one private data collection
  is owned by Org1MSP and the other private data collection is owned by Org3MSP. The gateway will
  not automatically gather an endorsement from an Org3MSP peer since that may leak Org1MSP
  private data, so in this case the client application must dictate that the gateway should
  gather endorsements from Org1MSP and Org3MSP.

* **exampleStateBasedEndorsement** - In this scenario a key is written with a state-based endorsement
  policy of Org1MSP. The value is then updated and gateway endorses on Org1MSP and identifies that Org1MSP
  is the only required endorser. The state-based endorsement policy is then changed to Org2MSP and
  Org3MSP, and again gateway identifies that Org1MSP is required to endorse this change. Then the value
  is updated yet again, and this time gateway endorses on Org1MSP but identifies Org2MSP and Org3MSP
  as the required endorsers, and therefore sends the request to available peers of Org2MSP and Org3MSP for endorsement.

* **exampleChaincodeEvents** - In this scenario the client application registers a chaincode
  event listener for the basic chaincode and then invokes the basic chaincode. As in the
  first scenarios, the gateway gathers endorsements from Org1MSP and Org2MSP, and when
  the transaction is submitted and committed, the client application receives the chaincode event.

* **exampleChaincodeEventReplay** - In this scenario the client application invokes the basic chaincode and then
  registers a chaincode event listener to replay the event emitted by the transaction invocation. This is achieved by
  noting the block number in which the transaction commits, and then starting event listening from that block number.

* **exampleErrorHandling** = This scenario demonstrates the different types of errors that can occur during submit of
  a transaction, and how the client application can distinguish between them in order to react appropriately.

## Running the samples

### Go

```sh
make run-samples-go
```

### Node

```sh
make run-samples-node
```

### Java

```sh
make run-samples-java
```

## Additional investigation

If you would like to dig deeper, take a look at the sample client application code.
You can update the samples to call each of the scenarios one at a time, or
try your own scenarios. Then use the following command to look at the
the peer0.org1.example.com log with gateway debugging enabled to see
how the gateway gathers endorsements.

```
docker logs peer0.org1.example.com
```

# Cleanup

When you are finished running the samples, the local docker network can be brought down with the following command:

```sh
make sample-network-clean
```
