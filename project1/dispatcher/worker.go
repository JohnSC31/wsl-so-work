// worker.go (en módulo dispatcher)
package main

import (
	"sync"
	"time"
	"encoding/json"
	"log"
	"net"
	"http-servidor/utils"

)

type Worker struct {
	ID            int
	URL           string // Ej: "http://worker1:8080"
	Status        bool
	mu            sync.RWMutex // Usamos RWMutex para permitir lecturas concurrentes
	lastChecked   time.Time
	activeTasks   int
	maxCapacity   int        // Máximo de tareas concurrentes
	taskQueue     chan *Task // Canal interno para manejar carga
	healthChecker *time.Ticker
	CompletedTasks      int // Contador de tareas cargadas
}

func NewWorker(id int, url string, capacity int) *Worker {
	w := &Worker{
		ID:          id,
		URL:         url,
		Status:      true,
		maxCapacity: capacity,
		taskQueue:   make(chan *Task, capacity),
	}

	//w.startHealthCheck()
	return w
}


//devuelve el estado del worker
func workerStatus(conn net.Conn, d *Dispatcher) {
	d.Metrics.mu.Lock()
	uptime := time.Since(d.Metrics.StartTime).Truncate(time.Second).String()
	totalRequests := d.Metrics.TotalRequests
	d.Metrics.mu.Unlock()

	workerActivo := 0

	// Armamos una estructura por worker
	workersStatus := make([]map[string]interface{}, 0)
	for _, worker := range d.Workers {
		if worker.Status {
			workerActivo ++
			log.Printf("Worker %d (%s) activo", worker.ID, worker.URL)
		} 
		worker.mu.RLock()
		status := map[string]interface{}{
			"pid":            worker.ID,
			"url":           worker.URL,
			"status":        worker.Status,
			"active_tasks":  worker.activeTasks,
			"CompletedTasks":      worker.CompletedTasks,
			"last_checked":  worker.lastChecked.Format(time.RFC3339),
			"max_capacity":  worker.maxCapacity,
		}
		worker.mu.RUnlock()
		workersStatus = append(workersStatus, status)
	}

	response := map[string]interface{}{
		"main_pid":        d.ID, // PID del proceso principal
		"uptime":          uptime,
		"total_requests":  totalRequests,
		"workers_status":  workersStatus,
		"total_workers":   workerActivo,
	}
	jsonData, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
			utils.SendResponse(conn, "500 Internal Server Error", "Error generando JSON")
			return
		}
	utils.SendJSON(conn, "200 OK", jsonData)
}