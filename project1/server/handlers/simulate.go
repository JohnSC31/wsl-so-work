package handlers

import (
	"net"
	"strconv"
	"time"
)

func Simulate(conn net.Conn, seconds string, nombre string, sendResponse SendResponseFunc) {

	secondsI, err := strconv.Atoi(seconds)
	if err != nil || secondsI <= 0 {
		sendResponse(conn, "400 Bad Request", "Seconds debe ser un numero valido, entero y positivo\n")
		return
	}

	time.Sleep(time.Duration(secondsI) * time.Second)

	body := "Simulacion completada\n\n"
	body += "Nombre de la tarea: " + nombre + "\n"
	body += "Duracion: " + seconds + " segundos\n"
	body += "Hora de finalizacion: " + time.Now().Format(time.RFC1123) + "\n"
	
	sendResponse(conn, "200 OK", body)
}
