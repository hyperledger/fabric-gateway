# Fabric Gateway HSM Samples

The samples in this repo show how to create client applications that invoke transactions with HSM Identities using the
new embedded Gateway in Fabric.

The samples will only run against the latest version of Fabric - v2.4.0-alpha.  The easiest way of setting up a gateway
enabled Fabric network is to use the scenario test framework that is part of this `fabric-gateway` repository using the
following command:

```
export PEER_IMAGE_PULL=hyperledger/fabric-peer:2.4.0-alpha
make sample-network
```

This will create a local docker network comprising five peers across three organisations and a single ordering node.
One of the peers (`peer0.org1.example.com`) has been configured with the gateway enabled.

A simple smart contract (named `basic`) will have been instantiated on all the peers.  The source code for the smart
contract can examined [here](https://github.com/hyperledger/fabric-gateway/blob/main/scenario/fixtures/chaincode/golang/basic/main.go).

A sample client application is provided for each of the supported SDKs.
Note that the SDKs implement the Fabric 'Gateway' programming model which has been in use since
Fabric v1.4, but these are new implementations that target the embedded peer gateway and they share no common code with
existing Fabric SDKs.

In each of the language samples, the client application submits a transaction (`put`) to update the ledger followed by
evaluating a transaction (`get`) to retrieve the value from the ledger (query).
The value that is being updated and retrieved is the current timestamp to demonstrate that the update is working.

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

## Enroll the HSM User

A user, `HSMUser`, who is HSM managed needs to be registered then enrolled for the sample

```bash
make enroll-hsm-user
```

This will register a user `HSMUser` with the CA in Org1 (if not already registered) and then enroll that user which will
generate a certificate on the file system for use by the sample. The private key is stored in SoftHSM

### Node SDK

```
cd <base-path>/fabric-gateway/hsm-samples/node
npm install
npm run build
npm start
```

When you are finished running the samples, the local docker network can be brought down with the following command:

`docker rm -f $(docker ps -aq) && docker network prune --force`
