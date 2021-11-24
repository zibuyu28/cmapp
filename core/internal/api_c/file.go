package api_c

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/zibuyu28/cmapp/common/log"
	"github.com/zibuyu28/cmapp/core/internal/service_c/file"
	"net/http"
)

func fileExec(g *gin.Context) {
	fileName := g.Param("file_name")
	err := fileExecHandler(g, fileName)
	if err != nil {
		fail(g, err)
	}
}

func fileExecHandler(g *gin.Context, fileName string) error {
	if len(fileName) == 0 {
		return errors.New("file name is nil")
	}
	switch g.Request.Method {
	case http.MethodPost:
		f, err := g.FormFile("file")
		if err != nil {
			return errors.Wrap(err, "get file")
		}
		err = file.RFDi.UploadFile(fileName, func(s string) error {
			e := g.SaveUploadedFile(f, s)
			if e != nil {
				return errors.Wrap(e, "save file")
			}
			return nil
		})
		domain := viper.GetString("domain")
		log.Debugf(g.Request.Context(), "get domain [%s]", domain)
		protocol := viper.GetString("protocol")
		log.Debugf(g.Request.Context(), "get protocol [%s]", protocol)
		port := viper.GetInt("http.port")
		log.Debugf(g.Request.Context(), "get http.port [%d]", port)
		ok(g, fmt.Sprintf("%s://%s:%d%s", protocol, domain, port, g.Request.RequestURI))
		return nil
	case http.MethodGet:

		f, err := file.RFDi.DownloadFile(fileName)
		if err != nil {
			return errors.Wrap(err, "download file")
		}
		stat, err := f.Stat()
		if err != nil {
			return errors.Wrap(err, "stat file")
		}
		contentLength := stat.Size()
		contentType := "text/plain"
		extraHeaders := map[string]string{
			"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, fileName),
		}
		g.DataFromReader(http.StatusOK, contentLength, contentType, f, extraHeaders)
		return nil
	default:
		return errors.Errorf("not support method [%s]", g.Request.Method)
	}
}
