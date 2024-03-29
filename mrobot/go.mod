module github.com/zibuyu28/cmapp/mrobot

go 1.15

require (
	github.com/agiledragon/gomonkey v2.0.1+incompatible
	github.com/bramvdbogaerde/go-scp v1.1.0
	github.com/go-playground/validator/v10 v10.4.1
	github.com/google/uuid v1.2.0
	github.com/intel-go/cpuid v0.0.0-20210602155658-5747e5cec0d9
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.8.0
	github.com/stretchr/testify v1.7.0
	github.com/zibuyu28/cmapp/common v0.0.0-incompatible
	github.com/zibuyu28/cmapp/core v0.0.0-incompatible
	github.com/zibuyu28/cmapp/plugin v0.0.0-incompatible
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	google.golang.org/grpc v1.38.0
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/zibuyu28/cmapp/core v0.0.0-incompatible => ../core

replace github.com/zibuyu28/cmapp/common v0.0.0-incompatible => ../common

replace github.com/zibuyu28/cmapp/plugin v0.0.0-incompatible => ../plugin
