# Hyperledger Fabric Gateway Client SDK for Java


The Fabric Gateway SDK allows applications to interact with a Fabric blockchain network.  It provides a simple API to submit transactions to a ledger or query the contents of a ledger with minimal code.

The Gateway SDK implements the Fabric programming model as described in the [Developing Applications](https://hyperledger-fabric.readthedocs.io/en/latest/developapps/developing_applications.html) chapter of the Fabric documentation.

## How to use

Samples showing how to create a client application that updates and queries the ledger
are available for each of the supported SDK languages here:
https://github.com/hyperledger/fabric-gateway/tree/main/samples

### API documentation

The Java Gateway SDK documentation is available here:
https://hyperledger.github.io/fabric-gateway/main/api/java/

### Installation with Maven

Add the following dependency to your project's `pom.xml` file:

```xml
<repositories>
    <repository>
        <id>nightly-repo</id>
        <url>https://hyperledger-fabric.jfrog.io/artifactory/fabric-maven</url>
    </repository>
</repositories>
<dependencies>
    <dependency>
        <groupId>org.hyperledger.fabric</groupId>
        <artifactId>fabric-gateway</artifactId>
        <version>0.1.0-dev-20210326-14</version>
    </dependency>
</dependencies>
```

### Compatibility

This SDK requires Fabric 2.4 with a Gateway enabled Peer.
