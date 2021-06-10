package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type room struct {
	forward chan []byte
	join chan *client
	leave chan *client
	clients map[*client]bool
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// 参加
			r.clients[client] = true
		case client := <-r.leave:
			// 退室
			delete(r.clients, client)
			close(client.send)
		case msg := <-r.forward:
			for client := range r.clients {
				select {
				case client.send <- msg:
				default:
					// 送信に失敗
					delete(r.clients, client)
					close(client.send)
				}
			}
		}
	}
}

const (
	socketBufferSize = 1024
	messageBuffersize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize:
	socketBufferSize, WriteBufferSize: socketBufferSize}
func (r *room) ServerHttp(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServerHTTP:", err)
		return
	}
	client := &client{
		socket: socket,
		send: make(chan []byte, messageBuffersize),
		room: r,
	}
	r.join <- client
	defer func() { r.leave <- client } ()
	go client.write()
	client.read()
}
