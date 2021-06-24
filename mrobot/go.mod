module github.com/zibuyu28/cmapp/mrobot

go 1.16

require (
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.8.0
	github.com/zibuyu28/cmapp/core v0.0.0-incompatible
	google.golang.org/grpc v1.38.0
)

replace github.com/zibuyu28/cmapp/core v0.0.0-incompatible => ../core
