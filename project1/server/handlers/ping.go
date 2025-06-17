package handlers

import (
	"net"
)

func HandlePing(conn net.Conn) {
	response := "HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/plain\r\n" +
		"Content-Length: 4\r\n" +
		"\r\n" +
		"pong"
	conn.Write([]byte(response))
}
