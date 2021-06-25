module github.com/zibuyu28/cmapp/core

go 1.13

require (
	github.com/gin-gonic/gin v1.7.2
	github.com/go-sql-driver/mysql v1.5.0
	github.com/google/uuid v1.1.2
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.8.0
	go.uber.org/zap v1.17.0
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 // indirect
	google.golang.org/genproto v0.0.0-20210617175327-b9e0b3197ced // indirect
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
	xorm.io/xorm v1.1.0
	github.com/zibuyu28/cmapp/common v0.0.0-incompatible
)

replace (
	github.com/zibuyu28/cmapp/common v0.0.0-incompatible => ../common
)
