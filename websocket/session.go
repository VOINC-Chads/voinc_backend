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
	Register   chan *Client
	Unregister chan *Client
	Host       *Client
	ClientID   string
	SecretID   string
}


// NewSession
// @return a generated Session
func NewSession(ClientID string, SecretID string) *Session {
	return &Session{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		ClientID:   ClientID,
		SecretID:   SecretID,
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
