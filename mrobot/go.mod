module github.com/zibuyu28/cmapp/mrobot

go 1.15

require (
	github.com/ghodss/yaml v1.0.0
	github.com/go-playground/validator/v10 v10.4.1
	github.com/goinggo/mapstructure v0.0.0-20140717182941-194205d9b4a9
	github.com/google/uuid v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.8.0
	github.com/stretchr/testify v1.7.0
	github.com/zibuyu28/cmapp/common v0.0.0-incompatible
	github.com/zibuyu28/cmapp/core v0.0.0-incompatible
	github.com/zibuyu28/cmapp/plugin v0.0.0-incompatible
	google.golang.org/grpc v1.38.0
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
)

replace github.com/zibuyu28/cmapp/core v0.0.0-incompatible => ../core

replace github.com/zibuyu28/cmapp/common v0.0.0-incompatible => ../common

replace github.com/zibuyu28/cmapp/plugin v0.0.0-incompatible => ../plugin
