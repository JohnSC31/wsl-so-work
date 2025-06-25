package main
import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"http-servidor/utils"
	"time"
)

// handleCalculatePi: Nueva función para coordinar el cálculo distribuido de Pi
func (d *Dispatcher) handleCalculatePi(conn net.Conn, method string, route string, params map[string]string) {
	totalIterationsStr, ok := params["iterations"]
	if !ok {
		utils.SendResponse(conn, "400 Bad Request", "Parámetro 'iterations' requerido para calcular Pi")
		d.Metrics.mu.Lock()
		d.Metrics.RequestsFailed++
		d.Metrics.mu.Unlock()
		return
	}

	totalIterations, err := strconv.Atoi(totalIterationsStr)
	if err != nil || totalIterations <= 0 {
		utils.SendResponse(conn, "400 Bad Request", "Parámetro 'iterations' debe ser un número entero positivo")
		d.Metrics.mu.Lock()
		d.Metrics.RequestsFailed++
		d.Metrics.mu.Unlock()
		return
	}

	numWorkers := len(d.Workers)
	if numWorkers == 0 {
		utils.SendResponse(conn, "503 Service Unavailable", "No hay workers disponibles para calcular Pi")
		d.Metrics.mu.Lock()
		d.Metrics.RequestsFailed++
		d.Metrics.mu.Unlock()
		return
	}

	iterationsPerWorker := totalIterations / numWorkers
	remainingIterations := totalIterations % numWorkers

	var wg sync.WaitGroup
	resultsChan := make(chan WorkerResult, numWorkers) // Canal para recolectar resultados de workers

	totalPointsGenerated := 0 // Para asegurar que sumamos el total real de iteraciones enviadas
	pointsInCircleTotal := 0  // Acumulador de puntos dentro del círculo

	for i := 0; i < numWorkers; i++ {
		workerIterations := iterationsPerWorker
		if i < remainingIterations {
			workerIterations++ // Distribuir las iteraciones restantes
		}
		if workerIterations == 0 {
			continue // Evitar enviar tareas vacías
		}

		newRequest := Request{
			Method: method,
			Path:   route,
			Params: params,
			Done:   make(chan bool),
		}

		newTask := Task{
			ID:         d.Metrics.TotalRequests,
			Conn:       conn,
			Request:    &newRequest,
			Response:   nil,
			Status:     TaskPending,
			CreatedAt:  time.Now(),
			RetryCount: 0,
			Content:    "",
		}

		worker := seleccionarWorker(d)
		if worker == nil {
			log.Printf("No se pudo seleccionar worker para la tarea de Pi del worker %d, saltando.", i)
			d.Metrics.mu.Lock()
			d.Metrics.RequestsFailed++
			d.Metrics.mu.Unlock()
			continue
		}

		if !d.checkWorkerStatus(worker) {
			log.Printf("Worker %d (%s) marcado como inactivo", worker.ID, worker.URL)
			log.Printf("Cantidad de %s tareas pendientes del worker %d", len(worker.taskQueue), worker.ID)
			d.redistributeTasks(worker)
			utils.SendResponse(conn, "503 Service Unavailable", "Worker no disponible")
			continue
		}

		worker.taskQueue <- &newTask

		worker.mu.Lock()
		worker.CompletedTasks++ // Incrementamos la carga del worker
		worker.activeTasks++ // Incrementamos el contador de tareas activas
		worker.mu.Unlock()

		totalPointsGenerated += workerIterations // Acumular el total real de puntos a generar

		wg.Add(1)
		go func(w *Worker, iterations int, workerID int) {
			defer wg.Done()
			log.Printf("Enviando tarea de Pi (%d iteraciones) a worker %d (%s)", iterations, workerID, w.URL)

			// Usar sendGetToWorker para esta tarea GET
			resultStr, err := d.sendGetToWorker(w, "/calculatepi", map[string]string{"iterations": strconv.Itoa(iterations)})
			if err != nil {
				resultsChan <- WorkerResult{WorkerID: fmt.Sprintf("Worker-%d", workerID), Error: fmt.Errorf("error enviando tarea de Pi a worker %s: %w", w.URL, err)}
				return
			}

			// El worker debe devolver solo el número de puntos dentro del círculo
			pointsInCircle, err := strconv.Atoi(strings.TrimSpace(resultStr))
			if err != nil {
				resultsChan <- WorkerResult{WorkerID: fmt.Sprintf("Worker-%d", workerID), Error: fmt.Errorf("error parseando resultado de Pi de worker %s: %w", w.URL, err)}
				return
			}
			resultsChan <- WorkerResult{WorkerID: fmt.Sprintf("Worker-%d", workerID), Count: pointsInCircle}

		}(worker, workerIterations, worker.ID)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	var errors []error
	for res := range resultsChan {
		if res.Error != nil {
			log.Printf("Error recibido de worker %s para Pi: %v", res.WorkerID, res.Error)
			errors = append(errors, res.Error)
		} else {
			pointsInCircleTotal += res.Count
			log.Printf("Worker %s contribuyó con %d puntos dentro del círculo.", res.WorkerID, res.Count)
		}
	}

	if len(errors) > 0 {
		errMsg := fmt.Sprintf("Errores durante el cálculo de Pi: %v. Conteo parcial de puntos: %d", errors, pointsInCircleTotal)
		utils.SendResponse(conn, "500 Internal Server Error", errMsg)
		d.Metrics.mu.Lock()
		d.Metrics.RequestsFailed++
		d.Metrics.mu.Unlock()
		return
	}

	// Cálculo final de Pi
	piEstimate := 4.0 * float64(pointsInCircleTotal) / float64(totalPointsGenerated)
	log.Printf("Estimación final de Pi: %f (Basado en %d puntos totales, %d dentro del círculo)", piEstimate, totalPointsGenerated, pointsInCircleTotal)
	utils.SendResponse(conn, "200 OK", fmt.Sprintf("Estimación de Pi: %f\n", piEstimate))

	d.Metrics.mu.Lock()
	d.Metrics.RequestsHandled++
	d.Metrics.mu.Unlock()
}