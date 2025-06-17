package handlers

import (
	"http-servidor/utils"
	"net"
	"strconv"
	"time"
)

func Sleep(conn net.Conn, seconds string) {
	print("Simulate handler called\n")

	secondsI, err := strconv.Atoi(seconds)
	if err != nil || secondsI <= 0 {
		utils.SendResponse(conn, "400 Bad Request", "Seconds debe ser un numero valido, entero y postivo\n")
		return
	}

	time.Sleep(time.Duration(secondsI) * time.Second)

	body := "Sleep realizado durante " + seconds + " segundos\n"
	utils.SendResponse(conn, "200 OK", body)
}
