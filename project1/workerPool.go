package main

import (
	"log"
	"sync"
)

// WorkerPool
type WorkerPool struct {
	cantidadW    int
	RequestChan  chan Request
	Wg           sync.WaitGroup
	WorkerChan   chan chan Request
	Workers      []*Worker
	ShutDownChan chan struct{}
}

func NewWorkerPool(cantidadW int) *WorkerPool {
	return &WorkerPool{
		cantidadW:    cantidadW,
		RequestChan:  make(chan Request),
		WorkerChan:   make(chan chan Request),
		ShutDownChan: make(chan struct{}),
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.cantidadW; i++ {
		worker := NewWorker(i, wp.WorkerChan, wp.RequestChan)
		wp.Workers = append(wp.Workers, worker)
		wp.Wg.Add(1)
		go worker.Start(wp)
	}

	go wp.dispatch()
	log.Printf("WorkerPool iniciado con %d workers", wp.cantidadW)
}

func (wp *WorkerPool) dispatch() {
	for {
		select {
		case req := <-wp.RequestChan:
			// Esperar un worker disponible
			workerChan := <-wp.WorkerChan
			// Enviar la solicitud al worker
			workerChan <- req

		case <-wp.ShutDownChan:
			// Cerrar todos los workers
			for _, worker := range wp.Workers {
				close(worker.ShutDownChan)
			}
			return
		}
	}
}
