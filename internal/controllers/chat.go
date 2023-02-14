package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	serviceconfig "github.com/eaddingtonwhite/momento-game-demo/internal/config"

	"github.com/gorilla/websocket"
	"github.com/momentohq/client-sdk-go/incubating"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	MomentoClient incubating.ScsClient
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
	sub, err := c.MomentoClient.SubscribeTopic(r.Context(), &incubating.TopicSubscribeRequest{
		CacheName: serviceconfig.CacheName,
		TopicName: chatRoomName,
	})
	if err != nil {
		writeFatalError(w, "fatal error occurred subscribing to chat room", err)
	}
	for {
		item, err := sub.Item()
		if err != nil {
			if s, ok := status.FromError(err); ok {
				// Handle Server disconnect happy path after inactivity
				// TODO move down into SDK so much thinner here
				if s.Code() == codes.Internal &&
					s.Message() == "stream terminated by RST_STREAM with error code: NO_ERROR" {

					// Re-establish connection
					fmt.Println("stream timed out re-establishing now")
					sub, err = c.MomentoClient.SubscribeTopic(r.Context(), &incubating.TopicSubscribeRequest{
						CacheName: serviceconfig.CacheName,
						TopicName: chatRoomName,
					})
					if err != nil {
						writeFatalError(w, "fatal error occurred trying to re-establish stream", err)
						return
					}
					err := c.MomentoClient.PublishTopic(r.Context(), &incubating.TopicPublishRequest{
						CacheName: serviceconfig.CacheName,
						TopicName: chatRoomName,
						Value:     &incubating.TopicValueString{Text: sysHeartBeat},
					})
					if err != nil {
						writeFatalError(w, "fatal error occurred trying to publish sys message to re-established stream", err)
						return
					}
				} else {
					writeFatalError(w, "fatal error occurred trying to read from stream err", err)
					return
				}
			} else {
				if err != nil {
					writeFatalError(w, "fatal error occurred reading from stream", err)
				}
			}
		}
		switch msg := item.(type) {
		case *incubating.TopicValueString:
			// Write message back to browser
			if msg.Text != sysHeartBeat {
				if err = conn.WriteMessage(websocket.TextMessage, []byte(msg.Text)); err != nil {
					writeFatalError(w, "fatal error occurred writing to client websocket", err)
					return
				}
			}
		}

	}
}

func (c *ChatController) SendMessage(w http.ResponseWriter, r *http.Request) {
	var t message
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeFatalError(w, "fatal error occurred decoding msg payload", err)
	}
	err := c.MomentoClient.PublishTopic(r.Context(), &incubating.TopicPublishRequest{
		CacheName: serviceconfig.CacheName,
		TopicName: chatRoomName,
		Value:     &incubating.TopicValueString{Text: fmt.Sprintf("%s: %s", t.User, t.Value)},
	})
	if err != nil {
		writeFatalError(w, "fatal error occurred writing to topic", err)
	}
}
