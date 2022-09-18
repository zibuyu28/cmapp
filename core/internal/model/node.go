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
	"xorm.io/xorm"
)

// Node node definition in db
type Node struct {
	ID         int               `xorm:"int(11) pk autoincr 'id'"`
	CreateTime time.Time         `xorm:"datetime created 'create_time'"`
	UpdateTime time.Time         `xorm:"datetime updated 'update_time'"`
	DeleteTime time.Time         `xorm:"datetime deleted 'delete_time'"`
	Name       string            `xorm:"varchar(256) 'name'"`
	UUID       string            `xorm:"char(64) 'uuid'"`
	Type       string            `xorm:"varchar(256) 'type'"`
	State      int               `xorm:"int(8) DEFAULT 0 'state'"`
	ChainID    int               `xorm:"int(11) 'chain_id'"`
	MachineID  int               `xorm:"int(11) 'machine_id'"`
	Tags       []string          `xorm:"varchar(1024) 'tags'"`
	CustomInfo map[string]string `xorm:"text 'custom_info'"`
}

// InsertNode insert node to db
func InsertNode(node *Node) error {
	_, err := ormEngine.Insert(node)
	if err != nil {
		return errors.Wrap(err, "node insert to db")
	}
	return nil
}

// InsertNodes insert nodes to db
func InsertNodes(ns []*Node) error {
	_, err := ormEngine.Transaction(func(session *xorm.Session) (interface{}, error) {
		for i := range ns {
			_, err := session.Table(&Node{}).InsertOne(ns[i])
			if err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return errors.Wrap(err, "insert nodes")
}

// DeleteNode delete node to db
func DeleteNode(node *Node) error {
	_, err := ormEngine.Delete(node)
	if err != nil {
		return errors.Wrapf(err, "node [%d] delete", node.ID)
	}
	return nil
}
