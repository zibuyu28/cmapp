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
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"xorm.io/xorm"
)

const DefaultDriverName = "mysql"
const DefaultDataSource = "root:admin123@tcp(127.0.0.1:3306)/cmapp?charset=utf8"

var ormEngine *xorm.Engine

// InitORMEngine init ORM engine
func InitORMEngine() error {
	engine, err := xorm.NewEngine(DefaultDriverName, DefaultDataSource)
	if err != nil {
		return errors.Wrap(err, "new orm engine")
	}
	ormEngine = engine
	err = InitTable()
	if err != nil {
		return errors.Wrap(err, "init table")
	}
	return nil
}

func InitTable() error {
	return ormEngine.Sync2(new(Machine), new(Driver), new(Chain), new(Node), new(Package))
}
