module github.com/hyperledger/fabric-gateway

go 1.14

require (
	github.com/golang/protobuf v1.4.2
	github.com/hyperledger/fabric-sdk-go v1.0.0-beta2
	github.com/pkg/errors v0.9.1
	google.golang.org/grpc v1.30.0
	google.golang.org/protobuf v1.25.0
)

replace github.com/hyperledger/fabric-sdk-go => ../fabric-sdk-go
