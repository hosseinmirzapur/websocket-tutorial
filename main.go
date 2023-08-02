package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type Server struct {
	connections map[*websocket.Conn]bool
}

func NewServer() *Server {
	return &Server{
		connections: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) handleWS(ws *websocket.Conn) {
	fmt.Println("New incoming connection from client...", ws.RemoteAddr())

	s.connections[ws] = true

	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)

		if err != nil {

			if err == io.EOF { // This line checks if the client is disconnected
				log.Fatal("client disconnected")
				break
			}
			log.Println("read error: ", err)
		}

		msg := buf[:n]

		s.broadcast(msg)
	}
}

func (s *Server) broadcast(b []byte) {
	for ws := range s.connections {
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(b); err != nil {
				fmt.Println("write error: ", err)
			}
		}(ws)
	}
}

func (s *Server) handleWSOrderBook(ws *websocket.Conn) {
	fmt.Println("New incoming connection from client to orderbook feed:", ws.RemoteAddr())

	for {
		payload := fmt.Sprintf("orderbook data -> %d \n", time.Now().UnixNano())
		ws.Write([]byte(payload))
		time.Sleep(2 * time.Second)
	}
}

func main() {
	server := NewServer()

	/** This is for chatting between clients who connect on the websocket
	If all users close their connection to the websocket, the server will disconnect as well
	*/

	http.Handle("/ws", websocket.Handler(server.handleWS))

	/** This is for orderbook feed, which is a websocket connection which send constant data
	However when client is connected to the websocket, the server will send the data automatically
	*/
	http.Handle("/orderbookfeed", websocket.Handler(server.handleWSOrderBook))

	// The difference between feed and websocket is that a feed has a data to feed to its client already
	// But a websocket connection only feed the data which is send from another client

	fmt.Println("connecting to port 3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Panic(err)
	}
}
