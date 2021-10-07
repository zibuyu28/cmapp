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

// Chain chain definition in db
type Chain struct {
	ID         int               `xorm:"int(11) pk 'id'"`
	CreateTime time.Time         `xorm:"datetime create 'create_time'"`
	UpdateTime time.Time         `xorm:"datetime update 'update_time'"`
	DeleteTime time.Time         `xorm:"datetime delete 'delete_time'"`
	Name       string            `xorm:"varchar(256) 'name'"`
	UUID       string            `xorm:"char(64) 'uuid'"`
	Type       string            `xorm:"varchar(256) 'type'"`
	Version    string            `xorm:"varchar(256) 'version'"`
	State      int               `xorm:"int(8) DEFAULT 0 'state'"`
	DriverID   int               `xorm:"int(11) 'driver_id'"`
	Tags       []string          `xorm:"varchar(1024) 'tags'"`
	CustomInfo map[string]string `xorm:"text 'custom_info'"`
}


// InsertChain insert chain to db
func InsertChain(chain *Chain) error {
	_, err := ormEngine.Insert(chain)
	if err != nil {
		return errors.Wrap(err, "chain insert to db")
	}
	return nil
}


// GetChainByID get chain by id
func GetChainByID(id int) (*Chain, error) {
	var drv = &Chain{}
	has, err := ormEngine.Table(&Chain{}).Where("id = ?", id).Get(drv)
	if err != nil {
		return nil, errors.Wrap(err, "query chain from db")
	}
	if !has {
		return nil, errors.Errorf("record not exist which id [%d]", id)
	}
	return drv, nil
}