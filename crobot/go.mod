module github.com/zibuyu28/cmapp/crobot

go 1.15

require (
	github.com/go-playground/validator/v10 v10.9.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.9.0
	github.com/zibuyu28/cmapp/common v0.0.0-incompatible
	github.com/zibuyu28/cmapp/core v0.0.0-incompatible
	github.com/zibuyu28/cmapp/plugin v0.0.0-incompatible
	google.golang.org/grpc v1.40.0
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/zibuyu28/cmapp/common v0.0.0-incompatible => ../common

replace github.com/zibuyu28/cmapp/plugin v0.0.0-incompatible => ../plugin

replace github.com/zibuyu28/cmapp/core v0.0.0-incompatible => ../core
