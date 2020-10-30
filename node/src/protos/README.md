Fabric and Gateway protobuf definitions

The two files `protos.js` and `protos.d.ts` are generated from the fabric and gateway `.proto` definitions using the following commands: 

- `pbjs -t static-module $(find $GOPATH/src/github.com/hyperledger/fabric-protos -name *.proto -type f) $GOPATH/src/github.com/hyperledger/fabric-gateway/protos/gateway.proto  -o bundle.js --keep-case`
- `pbts -o protos.d.ts bundle.js`

They need to be manually regenerated if there are (relevant) changes to the proto definitions.

The `pbjs` and `pbts` command line tools are part of the https://www.npmjs.com/package/protobufjs package.
