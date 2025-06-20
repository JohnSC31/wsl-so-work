package handlers

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
)

type SendResponseFunc func(conn net.Conn, status string, body string)

func Random(conn net.Conn, min string, max string, cantidad string, sendResponse SendResponseFunc) {

	cantidadI, err := strconv.Atoi(cantidad)
	if err != nil {
		sendResponse(conn, "400 Bad Request", "Cantidad debe ser un numero valido\n")
		return
	}

	minI, err := strconv.Atoi(min)
	if err != nil {
		sendResponse(conn, "400 Bad Request", "El numero minimo debe ser un numero valido\n")
		return
	}

	maxI, err := strconv.Atoi(max)
	if err != nil {
		sendResponse(conn, "400 Bad Request", "El numero maximo debe ser un numero valido\n")
		return
	}

	if(cantidadI <= 0){
		sendResponse(conn, "400 Bad Request", "La cantidad debe ser un numero entero positivo\n")
		return
	}

	if(minI >= maxI){
		sendResponse(conn, "400 Bad Request", "El minimo debe ser menor al maximo\n")
		return
	}

	listaNumRandom := make([]int, cantidadI)
	for i := 0; i < cantidadI; i++ {
		numrandom := rand.Intn(maxI-minI+1) + minI
		listaNumRandom[i] = numrandom
	}

	body := fmt.Sprintf(
		"Se generaron %d numeros aleatorios entre %d y %d:\n\n",
		cantidadI, minI, maxI,
	)
	body += "Indice\tNumero\n"
	body += "------\t------\n"
	for i, num := range listaNumRandom {
		body += fmt.Sprintf("%d\t%d\n", i+1, num)
	}
	sendResponse(conn, "200 OK", body)

}
