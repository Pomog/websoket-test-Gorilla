package main

import (
	"context"
	"fmt"
	"net/http"
)

const PORT = ":8080"

func setupAPI(ctx context.Context) {
	// Create a Manager instance used to handle WebSocket Connections
	manager := NewManager(ctx)

	// Serve the ./frontend directory at Route /
	http.Handle("/", http.FileServer(http.Dir("./frontend")))
	http.HandleFunc("/login", manager.loginHandler)
	http.HandleFunc("/ws", manager.serveWS)

	http.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, len(manager.clients))
	})
}

var (
	allowedOrigin = "https://localhost:" + PORT
)
