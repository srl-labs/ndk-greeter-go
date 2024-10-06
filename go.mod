module github.com/srl-labs/ndk-greeter-go

go 1.21.1

require (
	github.com/RackSec/srslog v0.0.0-20180709174129-a4725f04ec91
	github.com/openconfig/gnmic/pkg/api v0.1.8
	github.com/rs/zerolog v1.33.0
	github.com/srl-labs/bond v0.1.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
)

require (
	cloud.google.com/go/compute/metadata v0.5.2 // indirect
	github.com/AlekSi/pointer v1.2.0 // indirect
	github.com/bufbuild/protocompile v0.14.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/jhump/protoreflect v1.17.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/nokia/srlinux-ndk-go v0.4.0-rc1 // indirect
	github.com/openconfig/gnmi v0.11.0 // indirect
	github.com/openconfig/grpctunnel v0.1.0 // indirect
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/oauth2 v0.23.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240924160255-9d4c2d233b61 // indirect
	google.golang.org/grpc v1.67.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

// this is a fix to a weird deps issue
// cloud.google.com/go/compute/metadata in multiple modules:
// cloud.google.com/go v0.26.0
// cloud.google.com/go/compute/metadata v0.5.1
// not quite sure who is relient on v0.26.0 in the build cache, but this old package had a metadata module,
// and after v0.100 it has been moved to a separate module, so this is a workaround to force the correct version selection
replace cloud.google.com/go => cloud.google.com/go v0.115.1

// same thing for google.golang.org/genproto https://github.com/googleapis/go-genproto/issues/1015
replace google.golang.org/genproto => google.golang.org/genproto v0.0.0-20240903143218-8af14fe29dc1
