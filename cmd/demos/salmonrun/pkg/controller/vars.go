package controller

import (
	"github.com/gorilla/websocket"
)

var (
	conns = make(map[string]websocket.Conn)
)
