package handlers

import (
    "net"
    "time"
    "http-servidor/utils"
)

func Timestamp(conn net.Conn) {
    
    now := time.Now().Format(time.RFC3339)

    utils.SendResponse(conn, "200 OK", `{"timestamp":"`+now+`"}`+"\n")
}
