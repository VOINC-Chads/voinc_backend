package websocket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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
	Session    *Session
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

// MessageToClient
//- Status: 		The type of response
//- Content: Content of the response (not always there)
type MessageToClient struct {
	Status  string `json:"status"`
	Content string `json:"content"`
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

func sendCode(ip string, code string) (string, error) {
	data := map[string]string{"code": code}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "http://"+ip+":8000/code", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || !strings.Contains(resp.Status, "200") {
		return "", err
	}
	fmt.Println("CODE BODY: ", body)

	fmt.Println("Send Code Status:", resp.Status)
	return resp.Status, nil
}

func sendJobs(ip string, jobs []string) (string, error) {
	data := map[string]string{"jobs": "[" + strings.Join(jobs, ", ") + "]"}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "http://"+ip+":8000/jobs", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil || !strings.Contains(resp.Status, "200") {
		return "", err
	}

	fmt.Println("Send Job Status:", resp.Status)
	return string(body), nil
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
		// {"type":1,"job":{"jobs":["1","2"]}}
		// "{\"type\":1,\"job\":{\"jobs\":[\"1\",\"2\"]}}"
		// "{\"type\":1,\"job\":{\"jobs\":[\"1\",\"2\"]}}"
		// "{\"type\":1,\"job\":{\"jobs\":[\"1\",\"2\"]}}"
		err = json.Unmarshal([]byte(strings.Replace(string(p)[1:len(string(p))-1], `\"`, `"`, 100)), &messageContent)
		if err != nil {
			log.Println(err)
			c.Send(MessageToClient{
				Status:  "ERROR",
				Content: "Could not process json you sent",
			})
			return
		}
		switch messageContent.Type {
		case 0:
			code := messageContent.Code
			fmt.Println("ProcessCode:", code.ProcessCode)
			fmt.Println("ExecuteCode:", code.ExecuteCode)
			fmt.Println("Requirements", code.Requirements)
			c.Send(MessageToClient{
				Status:  "BRUH",
				Content: "Received the message :)",
			})
			// Make API Call to ipToSendTo with Code!
			ipToSendTo, ok := (*Sessions)[c.Session.UUID]
			if !ok {
				fmt.Printf("UUID: %s is not in Sessions map, maybe the EC2 instance isn't up yet?\n", c.Session.UUID)
				return
			}
			status, err := sendCode(ipToSendTo, code.ExecuteCode)
			if err != nil || !strings.Contains(status, "200") {
				fmt.Println(err)
				// TELL CLIENT WE DID BAD
				c.Send(MessageToClient{
					Status:  "ERROR",
					Content: "Sending your code failed",
				})
			}
		case 1:
			jobs := messageContent.Job
			fmt.Println(jobs.Jobs)
			c.Send(MessageToClient{
				Status:  "BRUH",
				Content: "Received the jobs :)",
			})
			// Make API Call to ipToSendTo to do the job
			ipToSendTo, ok := (*Sessions)[c.Session.UUID]
			if !ok {
				fmt.Printf("UUID: %s is not in Sessions map, maybe the EC2 instance isn't up yet?\n", c.Session.UUID)
				return
			}
			responseBody, err := sendJobs(ipToSendTo, jobs.Jobs)
			if err != nil {
				fmt.Println(err)
				// TELL CLIENT WE DID BAD
				c.Send(MessageToClient{
					Status:  "ERROR",
					Content: "Sending your jobs failed",
				})
			}
			fmt.Println(responseBody)
			c.Send(MessageToClient{
				Status:  "BRUH",
				Content: responseBody,
			})
		default:
			fmt.Println("Unrecognized type:", messageContent.Type)
			c.Send(MessageToClient{
				Status:  "ERROR",
				Content: "I did not recognize the message content type you provided",
			})
		}

	}
}
