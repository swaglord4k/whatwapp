package connection

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocket struct {
	socket *websocket.Conn
}

func NewWebSocket(w http.ResponseWriter, r *http.Request) (*WebSocket, error) {
	upgrader := websocket.Upgrader{
		WriteBufferSize: 1024,
		ReadBufferSize:  1024,
	}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	return &WebSocket{
		socket: ws,
	}, err
}

func (s *WebSocket) Close() {
	s.socket.Close()
}

func (s *WebSocket) Write(response interface{}) {
	err := s.socket.WriteJSON(response)
	if err != nil {
		fmt.Println(err)
	}
}

func (s *WebSocket) Listen(channel chan interface{}) {
	go func() {
		for {
			var message interface{}
			err := s.socket.ReadJSON(message)
			if err != nil {
				log.Println(err)
				channel <- "socket closed"
				return
			}
			fmt.Println("received new message")
			fmt.Println(message)
			channel <- message
		}
	}()
}
