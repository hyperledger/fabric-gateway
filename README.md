# fabric-gateway

Working prototype of Fabric Gateway

To use
- Clone this repo
- Start Fabcar sample in `..../fabric-samples/fabcar` using `./startFabric.sh`
- Follow fabcar instructions for javascript to enroll admin and register user
- Edit the 'registerUser.js' file to register another user for the gateway identity
- `cd ..../fabric-gateway/prototype`
- `export DISCOVERY_AS_LOCALHOST=TRUE`
- `go run gateway.go -h peer0.org1.example.com -p 7051 -m Org1MSP -id ../../fabric-samples/fabcar/javascript/wallet/gateway.id -tlscert ../../fabric-samples/test-network/organizations/peerOrganizations/org1.example.com/tlsca/tlsca.org1.example.com-cert.pem`
    - where the `id` flag points to the wallet id file created for the gateway identity
- In a separate command window:
- `cd ..../fabric-gateway/client/go`
- `go run client2.go  -id ../../../fabric-samples/fabcar/javascript/wallet/appUser.id`

Running the scenario tests
- Install Godog (https://github.com/cucumber/godog#install)
- `cd scenario`
- Run `godog`
