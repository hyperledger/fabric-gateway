module github.com/hyperledger/fabric-gateway

go 1.14

require (
	github.com/fsouza/go-dockerclient v1.6.5 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/hyperledger/fabric v2.1.1+incompatible
	github.com/hyperledger/fabric-amcl v0.0.0-20200424173818-327c9e2cf77a // indirect
	github.com/hyperledger/fabric-protos-go v0.0.0-20191121202242-f5500d5e3e85
	github.com/hyperledger/fabric-sdk-go v1.0.0-beta2
	github.com/pkg/errors v0.9.1
	github.com/sykesm/zap-logfmt v0.0.3 // indirect
	go.uber.org/zap v1.15.0 // indirect
	google.golang.org/grpc v1.30.0
	google.golang.org/protobuf v1.25.0
)

replace github.com/hyperledger/fabric-sdk-go => ../fabric-sdk-go
