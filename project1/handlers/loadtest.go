package handlers

import (
	"fmt"
	"http-servidor/utils"
	"net"
	"strconv"
	"sync"
	"time"
)

func Loadtest(conn net.Conn, tasks string, sleep string) {
	tasksI, err := strconv.Atoi(tasks)
	if err != nil || tasksI < 1 {
		utils.SendResponse(conn, "400 Bad Request", "El parametro 'tasks' debe ser un número valido mayor que 0")
		return
	}

	sleepI, err := strconv.Atoi(sleep)
	if err != nil || sleepI < 0 {
		utils.SendResponse(conn, "400 Bad Request", "El parametro 'sleep' debe ser un numero valido")
		return
	}

	var wg sync.WaitGroup
	horaInicio := time.Now()

	for i := 0; i < tasksI; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Comenzando tarea %d \n", id+1)
			time.Sleep(time.Duration(sleepI) * time.Second)
			fmt.Printf("Tarea %d finalizada\n", id+1)
		}(i)
	}

	// Espera a que terminen todas las tareas
	wg.Wait()

	horaFin := time.Now()
	duration := horaFin.Sub(horaInicio)

	body := fmt.Sprintf(
		"Se ejecutaron %d tareas concurrentes con %d segundos de espera cada una.\nInicio: %s\nFin: %s\nDuracion total: %.2f segundos",
		tasksI,
		sleepI,
		horaInicio.Format(time.RFC1123),
		horaFin.Format(time.RFC1123),
		duration.Seconds(),
	)

	utils.SendResponse(conn, "200 OK", body)
}
