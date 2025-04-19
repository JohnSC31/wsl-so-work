package handlers

import (
	"http-servidor/utils"
	"net"
)

func Help(conn net.Conn) {
	body := `
    Rutas disponibles:
    - /help
    - /timestamp
    - /fibonacci?num=N
    - /reverse?text=abc
    - /toupper?text=abc
    `
	utils.SendResponse(conn, "200 OK", body)
}
