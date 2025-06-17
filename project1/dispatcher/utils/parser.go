package utils

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

// mutex para los archivos
var FilesMutex = &sync.Mutex{}


func ParseRequestLine(request string) (method, path string) {
	lines := strings.Split(request, "\r\n")
	if len(lines) > 0 {
		parts := strings.Split(lines[0], " ")
		if len(parts) >= 2 {
			return parts[0], parts[1]
		}
	}
	return "", ""
}

func ParseRoute(path string) (string, map[string]string) {
	parts := strings.Split(path, "?")
	route := parts[0]
	params := make(map[string]string)

	if len(parts) > 1 {
		pairs := strings.Split(parts[1], "&")
		for _, pair := range pairs {
			kv := strings.Split(pair, "=")
			if len(kv) == 2 {
				params[kv[0]] = kv[1]
			}
		}
	}
	fmt.Println("Route:", route)
	fmt.Println("Params:", params)
	return route, params
}

func SendResponse(conn net.Conn, status, body string) {
	response := fmt.Sprintf("HTTP/1.0 %s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", status, len(body), body)
	print(response)
	conn.Write([]byte(response))
}

func SendJSON(conn net.Conn, status string, body []byte) {
    header := fmt.Sprintf("HTTP/1.0 %s\r\nContent-Type: application/json\r\nContent-Length: %d\r\n\r\n", status, len(body))
    conn.Write([]byte(header))
    conn.Write(body)
}
