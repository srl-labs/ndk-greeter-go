module github.com/srl-labs/ndk-greeter-go

go 1.21.1

toolchain go1.21.6

require (
	github.com/RackSec/srslog v0.0.0-20180709174129-a4725f04ec91
	github.com/nokia/srlinux-ndk-go v0.4.0-rc1
	github.com/openconfig/gnmic/pkg/api v0.1.5
	github.com/rs/zerolog v1.31.0
	google.golang.org/grpc v1.61.0
	google.golang.org/protobuf v1.33.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	cloud.google.com/go/compute v1.23.3 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/AlekSi/pointer v1.2.0 // indirect
	github.com/bufbuild/protocompile v0.6.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/jhump/protoreflect v1.15.3 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/openconfig/gnmi v0.10.0 // indirect
	github.com/openconfig/gnmic/pkg/types v0.1.2 // indirect
	github.com/openconfig/gnmic/pkg/utils v0.1.0 // indirect
	github.com/openconfig/grpctunnel v0.1.0 // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/oauth2 v0.17.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240123012728-ef4313101c80 // indirect
)

replace github.com/sr-linux/bond => ../bond
