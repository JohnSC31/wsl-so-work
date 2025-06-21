package handlers

import (
    "net"
    "time"
)

func Timestamp(conn net.Conn, sendResponse SendResponseFunc) {
    
    now := time.Now().Format(time.RFC3339)

    sendResponse(conn, "200 OK", `{"timestamp":"`+now+`"}`+"\n")
}
