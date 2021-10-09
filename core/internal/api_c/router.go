package api_c

import (
	"github.com/gin-gonic/gin"
	"github.com/zibuyu28/cmapp/core/internal/server"
	"net/http"
)

type routerVersion string

const (
	V1 routerVersion = "v1"
	V2 routerVersion = "v2"
)

func (r routerVersion) string() string {
	return string(r)
}

var machineApiV1 *gin.RouterGroup

type MDReq struct {
	AppUUID string      `json:"app_uuid" validate:"required"`
	Fnc     string      `json:"fnc" validate:"required"`
	Param   interface{} `json:"param" validate:"required"`
}

// MachineDriverV1Ctl machine driver v1 controller
func MachineDriverV1Ctl() {
	machineApiV1 = server.Group(V1.string(), "md")
	machineApiV1.Handle(http.MethodPost, "/exec", func(g *gin.Context) {
		var p = MDReq{}
		err := g.BindJSON(&p)
		if err != nil {
			fail(g, err)
			return
		}
		r, err := mdExecHandler(&p)
		if err != nil {
			fail(g, err)
			return
		}
		ok(g, r)
	})
}

// Resp response
type Resp struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func fail(g *gin.Context, e error) {
	g.JSON(http.StatusOK, Resp{Code: http.StatusBadRequest, Data: nil, Message: e.Error()})
}

func ok(g *gin.Context, d interface{}) {
	g.JSON(http.StatusOK, Resp{Code: http.StatusOK, Data: d})
}
