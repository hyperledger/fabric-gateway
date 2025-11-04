# Hyperledger Fabric Gateway Client API for Java

The Fabric Gateway client API allows applications to interact with a Hyperledger Fabric blockchain network. It implements the Fabric programming model, providing a simple API to submit transactions to a ledger or query the contents of a ledger with minimal code.

## How to use

Samples showing how to create client applications that connect to and interact with a Hyperledger Fabric network, are available in the [fabric-samples](https://github.com/hyperledger/fabric-samples) repository:

- [asset-transfer-basic](https://github.com/hyperledger/fabric-samples/tree/main/asset-transfer-basic) for examples of transaction submit and evaluate.
- [asset-transfer-events](https://github.com/hyperledger/fabric-samples/tree/main/asset-transfer-events) for examples of chaincode eventing.
- [off_chain_data](https://github.com/hyperledger/fabric-samples/tree/main/off_chain_data) for examples of block eventing.

## API documentation

The Gateway client API documentation for Java is available here:

- https://hyperledger.github.io/fabric-gateway/main/api/java/

## Installation

The Fabric Gateway client API package is published to [Maven Central](https://central.sonatype.com/artifact/org.hyperledger.fabric/fabric-gateway).

### Maven

Add the following dependency to your project's `pom.xml` file:

```xml
<dependencyManagement>
    <dependencies>
        <dependency>
            <groupId>com.google.protobuf</groupId>
            <artifactId>protobuf-bom</artifactId>
            <version>4.33.0</version>
            <type>pom</type>
            <scope>import</scope>
        </dependency>
    </dependencies>
</dependencyManagement>

<dependencies>
    <dependency>
        <groupId>org.hyperledger.fabric</groupId>
        <artifactId>fabric-gateway</artifactId>
        <version>1.10.0</version>
    </dependency>
</dependencies>
```

Note the **pom** import in the `dependencyManagement` section, which ensures that v4 of the Java protocol buffers package is resolved by your project.

A suitable gRPC channel service provider must also be specified (as described in the [gRPC security documentation](https://github.com/grpc/grpc-java/blob/master/SECURITY.md#transport-security-tls)), such as:

```xml
<dependencyManagement>
    <dependencies>
        <dependency>
            <groupId>io.grpc</groupId>
            <artifactId>grpc-bom</artifactId>
            <version>1.76.0</version>
            <type>pom</type>
            <scope>import</scope>
        </dependency>
    </dependencies>
</dependencyManagement>

<dependencies>
    <dependency>
        <groupId>io.grpc</groupId>
        <artifactId>grpc-api</artifactId>
    </dependency>
    <dependency>
        <groupId>io.grpc</groupId>
        <artifactId>grpc-netty-shaded</artifactId>
        <scope>runtime</scope>
    </dependency>
</dependencies>
```

### Gradle

Add the following dependency to your project's `build.gradle` file:

```groovy
implementation 'org.hyperledger.fabric:fabric-gateway:1.10.0'
implementation platform('com.google.protobuf:protobuf-bom:4.33.0')
```

Note the **platform** import, which ensures that v4 of the Java protocol buffers package is resolved by your project.

A suitable gRPC channel service provider must also be specified (as described in the [gRPC security documentation](https://github.com/grpc/grpc-java/blob/master/SECURITY.md#transport-security-tls)), such as:

```groovy
implementation platform('io.grpc:grpc-bom:1.76.0')
compileOnly 'io.grpc:grpc-api'
runtimeOnly 'io.grpc:grpc-netty-shaded'
```

## Compatibility

This API requires Fabric v2.4 (or later) with a Gateway enabled Peer. Additional compatibility information is available in the documentation:

- https://hyperledger.github.io/fabric-gateway/
