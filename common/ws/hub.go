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

package ws

import (
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/md5"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	clients sync.Map
	register chan *ConnIns
	unregister chan *ConnIns
}

var hub = &Hub{
	register:   make(chan *ConnIns),
	unregister: make(chan *ConnIns),
	clients:    sync.Map{},
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients.Store(client.uniqueMd5, client)
		case client := <-h.unregister:
			if _, ok := h.clients.Load(client.uniqueMd5); ok {
				h.clients.Delete(client.uniqueMd5)
				close(client.send)
			}
		}
	}
}

func SendToClient(unique, msg string) error {
	uuid := md5.MD5(unique)

	load, ok := hub.clients.Load(uuid)
	if !ok {
		return errors.Errorf("fail to found ws conn by unique [%s]", unique)
	}
	load.(*ConnIns).send <- []byte(msg)
	return nil
}

func ReceiveFromClient(unique string) (chan []byte, error) {
	uuid := md5.MD5(unique)

	load, ok := hub.clients.Load(uuid)
	if !ok {
		return nil, errors.Errorf("fail to found ws conn by unique [%s]", unique)
	}
	return load.(*ConnIns).receive, nil
}