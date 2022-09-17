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
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"xorm.io/xorm"
)

const DefaultDriverName = "mysql"

var ormEngine *xorm.Engine

// InitORMEngine init ORM engine
func InitORMEngine() error {
	host := viper.GetString("mysql.host")
	port := viper.GetInt("mysql.port")
	user := viper.GetString("mysql.username")
	password := viper.GetString("mysql.password")
	engine, err := xorm.NewEngine(DefaultDriverName, fmt.Sprintf("%s:%s@tcp(%s:%d)/cmapp?charset=utf8", user, password, host, port))
	if err != nil {
		return errors.Wrap(err, "new orm engine")
	}
	ormEngine = engine
	err = InitTable()
	if err != nil {
		return errors.Wrap(err, "init table, check database 'cmapp' first")
	}
	return nil
}

func InitTable() error {
	return ormEngine.Sync2(new(Machine), new(Driver), new(Chain), new(Node), new(Package))
}
