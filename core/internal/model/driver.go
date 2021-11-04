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

// Driver driver definition in db
type Driver struct {
	ID          int       `xorm:"int(11) pk autoincr 'id'"`
	CreateTime  time.Time `xorm:"datetime created 'create_time'"`
	UpdateTime  time.Time `xorm:"datetime updated 'update_time'"`
	DeleteTime  time.Time `xorm:"datetime deleted 'delete_time'"`
	Name        string    `xorm:"varchar(256) 'name'"`
	Type        int8      `xorm:"tinyint(8) 'type'"`
	Version     string    `xorm:"char(32) 'version'"`
	Description string    `xorm:"text 'description'"`
}

// InsertDriver insert driver to db
func InsertDriver(driver *Driver) error {
	_, err := ormEngine.Insert(driver)
	if err != nil {
		return errors.Wrap(err, "driver insert to db")
	}
	return nil
}

// GetDriverByID get driver by id
func GetDriverByID(id int) (*Driver, error) {
	var drv = &Driver{}
	has, err := ormEngine.Table(&Driver{}).Where("id = ?", id).Get(drv)
	if err != nil {
		return nil, errors.Wrap(err, "query driver from db")
	}
	if !has {
		return nil, errors.Errorf("record not exist which id eq [%d]", id)
	}
	return drv, nil
}
