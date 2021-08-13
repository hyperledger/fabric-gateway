# Fabric Gateway HSM Samples

The samples in this repo show how to create client applications that invoke transactions with HSM Identities using the
new embedded Gateway in Fabric.

The samples will only run against Fabric v2.4 and higher.  The easiest way of setting up a gateway
enabled Fabric network is to use the scenario test framework that is part of this `fabric-gateway` repository using the
following command:

```
export PEER_IMAGE_PULL=hyperledger/fabric-peer:2.4.0-beta
make sample-network
```

This will create a local docker network comprising five peers across three organisations and a single ordering node.

Sample client applications are available to demonstrate the features of the Fabric Gateway and associated SDKs using this network.
More details of the samples can be found on the [samples page](https://github.com/hyperledger/fabric-gateway/tree/main/samples).

## C Compilers

In order for the client application to run successfully you must ensure you have C compilers and Python 3 (Note that Python 2 may still work however Python 2 is out of support and could stop working in the future) installed otherwise the node dependency `pkcs11js` will not be built and the application will fail. The failure will have an error such as

```
Error: Cannot find module 'pkcs11js'
```

how to install the required C Compilers and Python will depend on your operating system and version.

## Install SoftHSM

In order to run the application in the absence of a real HSM, a software
emulator of the PKCS#11 interface is required.
For more information please refer to [SoftHSM](https://www.opendnssec.org/softhsm/).

SoftHSM can either be installed using the package manager for your host system:

* Ubuntu: `sudo apt install softhsm2`
* macOS: `brew install softhsm`
* Windows: **unsupported**

Or compiled and installed from source:

1. install openssl 1.0.0+ or botan 1.10.0+
2. download the source code from <https://dist.opendnssec.org/source/softhsm-2.5.0.tar.gz>
3. `tar -xvf softhsm-2.5.0.tar.gz`
4. `cd softhsm-2.5.0`
5. `./configure --disable-gost` (would require additional libraries, turn it off unless you need 'gost' algorithm support for the Russian market)
6. `make`
7. `sudo make install`

## Initialize a token to store keys in SoftHSM

If you have not initialized a token previously (or it has been deleted) then you will need to perform this one time operation

```bash
echo directories.tokendir = /tmp > $HOME/softhsm2.conf
export SOFTHSM2_CONF=$HOME/softhsm2.conf
softhsm2-util --init-token --slot 0 --label "ForFabric" --pin 98765432 --so-pin 1234
```

This will create a SoftHSM configuration file called `softhsm2.conf` and will be stored in your home directory. This is
where the sample expects to find a SoftHSM configuration file

The Security Officer PIN, specified with the `--so-pin` flag, can be used to re-initialize the token,
and the user PIN (see below), specified with the `--pin` flag, is used by applications to access the token for
generating and retrieving keys.

## Install PKCS#11 enabled fabric-ca-client binary
To be able to register and enroll identities using an HSM you need a PKCS#11 enabled version of `fabric-ca-client`
To install this use the following command

```bash
go get -tags 'pkcs11' github.com/hyperledger/fabric-ca/cmd/fabric-ca-client
```
## Enroll the HSM User

A user, `HSMUser`, who is HSM managed needs to be registered then enrolled for the sample

```bash
make enroll-hsm-user
```

This will register a user `HSMUser` with the CA in Org1 (if not already registered) and then enroll that user which will
generate a certificate on the file system for use by the sample. The private key is stored in SoftHSM

### Go SDK

For HSM support you need to ensure you include the `pkcs11` build tag.

```
cd <base-path>/fabric-gateway/hsm-samples/go
go run -tags pkcs11 hsm-sample.go
```

### Node SDK

```
cd <base-path>/fabric-gateway/hsm-samples/node
npm install
npm run build
npm start
```

When you are finished running the samples, the local docker network can be brought down with the following command:

`docker rm -f $(docker ps -aq) && docker network prune --force`