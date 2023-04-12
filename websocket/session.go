package websocket

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

// For infra.json
var (
	mutex *sync.Mutex
)

func InitMutex(m *sync.Mutex) {
	mutex = m
}

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

type InfraSession struct {
	UUID       string `json:"UUID"`
	NumWorkers int    `json:"NumWorkers"`
}

func RegisterInfraSession(session Session) error {
	// Open infra.json
	mutex.Lock()
	file, err := os.Open("infra/infra.json")
	if err != nil {
		mutex.Unlock()
		return err
	}
	defer file.Close()

	// Read the file contents
	contents, err := ioutil.ReadAll(file)
	if err != nil {
		mutex.Unlock()
		return err
	}

	// Unmarshal the JSON data into an Array of InfraSession structs
	var infraSessions []InfraSession
	err = json.Unmarshal(contents, &infraSessions)
	if err != nil {
		mutex.Unlock()
		return err
	}
	newInfraSession := InfraSession{
		UUID:       session.UUID,
		NumWorkers: session.NumWorkers,
	}
	// Add our new InfraSession
	infraSessions = append(infraSessions, newInfraSession)

	// write the updated list back to the JSON file
	newJson, err := json.MarshalIndent(infraSessions, "", "  ")
	if err != nil {
		mutex.Unlock()
		return err
	}

	errWrite := ioutil.WriteFile("infra/infra.json", newJson, 0644)
	mutex.Unlock()
	if errWrite != nil {
		return errWrite
	}

	return nil
}

func deRegisterInfraSession(session Session) {
	// Open infra.json
	mutex.Lock()
	file, err := os.Open("infra/infra.json")
	if err != nil {
		mutex.Unlock()
		fmt.Println(err)
	}
	defer file.Close()

	// Read the file contents
	contents, err := ioutil.ReadAll(file)
	if err != nil {
		mutex.Unlock()
		fmt.Println(err)
	}

	// Unmarshal the JSON data into an Array of InfraSession structs
	var infraSessions []InfraSession
	err = json.Unmarshal(contents, &infraSessions)
	if err != nil {
		mutex.Unlock()
		fmt.Println(err)
	}
	// Identify the index of the Session to delete
	var indexToDelete int = -1
	for i, p := range infraSessions {
		if p.UUID == session.UUID {
			indexToDelete = i
			break
		}
	}

	// Couldn't find session
	if indexToDelete == -1 {
		mutex.Unlock()
		fmt.Printf("Tried to delete session with UUID %s from infra.json but it wasn't there :(\n", session.UUID)
		return
	}

	// Remove the session
	infraSessions = append(infraSessions[:indexToDelete], infraSessions[indexToDelete+1:]...)

	// write the updated list back to the JSON file
	newJson, err := json.MarshalIndent(infraSessions, "", "  ")
	if err != nil {
		mutex.Unlock()
		fmt.Println(err)
	}

	errWrite := ioutil.WriteFile("infra/infra.json", newJson, 0644)
	mutex.Unlock()
	if errWrite != nil {
		fmt.Println(err)
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
			deRegisterInfraSession(*session)
			if client == session.Host {
				log.Println("Host unregister")
			}
		}

	}
}
