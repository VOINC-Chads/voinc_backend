package websocket

import (
	"fmt"
	"log"
)

/**
* Lobby
*  - Register: Channel for clients to register to the Lobby
*  - Unregister: Channel for clients to unregister from the Lobby
*  - Clients: A map of clients connected to the Lobby
*  - Broadcast: Channel for messaging all clients in Lobby
 */
type Lobby struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Host       *Client
	HostName   string
	Category   string
	ID         string
	ClientID   string
	SecretID   string
}

/*
 * ScoreUpdate
 *  - Name
 *  - Score
 */
type ScoreUpdate struct {
	Name          string
	Score         int
	ScoreIncrease int
	Guess         float32
}

/*
 * NewLobby
 * @return a generated Lobby
 */
func NewLobby(ClientID string, SecretID string) *Lobby {
	return &Lobby{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		ClientID:   ClientID,
		SecretID:   SecretID,
	}
}

func (lobby *Lobby) Start() {

	for {
		select {
		case client := <-lobby.Register:
			if len(lobby.Clients) == 0 {
				lobby.Host = client
			}

			lobby.Clients[client] = true
			client.Send(Message{Type: 0, Body: lobby.ID})
			client.Send(Message{Type: 2, Body: client.ID})

			fmt.Println("Size of Connection Lobby: ", len(lobby.Clients))
			fmt.Println("Lobby ID:", lobby.ID)
		case client := <-lobby.Unregister:
			if client == lobby.Host {
				log.Println("Host unregister")

				for client := range lobby.Clients {
					client.Send(Message{Type: 7, Body: "Session Ended"})
					delete(lobby.Clients, client)
				}
			} else {
				delete(lobby.Clients, client)
				fmt.Println("Size of Connection Lobby: ", len(lobby.Clients))
			}
		}

	}
}
