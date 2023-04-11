package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"voinc-backend/stringgen"
	"voinc-backend/websocket"
)

type secretsJSON struct {
	ClientID string `json:clientID`
	SecretID string `json:secretID`
}

var (
	clientID = ""
	secretID = ""
	lobbies  = make(map[string]*websocket.Session)
)

func serveWs(session *websocket.Session, w http.ResponseWriter, r *http.Request, name string) {
	fmt.Println("Endpoint Hit: WebSocket")
	conn, err := websocket.Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
		return
	}

	clientPublicInfo := &websocket.ClientPublicInfo{
		Name:   name,
		Ready:  false,
		Score:  0,
		Answer: 0,
	}

	client := &websocket.Client{
		ID:         stringgen.String(10),
		PublicInfo: clientPublicInfo,
		Conn:       conn,
		Session:      session,
	}

	session.Register <- client
	client.Read()
}

func setupRoutes() {
	rtr := mux.NewRouter()

	rtr.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		// enable CORS to allow browser to make call to API
		enableCors(&w)

		fmt.Fprintf(w, "hello world")
	})

	rtr.HandleFunc("/start-session", func(w http.ResponseWriter, r *http.Request) {
		// enable CORS to allow browser to make call to API
		enableCors(&w)

		lobby := websocket.NewSession(clientID, secretID)

		go lobby.Start()

		serveWs(lobby, w, r, "test")
	})

	http.Handle("/", rtr)
}

func main() {
	secretsFile, _ := ioutil.ReadFile("./secrets.json")

	secrets := secretsJSON{}

	// terraform.Initialize()

	_ = json.Unmarshal(secretsFile, &secrets)
	clientID = secrets.ClientID
	secretID = secrets.SecretID

	setupRoutes()

	fmt.Println("Running Go Backend on port 8080")

	http.ListenAndServe(":8080", nil)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}
