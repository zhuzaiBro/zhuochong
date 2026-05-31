package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

type RRoom struct {
	address   string
	callback  string
	connected int32
	wsConnect *websocket.Conn
}

func NewRRoom(address string, callback string) *RRoom {
	return &RRoom{
		address:   address,
		callback:  callback,
		connected: 0,
	}
}

func (r *RRoom) Connect() {
	rr, err := NewRoom(r.address, r.callback)
	if err != nil {
		log.Println(err)
		return
	}
	err = rr.Connect()
	if err != nil {
		log.Println(err)
		return
	}
	r.wsConnect = rr.wsConnect
	atomic.StoreInt32(&r.connected, 1)
	rooms[r.address] = r
}

func (r *RRoom) Close() {
	if atomic.CompareAndSwapInt32(&r.connected, 1, 0) {
		r.wsConnect.Close()
	}
}

var rooms = make(map[string]*RRoom)
var mu sync.Mutex

func handleAPI(w http.ResponseWriter, req *http.Request) {
	address := strings.Trim(req.FormValue("roomId"), `"'`)
	command := req.FormValue("command")
	callback := req.FormValue("callback")
	if address == "" || command == "" {
		http.Error(w, "Address and command are required parameters", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if room, ok := rooms[address]; ok {
		if command == "connect" {
			fmt.Fprintln(w, "OK")
			return
		} else if command == "close" {
			room.Close()
			delete(rooms, address)
		}
	} else if command == "connect" {
		room = NewRRoom(address, callback)
		go room.Connect()
	}

	fmt.Fprintln(w, "OK")
}
