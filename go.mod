module github.com/iris-gateway/eps

go 1.13

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/kiprotect/go-helpers v0.0.0-20210501184624-677c272d4158
	github.com/protocolbuffers/protobuf v3.15.8+incompatible // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli v1.22.5
	google.golang.org/grpc v1.37.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0 // indirect
	google.golang.org/protobuf v1.26.0
)

// for local testing against a modified go-helpers library
// replace github.com/kiprotect/go-helpers => ../../../geordi/kiprotect/go-helpers
