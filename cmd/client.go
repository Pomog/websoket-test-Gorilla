package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

// ClientList is a map used to help manage a map of clients
type ClientList map[*Client]bool

// Client is a websocket client, basically a frontend visitor
type Client struct {
	// the websocket connection
	connection *websocket.Conn

	// manager is the manager used to manage the client
	manager *Manager

	// egress is used to avoid concurrent writes on the WebSocket
	egress chan Event

	// chatroom is used to know what room user is in
	chatroom string
}

var (
	// pongWait is how long we will await a pong response from client
	pongWait = 10 * time.Second

	// pingInterval has to be less than pongWait, We cant multiply by 0.9 to get 90% of time
	// Because that can make decimals, so instead *9 / 10 to get 90%
	pingInterval = (pongWait * 9) / 10

	// maxMessageLength maximum length of the chat message
	maxMessageLength int64 = 512
)

// NewClient is used to initialize a new Client with all required values initialized
func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan Event),
	}
}

// readMessages will start the client to read messages and handle them
// appropriately.
// This is supposed to ran as a goroutine
func (c *Client) readMessages() {
	defer func() {
		// Graceful Close the Connection once this
		// function is done
		c.manager.removeClient(c)
	}()

	// Set Max Size of Messages in Bytes
	c.connection.SetReadLimit(maxMessageLength)

	// Configure Wait time for Pong response, use Current time + pongWait
	// This has to be done here to set the first initial timer.
	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Println(err)
		return
	}
	// Configure how to handle Pong responses
	c.connection.SetPongHandler(c.pongHandler)

	// Loop Forever
	for {
		// ReadMessage is used to read the next message in queue
		// in the connection
		_, payload, err := c.connection.ReadMessage()

		if err != nil {
			// If Connection is closed, we will Receive an error here
			// We only want to log Strange errors, but not simple Disconnection
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break // Break the loop to close conn & Cleanup
		}

		log.Println("Payload: ", string(payload))

		// Marshal incoming data into an Event struct
		var request Event
		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("error (json.Unmarshal(payload, &request)) marshalling message: %v", err)
			break // Breaking the connection here might be harsh xD
		}
		// Route the Event
		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println("Error handling Message: ", err)
		}
	}
}

// pongHandler is used to handle PongMessages for the Client
func (c *Client) pongHandler(pongMsg string) error {
	log.Println("pong")
	// Current time + Pong Wait time
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}

// writeMessages continuously reads from the client's egress channel
// and writes messages to the WebSocket connection.
func (c *Client) writeMessages() {
	// Create a ticker that triggers a ping at given interval
	ticker := time.NewTicker(pingInterval)

	defer func() {
		// Graceful close if this triggers a closing
		c.manager.removeClient(c)
	}()

	for {
		select {
		case message, ok := <-c.egress:
			// Ok will be false in case the egress channel is closed
			if !ok {
				// Manager has closed this connection channel, so communicate that to frontend
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					// Log that the connection is closed and the reason
					log.Println("connection closed: ", err)
				}
				// Return to close the goroutine
				return
			}

			// Marshal the data before sending it
			data, err := json.Marshal(message)
			if err != nil {
				log.Println("Error in the func (c *Client) writeMessages()")
				log.Println(err)
				return // closes the connection, should we really
			}

			// Write a Regular text message to the connection
			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println(err)
			}
			log.Println("sent message")

		case <-ticker.C:
			log.Println("ping")
			// Send the Ping
			if err := c.connection.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Println("write msg: ", err)
				return // return to break this goroutine triggering cleanup
			}
		}

	}
}
