module github.com/iris-connect/eps

go 1.16

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.1 // indirect
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/golang/protobuf v1.5.2
	github.com/kiprotect/go-helpers v0.0.0-20211210144244-79ce90e73e79
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.17.0 // indirect
	github.com/prometheus/client_golang v1.12.1
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli v1.22.5
	golang.org/x/crypto v0.0.0-20220131195533-30dcbda58838 // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/sys v0.1.0 // indirect
	google.golang.org/genproto v0.0.0-20220126215142-9970aeb2e350 // indirect
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.27.1
)

// for local testing against a modified go-helpers library
// replace github.com/kiprotect/go-helpers => ../../../geordi/kiprotect/go-helpers
