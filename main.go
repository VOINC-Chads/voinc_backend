package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
	"sync"
	"voinc-backend/stringgen"
	"voinc-backend/terraform"
	"voinc-backend/websocket"
)

var (
	lobbies  = make(map[string]*websocket.Session)
	mutex    sync.Mutex // for infra.json
	Sessions map[string]string
)

func serveWs(session *websocket.Session, w http.ResponseWriter, r *http.Request, name string) {
	fmt.Println("Endpoint Hit: WebSocket")
	conn, err := websocket.Upgrade(w, r)
	if err != nil {
		fmt.Fprintf(w, "%+v\n", err)
		return
	}
	errInfra := websocket.RegisterInfraSession(*session)
	if errInfra != nil {
		fmt.Fprintf(w, "%+v\n", errInfra)
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
		Session:    session,
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

		lobby := websocket.NewSession("clientID", "secretID", uuid.New().String(), 0)

		go lobby.Start()

		serveWs(lobby, w, r, "test")
	})

	http.Handle("/", rtr)
}

func main() {
	// secretsFile, _ := ioutil.ReadFile("./secrets.json")

	// Let's lock down infra.json eh
	mutex := &sync.Mutex{}
	websocket.InitMutex(mutex)

	// Let's map UUID's to Session objects
	sessions := &map[string]string{}
	websocket.InitSessionMap(sessions)

	terraformInstance := terraform.GetInstance()

	terraformInstance.Apply()

	setupRoutes()

	fmt.Println("Running Go Backend on port 8080")

	http.ListenAndServe(":8080", nil)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}
