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
	"bytes"
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/zibuyu28/cmapp/common/log"
	"net/url"
	"time"
)

type Client struct {
	Unique  string
	Target  string
	conn    *websocket.Conn
	Send    chan []byte
	Receive chan []byte
}

func NewClient(unique, target string) *Client {
	return &Client{
		Unique:  unique,
		Target:  target,
		Send:    make(chan []byte, 128),
		Receive: make(chan []byte, 128),
	}
}

func (c *Client) Connect(ctx context.Context) error {
	if len(c.Unique) == 0 || len(c.Target) == 0 {
		return errors.Errorf("client unique [%s] or target [%s] is empty", c.Unique, c.Target)
	}
	u := url.URL{Scheme: "ws", Host: c.Target, Path: "/ws", RawQuery: fmt.Sprintf("unique=%s", c.Unique)}
	log.Debugf(ctx, "connecting to [%s]", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf(ctx, "dial:[%v]", err)
	}
	defer conn.Close()

	c.conn = conn
	return nil
}

// ReadPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() {
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				//log.Debug(c.ctx, "Currently conn is closed")
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.Receive <- message
	}
}

// WritePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
