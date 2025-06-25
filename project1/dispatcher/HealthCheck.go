package main
import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
    "http-servidor/utils"
)

// revisa el estado del worker
func (d *Dispatcher) checkWorkerStatus(w *Worker) bool {
    conn, err := net.DialTimeout("tcp", w.URL, 5*time.Second)
    if err != nil {
        w.mu.Lock()
        w.Status = false
        w.mu.Unlock()
		d.redistributeTasks(w)
        return false
    }
    defer conn.Close()

    // Enviar solicitud HTTP de verificación
    _, err = fmt.Fprintf(conn, "GET /ping HTTP/1.1\r\nHost: %s\r\n\r\n", w.URL)
    if err != nil {
        w.mu.Lock()
        w.Status = false
        w.mu.Unlock()
		d.redistributeTasks(w)
        return false
    }

    // Verificar respuesta
    scanner := bufio.NewScanner(conn)
    if scanner.Scan() && strings.Contains(scanner.Text(), "200 OK") {
        w.mu.Lock()
        w.Status = true
        w.mu.Unlock()
        return true
    }

    w.mu.Lock()
    w.Status = false
    w.mu.Unlock()
    return false
}

// hace el health check
func (d *Dispatcher) HealthCheck() {
    for _, worker := range d.Workers {
        
        d.checkWorkerStatus(worker)
        
		log.Printf("Worker %d (%s) estado: %t", worker.ID, worker.URL, worker.Status)
    }
	
}


// redistribuye las tareas pendientes de un worker apagado
/*func (d *Dispatcher) redistributeTasks(failedWorker *Worker) {
    failedWorker.mu.Lock()
    defer failedWorker.mu.Unlock()

    var pendingTasks []*Task
    for {
        select {
        case task := <-failedWorker.taskQueue:
            pendingTasks = append(pendingTasks, task)
        default:
            // Redistribuir tareas
            for _, task := range pendingTasks {
                newWorker := seleccionarWorker(d)
                if newWorker != nil {
                    newWorker.taskQueue <- task
					log.Printf("Redistribuyendo tarea %s del worker %d al worker %d", task.ID, failedWorker.ID, newWorker.ID)

                }
            }
            return
        }
    }
}*/

// Redistribuye las tareas pendientes de un worker apagado
func (d *Dispatcher) redistributeTasks(failedWorker *Worker) {
    failedWorker.mu.Lock()
    defer failedWorker.mu.Unlock()

    var pendingTasks []*Task
    for {
        select {
        case task := <-failedWorker.taskQueue:
            pendingTasks = append(pendingTasks, task)
        default:
            // Redistribuir tareas
            for _, task := range pendingTasks {
                newWorker := seleccionarWorker(d)
                if newWorker != nil {
                    newWorker.taskQueue <- task
                    log.Printf("Redistribuyendo tarea %s del worker %d al worker %d", task.ID, failedWorker.ID, newWorker.ID)
                } else {
                    log.Printf("Redistribución de tarea %s fallida, no hay workers disponibles", task.ID)
                }
            }
            return
        }
    }
}


// selecciona el worker que se va a usar para procesar la tarea
func seleccionarWorker(d *Dispatcher) *Worker {
	// Estrategia de round robin
	if EstrategiaRed == 1 {
		log.Printf("workers disponibles: %d", len(d.Workers))
		
		for i := 0; i < len(d.Workers); i++ {
			// Buscar el siguiente worker disponible después del último usado
			log.Printf("estado del worker %d: %v", i, d.Workers[i].Status)
			idx := (d.lastWorkerIndex + i + 1) % len(d.Workers)
			worker := d.Workers[idx]

			if worker.Status {
				d.lastWorkerIndex = idx

				return worker
			}
		}

	}
	// Estrategia de least loaded
	if EstrategiaRed == 2 {
		var minLoad = -1
		var selectedWorker *Worker = nil

		for _, worker := range d.Workers {
			if worker.Status {
				// Si es el primer worker disponible o tiene menos carga
				if minLoad == -1 || worker.CompletedTasks < minLoad {
					minLoad = worker.CompletedTasks
					selectedWorker = worker
				}
			}
		}

		if selectedWorker != nil {
			selectedWorker.CompletedTasks++ // Incrementamos su carga
			return selectedWorker
		}

	}
	return nil // Aquí se implementaría la lógica para seleccionar un worker
}

// ver si lo paso a worker.go
func (d *Dispatcher) suscribirHandler(conn net.Conn,params map[string]string) {
    
	// Verificar si el worker ya está registrado
	workerURL, ok := params["url"]
	cleanWorkerURL := strings.ReplaceAll(workerURL, "%3A", ":")
    cleanWorkerURL = strings.ReplaceAll(cleanWorkerURL, "%2F", "/")
	cleanWorkerURL = strings.Replace(cleanWorkerURL, "https:/", "", 1)
    cleanWorkerURL = strings.ReplaceAll(cleanWorkerURL, "//", "/") 
	log.Printf("Url del worker: %v", cleanWorkerURL)
	if !ok || cleanWorkerURL == "" {
        utils.SendResponse(conn, "400 Bad Request", "URL del worker requerida")
        return
    }
	log.Printf("Intento de registro de worker: %s", cleanWorkerURL)

    d.Mu.Lock()
    defer d.Mu.Unlock()
	// Verificar si el worker ya está registrado
    for _, w := range d.Workers {
        if w.URL == cleanWorkerURL {
            log.Printf("Worker ya registrado: %s", cleanWorkerURL)
            utils.SendResponse(conn, "200 OK", `{"status": "already_registered"}`)
            return
        }
    }
// Crear nuevo worker
    workerID := len(d.Workers) + 1
    newWorker := &Worker{
        ID:           workerID,
        URL:          cleanWorkerURL,
        Status:       true,
        lastChecked:  time.Now(),
        activeTasks:  0,
        taskQueue:    make(chan *Task, 1000),
    }

    d.Workers = append(d.Workers, newWorker)
    d.Metrics.mu.Lock()
    d.Metrics.WorkersRegistered++
    d.Metrics.mu.Unlock()

    log.Printf("Worker %d registrado en %s", workerID, cleanWorkerURL)
    
    // Construir respuesta similar a sendToWorker
    response := fmt.Sprintf(`{"id": "%d", "status": "registered"}`, workerID)
    
    // se da una respuesta al worker
    var responseBuilder strings.Builder
    responseBuilder.WriteString("HTTP/1.1 200 OK\r\n")
    responseBuilder.WriteString("Content-Type: application/json\r\n")
    responseBuilder.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(response)))
    responseBuilder.WriteString("\r\n")
    responseBuilder.WriteString(response)
    
    _, err := conn.Write([]byte(responseBuilder.String()))
    if err != nil {
        log.Printf("Error enviando respuesta de registro: %v", err)
    }
}
