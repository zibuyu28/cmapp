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

// Machine machine definition in db
type Machine struct {
	ID         int               `xorm:"int(11) pk 'id'"`
	CreateTime time.Time         `xorm:"datetime DEFAULT CURRENT_TIMESTAMP 'create_time'"`
	UpdateTime time.Time         `xorm:"datetime DEFAULT CURRENT_TIMESTAMP 'update_time'"`
	DeleteTime time.Time         `xorm:"datetime 'delete_time'"`
	State      int               `xorm:"int(8) DEFAULT 0 'state'"`
	UUID       string            `xorm:"char(64) 'uuid'"`
	DriverID   int               `xorm:"int(11) 'driver_id'"`
	Tags       []string          `xorm:"text 'tags'"`
	CustomInfo map[string]string `xorm:"text 'custom_info'"`
}

// InsertMachine insert machine to db
func InsertMachine(machine *Machine) error {
	_, err := ormEngine.Insert(machine)
	if err != nil {
		return errors.Wrap(err, "machine insert to db")
	}
	return nil
}