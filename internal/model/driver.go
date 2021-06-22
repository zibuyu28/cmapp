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

import "time"

// Driver driver definition in db
type Driver struct {
	ID          int       `xorm:"int(11) pk 'id'"`
	CreateTime  time.Time `xorm:"datetime DEFAULT CURRENT_TIMESTAMP 'create_time'"`
	UpdateTime  time.Time `xorm:"datetime DEFAULT CURRENT_TIMESTAMP 'update_time'"`
	DeleteTime  time.Time `xorm:"datetime 'delete_time'"`
	Name        string    `xorm:"varchar(256) 'name'"`
	Type        int8      `xorm:"tinyint(8) 'type'"`
	Version     string    `xorm:"char(32) 'version'"`
	Description string    `xorm:"text 'description'"`
}
