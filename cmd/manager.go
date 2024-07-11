package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	/**
	websocketUpgrader is used to upgrade incoming HTTP requests into a persistent websocket connection
	*/
	websocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// Manager is used to hold references to all Clients Registered, and Broadcasting etc
type Manager struct {
}

// NewManager is used to initialize all the values inside the manager
func NewManager() *Manager {
	return &Manager{}
}

// serveWS is an HTTP Handler that the has the Manager that allows connections
func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {
	log.Println("New connection from: ", r.RemoteAddr)
	// Begin by upgrading the HTTP request
	conn, err := websocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// We won't do anything yet so close connection again
	errClose := conn.Close()
	if errClose != nil {
		return
	}
}
