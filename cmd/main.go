package main

import (
	"context"
	"log"
	"net/http"
	"path/filepath"
)

func main() {

	// Create a root ctx and a CancelFunc which can be used to cancel retentionMap goroutine
	rootCtx := context.Background()
	ctx, cancel := context.WithCancel(rootCtx)

	defer cancel()

	setupAPI(ctx)

	certFile := filepath.Join("cmd", "server.crt")
	keyFile := filepath.Join("cmd", "server.key")

	err := http.ListenAndServeTLS(PORT, certFile, keyFile, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
