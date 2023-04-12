package websocket

import (
	"log"
)

// Session
//  - Register: Channel for clients to register to the Lobby
//  - Unregister: Channel for clients to unregister from the Lobby
//  - Clients: A map of clients connected to the Lobby
//  - Broadcast: Channel for messaging all clients in Lobby
type Session struct {
	Register     chan *Client
	Unregister   chan *Client
	Host         *Client
	ClientID     string
	SecretID     string
	ProcessCode  string
	ExecuteCode  string
	Requirements string
	Jobs         []string
	UUID         string
	NumWorkers   int
}

type SessionInfra struct {
	UUID       string `json:"UUID"`
	NumWorkers int    `json:"NumWorkers"`
}

func NewInfraSession(session Session) *SessionInfra {
	return &SessionInfra{
		UUID:       session.UUID,
		NumWorkers: session.NumWorkers,
	}
}

// NewSession
// @return a generated Session
func NewSession(ClientID string, SecretID string, UUID string, NumWorkers int) *Session {
	return &Session{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		ClientID:   ClientID,
		SecretID:   SecretID,
		UUID:       UUID,
		NumWorkers: NumWorkers,
	}
}

func (session *Session) Start() {

	for {
		select {
		case client := <-session.Register:
			client.Send(Message{Type: 1, Body: "Connected to session"})

		case client := <-session.Unregister:
			log.Println("Session ended :'(")
			if client == session.Host {
				log.Println("Host unregister")
			}
		}

	}
}
