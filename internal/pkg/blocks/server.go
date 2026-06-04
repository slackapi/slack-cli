// Copyright 2022-2026 Salesforce, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package blocks

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/slackapi/slack-cli/internal/slackerror"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var netListen = net.Listen

type wsConn interface {
	readMessage(timeout time.Duration) (wsMessage, error)
	writeMessage(msg wsMessage) error
	Close() error
}

type webSocket struct {
	conn *websocket.Conn
}

func (ws *webSocket) readMessage(timeout time.Duration) (wsMessage, error) {
	_ = ws.conn.SetReadDeadline(time.Now().Add(timeout))
	_, data, err := ws.conn.ReadMessage()
	_ = ws.conn.SetReadDeadline(time.Time{})
	if err != nil {
		return wsMessage{}, err
	}
	var msg wsMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return wsMessage{}, fmt.Errorf("invalid message from Block Kit Builder: %w", err)
	}
	return msg, nil
}

func (ws *webSocket) writeMessage(msg wsMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return ws.conn.WriteMessage(websocket.TextMessage, data)
}

func (ws *webSocket) Close() error {
	return ws.conn.Close()
}

type webSocketServer struct {
	server   *http.Server
	Port     int
	connChan <-chan *websocket.Conn
	errChan  <-chan error
}

func newWebSocketServer() (*webSocketServer, error) {
	listener, err := netListen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	port := listener.Addr().(*net.TCPAddr).Port

	connChan := make(chan *websocket.Conn, 1)
	errChan := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, upgradeErr := upgrader.Upgrade(w, r, nil)
		if upgradeErr != nil {
			errChan <- upgradeErr
			return
		}
		connChan <- conn
	})

	server := &http.Server{Handler: mux}
	go func() { _ = server.Serve(listener) }()

	return &webSocketServer{
		server:   server,
		Port:     port,
		connChan: connChan,
		errChan:  errChan,
	}, nil
}

func (s *webSocketServer) Accept(ctx context.Context, timeout time.Duration) (wsConn, error) {
	select {
	case conn := <-s.connChan:
		return &webSocket{conn: conn}, nil
	case err := <-s.errChan:
		return nil, err
	case <-time.After(timeout):
		return nil, slackerror.New(slackerror.ErrBlocksPreview).
			WithMessage("Timed out waiting for Block Kit Builder to connect")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *webSocketServer) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.server.Shutdown(ctx)
}
