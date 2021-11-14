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

type RouterGroup string
type MethodPath string

var GMR = map[RouterGroup]map[MethodPath]func(g *gin.Context){
	RouterGroup(fmt.Sprintf("%s/md", V1.string())): {
		mpf(http.MethodPost, "/exec"): mdExec,
	},
	RouterGroup(fmt.Sprintf("%s/cw", V1.string())): {
		mpf(http.MethodPost, "/exec"): cwExec,
	},
	RouterGroup(fmt.Sprintf("%s/mw", V1.string())): {
		mpf(http.MethodPost, "/exec"): mwExec,
	},
	RouterGroup(fmt.Sprintf("%s/file", V1.string())): {
		mpf(http.MethodPost, "/:file_name"): fileExec,
		mpf(http.MethodGet, "/:file_name"):  fileExec,
	},
}

func mpf(httpMethod, relativePath string) MethodPath {
	// TODO: check path
	return MethodPath(fmt.Sprintf("%s@%s", httpMethod, relativePath))
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
