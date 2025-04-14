package handlers

import (
	"fmt"
	"http-servidor/utils"
	"math/rand"
	"net"
	"strconv"
)

func Random(conn net.Conn, min string, max string, cantidad string) {
	print("Random handler called\n")

	fmt.Println("Min:", min)
	fmt.Println("Max:", max)
	fmt.Println("Cantidad:", cantidad)

	cantidadI, err := strconv.Atoi(cantidad)
	if err != nil {
		utils.SendResponse(conn, "400 Bad Request", "Cantidad debe ser un numero valido")
		return
	}

	minI, err := strconv.Atoi(min)
	if err != nil {
		utils.SendResponse(conn, "400 Bad Request", "El numero minimo debe ser un numero valido")
		return
	}

	maxI, err := strconv.Atoi(max)
	if err != nil {
		utils.SendResponse(conn, "400 Bad Request", "El numero maximo debe ser un numero valido")
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
	utils.SendResponse(conn, "200 OK", body)

}
