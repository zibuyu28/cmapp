package api_c

import (
	"fmt"
	"github.com/gin-gonic/gin"
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

var GMR = map[string]map[string]func(g *gin.Context){
	fmt.Sprintf("%s/md", V1.string()): {
		mpf(http.MethodPost, "/exec"): mdExec,
	},
	fmt.Sprintf("%s/mc", V1.string()): {},
}

func mpf(httpMethod, relativePath string) string {
	// TODO: check path
	return fmt.Sprintf("%s@%s", httpMethod, relativePath)
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
