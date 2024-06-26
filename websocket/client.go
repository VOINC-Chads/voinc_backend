package websocket

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
	"voinc-backend/commons"

	"github.com/gorilla/websocket"
	"voinc-backend/client"
)

// Client
// - ID:	Client ID,
// - Conn: Reference to websocket connection
// - Session: Reference to session
type Client struct {
	ID         string
	PublicInfo *ClientPublicInfo
	Conn       *websocket.Conn
	Session    *Session
	ZMQClient  *client.Client
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

type CodeMessage struct {
	ProcessCode  string `json:"processCode"`
	ExecuteCode  string `json:"executeCode"`
	Requirements string `json:"requirements"`
}

type JobMessage struct {
	Jobs []string `json:"jobs"`
}

// MessageContent
//  - Type: 		String containing type of data (eg. textMsg)
//  - Content:	Content struct
type MessageContent struct {
	Type int         `json:"type"`
	Code CodeMessage `json:"code,omitempty"`
	Job  JobMessage  `json:"job,omitempty"`
}

type ReadyMessage struct {
	Status bool `json:"status"`
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

func (c *Client) ReadyPoller() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if c.isReady() {
				return
			}
		}
	}
}

func (c *Client) isReady() bool {
	fmt.Println("isReady()")
	if c.ZMQClient != nil {
		return true
	}

	ip, ok := (*Sessions)[c.Session.UUID]
	// If the key exists
	if ok {
		// Do something
		c.ZMQClient = client.InitializeClient(ip, 8000)
		go c.ZMQClient.Listen(c.Send)
		c.ZMQClient.SendHeartbeat()
		return true
	}
	return false
}

func (c *Client) sendCode(requirements string, processCode string, executeCode string) {
	if !c.isReady() {
		errors.New("backend not ready")
	}

	fmt.Println("Sending code", requirements, processCode, executeCode)
	c.ZMQClient.SendCode(requirements, processCode, executeCode)
}

func (c *Client) sendJobs(jobs []string) {
	if !c.isReady() {
		errors.New("backend not ready")
	}

	fmt.Println("Sending jobs", jobs)
	c.ZMQClient.SendJobs(jobs)
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
		unquotedP, _ := strconv.Unquote(string(p))
		log.Println(unquotedP)
		err = json.Unmarshal([]byte(unquotedP), &messageContent)
		if err != nil {
			log.Println(err)
			c.Send(commons.MessageToClient{
				Status:  "ERROR",
				Content: "Could not process json you sent",
			})
			return
		}
		switch messageContent.Type {
		case 0:
			code := messageContent.Code
			c.Send(commons.MessageToClient{
				Status:  "BRUH",
				Content: "Received the message :)",
			})

			fmt.Println("Sending code", code.Requirements, code.ProcessCode, code.ExecuteCode)
			c.sendCode(code.Requirements, code.ProcessCode, code.ExecuteCode)
		case 1:
			jobs := messageContent.Job
			fmt.Println(jobs.Jobs)
			c.Send(commons.MessageToClient{
				Status:  "BRUH",
				Content: "Received the jobs :)",
			})

			// Make API Call to ipToSendTo to do the job
			c.sendJobs(jobs.Jobs)
		default:
			fmt.Println("Unrecognized type:", messageContent.Type)
			c.Send(commons.MessageToClient{
				Status:  "ERROR",
				Content: "I did not recognize the message content type you provided",
			})
		}

	}
}
