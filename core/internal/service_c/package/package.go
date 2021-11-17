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

package _package

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/zibuyu28/cmapp/common/file"
	"github.com/zibuyu28/cmapp/common/md5"
	"github.com/zibuyu28/cmapp/core/internal/model"
	"io/ioutil"
	"os"
	"path/filepath"
)

// PM  package manager
type PM struct {
	baseDir string
}

const packageDir = "./package"

var PMi = PM{baseDir: packageDir}

// RegisterPackage register package
// 注册package信息，包括，二进制相关包，二进制包的hash，二镜像，镜像的hash
func (p *PM) RegisterPackage(tar string) error {
	dir := filepath.Dir(tar)
	defer os.RemoveAll(dir)
	// 必须是 tar 包, 解压
	split, f := filepath.Split(dir)
	err := file.UntargzWithName(tar, split, f)
	if err != nil {
		return errors.Wrap(err, "un tar file")
	}
	infoJsonPath := filepath.Join(dir, "info.json")
	infoJson, err := ioutil.ReadFile(infoJsonPath)
	if err != nil {
		return errors.Wrap(err, "read info.json")
	}
	// 读取 info.json 解析为Packages
	var pkgs Packages
	err = json.Unmarshal(infoJson, &pkgs)
	if err != nil {
		return errors.Wrapf(err, "unmarshal info json [%s]", string(infoJson))
	}
	err = validator.New().Struct(pkgs)
	if err != nil {
		return errors.Wrap(err, "check param")
	}
	// 检查每个 package
	var mps []model.Package
	for _, pkg := range pkgs.Packages {
		fileName := filepath.Join(dir, pkg.Binary.FileName)
		_, err = os.Stat(fileName)
		if err != nil {
			return errors.Wrapf(err, "stat binary file [%s]", fileName)
		}
		fileMD5, err := md5.FileMD5(fileName)
		if err != nil {
			return errors.Wrapf(err, "get file [%s] md5", pkg.Binary.FileName)
		}
		if fileMD5 != pkg.Binary.CheckSum {
			return errors.Errorf("file md5 [%s] not equal to checksum [%s]", fileMD5, pkg.Binary.CheckSum)
		}
		mp := model.Package{
			Name:                      pkg.Name,
			Version:                   pkg.Version,
			BinaryName:                pkg.Binary.FileName,
			BinaryCheckSum:            pkg.Binary.CheckSum,
			BinaryPackageHandleShells: pkg.Binary.PackageHandleShells,
			BinaryStartCommands:       pkg.Binary.StartCommands,
		}
		if pkg.Image != nil {
			fileName = filepath.Join(dir, pkg.Image.FileName)
			_, err = os.Stat(fileName)
			if err != nil {
				return errors.Wrapf(err, "stat image file [%s]", fileName)
			}
			// TODO: load image, than upload to repository, get full name

			mp.ImageFullName = fmt.Sprintf("myrepository/platform/%s", pkg.Name)
			mp.ImageTag = pkg.Version
			mp.ImageWorkDir = pkg.Image.WorkDir
			mp.ImageStartCommands = pkg.Image.StartCommands
		}
		mps = append(mps, mp)
	}
	// 全部检查通过，入库
	for _, mp := range mps {
		err = model.InsertPackage(&mp)
		if err != nil {
			return errors.Wrap(err, "store package info")
		}
	}
	for _, pk := range pkgs.Packages {
		pkp := filepath.Join(p.baseDir, fmt.Sprintf("%s-%s", pk.Name, pk.Version))
		_ = os.MkdirAll(pkp, os.ModePerm)
		_ = os.Rename(filepath.Join(dir, pk.Binary.FileName), filepath.Join(pkp, pk.Binary.FileName))
		_ = os.Rename(filepath.Join(dir, pk.Image.FileName), filepath.Join(pkp, pk.Image.FileName))
		marshal, _ := json.Marshal(pk)
		_ = ioutil.WriteFile(filepath.Join(pkp,"info.json"), marshal, os.ModePerm)
	}
	return nil
}

type Packages struct {
	Packages []Package `json:"packages"`
}
type Package struct {
	Name    string  `json:"name" validate:"required"`
	Version string  `json:"version" validate:"required"`
	Binary  *Binary `json:"binary" validate:"required"`
	Image   *Image  `json:"image"`
}
type Binary struct {
	FileName            string   `json:"file_name" validate:"required"`
	CheckSum            string   `json:"check_sum" validate:"required"`
	PackageHandleShells []string `json:"package_handle_shells"`
	StartCommands       []string `json:"start_commands" validate:"required"`
}

type Image struct {
	FileName      string   `json:"file_name" validate:"required"`
	WorkDir       string   `json:"work_dir" validate:"required"`
	StartCommands []string `json:"start_commands" validate:"required"`
}

func (p *PM) GetPackageInfo(name, version string) (*PackageInfo, error) {
	// 查询数据库
	pkg, err := model.GetPackageByNameVersion(name, version)
	if err != nil {
		return nil, errors.Wrap(err, "get package info")
	}
	domain := viper.GetString("domain")
	return &PackageInfo{
		Name:    pkg.Name,
		Version: pkg.Version,
		Image: struct {
			ImageName     string   `json:"image_name" validate:"required"`
			Tag           string   `json:"tag" validate:"required"`
			WorkDir       string   `json:"work_dir" validate:"required"`
			StartCommands []string `json:"start_command" validate:"required"`
		}{
			ImageName:     pkg.ImageFullName,
			Tag:           pkg.ImageTag,
			WorkDir:       pkg.ImageWorkDir,
			StartCommands: pkg.ImageStartCommands,
		},
		Binary: struct {
			Download            string   `json:"download" validate:"required"`
			CheckSum            string   `json:"check_sum" validate:"required"`
			PackageHandleShells []string `json:"package_handle_shells"`
			StartCommands       []string `json:"start_command" validate:"required"`
		}{
			Download:            fmt.Sprintf("%s/api/v1/package/%s/%s/%s", domain, pkg.Name, pkg.Version, pkg.BinaryName),
			CheckSum:            pkg.BinaryCheckSum,
			PackageHandleShells: pkg.BinaryPackageHandleShells,
			StartCommands:       pkg.BinaryStartCommands,
		},
	}, nil

}

type PackageInfo struct {
	Name    string `json:"name" validate:"required"`
	Version string `json:"version" validate:"required"`
	Image   struct {
		ImageName     string   `json:"image_name" validate:"required"`
		Tag           string   `json:"tag" validate:"required"`
		WorkDir       string   `json:"work_dir" validate:"required"`
		StartCommands []string `json:"start_command" validate:"required"`
	}
	Binary struct {
		Download            string   `json:"download" validate:"required"`
		CheckSum            string   `json:"check_sum" validate:"required"`
		PackageHandleShells []string `json:"package_handle_shells"`
		StartCommands       []string `json:"start_command" validate:"required"`
	}
}

func (p *PM) DownloadPackage(pkgName, pkgVersion, file string) (*os.File, error) {
	filePath := filepath.Join(p.baseDir, fmt.Sprintf("%s-%s", pkgName, pkgVersion), file)
	open, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "op file [%s]", filePath)
	}
	return open, nil
}

func (p *PM) TempDir() string {
	join := filepath.Join(p.baseDir, uuid.New().String())
	_ = os.MkdirAll(join, os.ModePerm)
	return join
}
