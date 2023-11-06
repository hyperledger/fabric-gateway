module github.com/hyperledger/fabric-gateway/scenario/go

go 1.20

replace github.com/hyperledger/fabric-gateway v0.0.0-unpublished => ../..

require (
	github.com/cucumber/godog v0.13.0
	github.com/cucumber/messages/go/v21 v21.0.1
	github.com/hyperledger/fabric-gateway v0.0.0-unpublished
	github.com/hyperledger/fabric-protos-go-apiv2 v0.2.1
	github.com/spf13/pflag v1.0.5
	google.golang.org/grpc v1.59.0
)

require (
	github.com/cucumber/gherkin/go/v26 v26.2.0 // indirect
	github.com/gofrs/uuid v4.3.1+incompatible // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-memdb v1.3.4 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231030173426-d783a09b4405 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)
