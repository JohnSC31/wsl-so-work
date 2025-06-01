package main

import "fmt"

// Worker
type Worker struct {
	ID           int
	RequestChan  chan Request
	WorkerChan   chan chan Request
	ReqActual    *Request
	ShutDownChan chan struct{}
	Status		string
}

func NewWorker(id int, workerChan chan chan Request, requestChan chan Request) *Worker {
	return &Worker{
		ID:           id,
		RequestChan:  requestChan,
		WorkerChan:   workerChan,
		ShutDownChan: make(chan struct{}),
		Status: 	"disponible",
	}
}

func (w *Worker) Start(wp *WorkerPool) {

	defer wp.Wg.Done()
	for {
		// 1. Notificar al pool que este worker está disponible
		wp.WorkerChan <- w.RequestChan 

		select {
		case req := <-w.RequestChan:
			fmt.Printf("Worker %d recibió solicitud %d", w.ID, req.ID)
			// 2. Actualizar estado del worker
			w.ReqActual = &req
			w.Status = "ocupado"

			// 3. Procesar la solicitud
			HandleRequest(req)

			// 4. Limpiar estado
			w.ReqActual = nil
			w.Status = "disponible"

		case <-w.ShutDownChan:
			// 5. Salir si se recibe señal de shutdown
			return
		}
	}
}
