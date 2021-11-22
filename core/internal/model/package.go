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

package model

import (
	"github.com/pkg/errors"
	"time"
)

// Package package definition in db
type Package struct {
	ID                        int       `xorm:"int(11) pk autoincr 'id'"`
	CreateTime                time.Time `xorm:"datetime created 'create_time'"`
	UpdateTime                time.Time `xorm:"datetime updated 'update_time'"`
	DeleteTime                time.Time `xorm:"datetime deleted 'delete_time'"`
	Name                      string    `xorm:"varchar(256) 'name'"`
	Version                   string    `xorm:"varchar(256) 'version'"`
	BinaryName                string    `xorm:"varchar(256) 'binary_name'"`
	BinaryCheckSum            string    `xorm:"varchar(128) 'binary_check_sum'"`
	BinaryPackageHandleShells []string  `xorm:"varchar(2048) 'binary_package_handle_shells'"`
	BinaryStartCommands       []string  `xorm:"varchar(2048) 'binary_start_commands'"`
	ImageFullName             string    `xorm:"varchar(1024) 'image_full_name'"`
	ImageTag                  string    `xorm:"varchar(1024) 'image_tag'"`
	ImageWorkDir              string    `xorm:"varchar(1024) 'image_work_dir'"`
	ImageStartCommands        []string  `xorm:"varchar(2048) 'image_start_commands'"`
}

// InsertPackage insert package to db
func InsertPackage(pkg *Package) error {
	_, err := ormEngine.Insert(pkg)
	if err != nil {
		return errors.Wrap(err, "package insert to db")
	}
	return nil
}

// DeletePackage delete package to db
func DeletePackage(pkg *Package) error {
	_, err := ormEngine.Delete(pkg)
	if err != nil {
		return errors.Wrapf(err, "package [%d] delete", pkg.ID)
	}
	return nil
}

// GetPackageByNameVersion get package by name version
func GetPackageByNameVersion(name, version string) (*Package, error) {
	var drv = &Package{}
	_, err := ormEngine.Table(&Package{}).Where("name = ? and version = ?", name, version).Get(drv)
	if err != nil {
		return nil, errors.Wrap(err, "query package from db")
	}
	//if !has {
	//	return nil, errors.Errorf("record not exist which name [%s], version [%s]", name, version)
	//}
	return drv, nil
}
