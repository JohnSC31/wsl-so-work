// worker.go (en módulo dispatcher)
package main

import (
	"sync"
	"time"
)

type Worker struct {
	ID            int
	URL           string // Ej: "http://worker1:8080"
	Status        bool
	mu            sync.RWMutex // RWMutex para permitir lecturas concurrentes
	lastChecked   time.Time
	activeTasks   int
	maxCapacity   int        // Máximo de tareas concurrentes
	taskQueue     chan *Task // Canal interno para manejar carga
	healthChecker *time.Ticker
	cargadas      int        // Contador de tareas cargadas
	tasksDone     chan *Task // Canal para tareas listas
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

/*func (w *Worker) startHealthCheck() {
	w.healthChecker = time.NewTicker(10 * time.Second)
	go func() {
		for range w.healthChecker.C {
			if !w.checkHealth() {
				w.mu.Lock()
				w.Status = false
				w.mu.Unlock()
			}
		}
	}()

}*/
