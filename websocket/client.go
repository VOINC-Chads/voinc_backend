package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Client
// - ID:	Client ID,
// - Conn: Reference to websocket connection
// - Session: Reference to session
type Client struct {
	ID         string
	PublicInfo *ClientPublicInfo
	Conn       *websocket.Conn
	Session      *Session
	mu         sync.Mutex
}

type ClientPublicInfo struct {
	Name          string
	Ready         bool
	ScoreIncrease int
	Score         int
	Answer        float32
}

// Message
//- Type: 0 if bytes, 1 if string (I think)
//- Body: String body containing content of message
type Message struct {
	Type int    `json:"type"`
	Body string `json:"body"`
}

// MessageContent
//  - Type: 		String containing type of data (eg. textMsg)
//  - Content:	Content struct
type MessageContent struct {
	MapCode 	string `json:"mapCode"`
	ReduceCode 	string `json:"reduceCode"`
}

type ReadyMessage struct {
	Status bool `json:"status"`
}

// MessageToClient
//- Type: 		The type of response
//- Response: Content of the response (not always there)
type MessageToClient struct {
	Type     string    `json:"type"`
	Response *Response `json:"response"`
}

// Response
// - Session:	Tells the client where it has been moved (likely Session or game)
type Response struct {
	Session string `json:"session,omitempty"`
}

// Content
// - TextMsg: If is of textMsg type,
type Content struct {
	TextMsg string `json:"textMsg,omitempty"`
	Song    string `json:"songSearch,omitempty"`
	SongID  string `json:"songID,omitempty"`
}

// ContentClient
// - Client:
// - Content:
type ContentClient struct {
	Client  *Client
	Content *Content
}

// CreateConversation
//- Participants: A string delimited by | containing a list of participants in the conversation
//- Name:					The name of the conversation
type CreateConversation struct {
	Participants string `json:"participants"`
	Name         string `json:"name"`
}

// GetConversation
// - ConversationID: The hash id of the conversation
// - Offset:					Integer offset of range of messages you're grabbing
// - ClientID:				ID of the client making the get request
type GetConversation struct {
	ConversationID string `json:"conversationID"`
	Offset         int    `json:"offset"`
	ClientID       string `json:"clientID"`
}

func (c *Client) Send(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Conn.WriteJSON(v)
}

// Read function
func (c *Client) Read() {
	defer func() {
		c.Session.Unregister <- c

		err := c.Conn.Close()
		if err != nil {
			return 
		}
	}()

	for {
		messageType, p, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println("Message Received")
		message := Message{Type: messageType, Body: string(p)}
		fmt.Println(message)

		messageContent := &MessageContent{}

		err = json.Unmarshal(p, &messageContent)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Println("MapCode:", messageContent.MapCode)
		fmt.Println("ReduceCode:", messageContent.ReduceCode)
		c.Send(Message{
			Type: 1,
			Body: "Received the message :)",
		})
	}
}
