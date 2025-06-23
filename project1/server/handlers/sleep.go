package handlers

import (
	"net"
	"strconv"
	"time"
)

func Sleep(conn net.Conn, seconds string, sendResponse SendResponseFunc) {
	print("Simulate handler called\n")

	secondsI, err := strconv.Atoi(seconds)
	if err != nil || secondsI <= 0 {
		sendResponse(conn, "400 Bad Request", "Seconds debe ser un numero valido, entero y postivo\n")
		return
	}

	time.Sleep(time.Duration(secondsI) * time.Second)

	body := "Sleep realizado durante " + seconds + " segundos\n"
	sendResponse(conn, "200 OK", body)
}
