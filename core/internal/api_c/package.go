/*
 * Copyright © 2021 zibuyu28
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api_c

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	_package "github.com/zibuyu28/cmapp/core/internal/service_c/package"
	"net/http"
	"path/filepath"
)

func packageRegisterExec(g *gin.Context) {
	f, err := g.FormFile("package")
	if err != nil {
		fail(g, errors.Wrap(err, "get file"))
		return
	}
	tempDir := _package.PMi.TempDir()
	join := filepath.Join(tempDir, fmt.Sprintf("%s.tar.gz", uuid.New().String()))
	e := g.SaveUploadedFile(f, join)
	if e != nil {
		fail(g, errors.Wrap(e, "save file"))
		return
	}

	err = _package.PMi.RegisterPackage(join)
	if err != nil {
		fail(g, errors.Wrap(err, "register package"))
		return
	}
	//ok(g, fmt.Sprintf("%s%s", domain, g.Request.RequestURI))
	ok(g, "success")
}

func packageDownloadExec(g *gin.Context) {
	name := g.Param("name")
	version := g.Param("version")
	filen := g.Param("file")
	if len(name) == 0 {
		fail(g, errors.Errorf("name [%s] is nil", name))
		return
	}
	if len(version) == 0 {
		fail(g, errors.Errorf("version [%s] is nil", version))
		return
	}
	if len(filen) == 0 {
		fail(g, errors.Errorf("file [%s] is nil", filen))
		return
	}
	f, err := _package.PMi.DownloadPackage(name, version, filen)
	if err != nil {
		fail(g, errors.Wrap(err, "download file"))
		return
	}
	stat, err := f.Stat()
	if err != nil {
		fail(g,errors.Wrap(err, "stat file"))
		return
	}
	contentLength := stat.Size()
	contentType := "text/plain"
	extraHeaders := map[string]string{
		"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"`, filen),
	}
	g.DataFromReader(http.StatusOK, contentLength, contentType, f, extraHeaders)
}

func packageInfoExec(g *gin.Context) {
	name := g.Param("name")
	version := g.Param("version")
	if len(name) == 0 {
		fail(g, errors.Errorf("name [%s] is nil", name))
		return
	}
	if len(version) == 0 {
		fail(g, errors.Errorf("version [%s] is nil", version))
		return
	}
	info, err := _package.PMi.GetPackageInfo(name, version)
	if err != nil {
		fail(g, errors.Wrap(err, "get package info"))
		return
	}
	ok(g, info)
}