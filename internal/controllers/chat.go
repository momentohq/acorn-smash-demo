package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	serviceconfig "github.com/eaddingtonwhite/momento-game-demo/internal/config"

	"github.com/gorilla/websocket"
	"github.com/momentohq/client-sdk-go/momento"
)

var socketUpgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type message struct {
	Value string `json:"Value"`
	User  string `json:"User"`
}

type ChatController struct {
	MomentoClient momento.SimpleCacheClient
}

const (
	chatRoomName = "first-chat-room"
	sysHeartBeat = "SYS:Resubscribe"
)

func (c *ChatController) Connect(w http.ResponseWriter, r *http.Request) {
	conn, err := socketUpgrade.Upgrade(w, r, nil)
	if err != nil {
		writeFatalError(w, "fatal error occurred upgrading client connection to websocket", err)
	}
	// Instantiate subscriber
	sub, err := c.MomentoClient.TopicSubscribe(r.Context(), &momento.TopicSubscribeRequest{
		CacheName: serviceconfig.CacheName,
		TopicName: chatRoomName,
	})
	if err != nil {
		writeFatalError(w, "fatal error occurred subscribing to chat room", err)
	}
	for {
		item, err := sub.Item(r.Context())
		if err != nil {
			writeFatalError(w, "fatal error occurred reading from stream", err)
		}
		switch msg := item.(type) {
		case *momento.TopicValueString:
			// Write message back to browser
			if err = conn.WriteMessage(websocket.TextMessage, []byte(msg.Text)); err != nil {
				writeFatalError(w, "fatal error occurred writing to client websocket", err)
				return
			}
		}

	}
}

func (c *ChatController) SendMessage(w http.ResponseWriter, r *http.Request) {
	var t message
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeFatalError(w, "fatal error occurred decoding msg payload", err)
	}
	_, err := c.MomentoClient.TopicPublish(r.Context(), &momento.TopicPublishRequest{
		CacheName: serviceconfig.CacheName,
		TopicName: chatRoomName,
		Value: &momento.TopicValueString{
			Text: fmt.Sprintf("%s: %s", t.User, t.Value),
		},
	})
	if err != nil {
		writeFatalError(w, "fatal error occurred writing to topic", err)
	}
}
