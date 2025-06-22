package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"http-servidor/utils"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	DispatcherPort      = ":8080"
	HealthCheckInterval = 10 * time.Second
	WorkerTimeout       = 100 * time.Second
	EstrategiaRed       = 1 //cambiar a 2 si se quiere usar least loaded
	primero             = 1 // Usar round robin para seleccionar el primer worker
)

type Dispatcher struct {
	Workers         []*Worker
	TasksChan       chan *Task
	Tasks           sync.Map // Concurrent map para tareas activas [taskID]*Task
	Listener        net.Listener
	Mu              sync.RWMutex
	DoneChan        chan struct{}
	Commands        []string // Pool de workers por comando
	Metrics         *DispatcherMetrics
	lastWorkerIndex int
}

type DispatcherMetrics struct {
	RequestsHandled   int
	RequestsFailed    int
	TotalRequests     int
	WorkersRegistered int
	StartTime         time.Time

	mu sync.Mutex
}

type PendingTask struct {
	Conn     net.Conn
	Response chan []byte
}

var (
	pendingTasks = make(map[string]*PendingTask) // requestID → PendingTask
	pendingMu    sync.RWMutex
)

// inicializa el dispatcher y los workers
func newDispatcher() *Dispatcher {
	print("Inicializando Dispatcher...\n")
	metrics := &DispatcherMetrics{
		StartTime:         time.Now(),
		RequestsHandled:   0,
		RequestsFailed:    0,
		TotalRequests:     0,
		WorkersRegistered: 0,
	}

	dispatcher := &Dispatcher{
		Workers:   make([]*Worker, 0),
		TasksChan: make(chan *Task, 1000), // Canal para recibir tareas
		DoneChan:  make(chan struct{}),
		Commands: []string{
			"/help",
			"/timestamp",
			"/fibonacci",
			"/createfile",
			"/deletefile",
			"/reverse",
			"/toupper",
			"/random",
			"/hash",
			"/simulate",
			"/sleep",
			"/loadtest",
		},
		Metrics: metrics,
	}

	workerURL := []string{
		"worker1:8080",
		"worker2:8080",
		"worker3:8080",
		//"worker4:8084",
		//"worker5:8085",
	}

	for i, url := range workerURL {
		worker := NewWorker(i+1, url, 10) // Cada worker con capacidad de 10 tareas concurrentes
		dispatcher.Workers = append(dispatcher.Workers, worker)
		dispatcher.Metrics.mu.Lock()
		dispatcher.Metrics.WorkersRegistered++
		dispatcher.Metrics.mu.Unlock()
		dispatcher.Workers[i].Status = true // Inicialmente todos los workers están activos
		dispatcher.Workers[i].lastChecked = time.Now()
		dispatcher.Workers[i].activeTasks = 0
		dispatcher.Workers[i].taskQueue = make(chan *Task, 100)
		log.Printf("Worker %d registrado en %s", i+1, url)
	}

	//dispatcher.HealthCheck()

	return dispatcher
}

// hace health checks periodicamente a los workers
func (d *Dispatcher) HealthCheck() {
	print("Iniciando health check de workers...\n")
	for _, worker := range d.Workers {
		go func(w *Worker) {
			conn, err := net.DialTimeout("tcp", w.URL, 5*time.Second)
			log.Printf("Estado del worker %d (%s)", w.ID, w.Status)
			if err != nil {
				w.mu.Lock()
				w.Status = false
				w.mu.Unlock()
				log.Printf("Worker %d (%s) inactivo: %v", w.ID, w.URL, err)
				tareaPendientes := make([]*Task, 0)
				for {
					select {
					case t := <-w.taskQueue:
						tareaPendientes = append(tareaPendientes, t)
					default:
						for _, tarea := range tareaPendientes {
							log.Printf("Redistribuyendo tarea %s desde worker %d", tarea.ID, w.ID)
							workerR := seleccionarWorker(d)
							if workerR == nil {
								log.Printf("No hay workers disponibles para redistribuir la tarea %s", tarea.ID)

							}
							workerR.mu.Lock()
							workerR.cargadas++ // Incrementamos la carga del nuevo worker
							workerR.mu.Unlock()
							workerR.taskQueue <- tarea // Redistribuir la tarea al nuevo worker
							log.Printf("Tarea %s redistribuida al worker %d", tarea.ID, workerR.ID)
							// Aquí se podría enviar una notificación al cliente si es necesario
						}

					}

				}

			}
			defer conn.Close()

			// Enviar solicitud HTTP válida
			_, err = fmt.Fprintf(conn, "GET /ping HTTP/1.1\r\nHost: %s\r\n\r\n", w.URL)
			if err != nil {
				w.mu.Lock()
				w.Status = false
				w.mu.Unlock()
				return
			}

			// Leer respuesta
			scanner := bufio.NewScanner(conn)
			scanner.Scan() // Lee la primera línea (HTTP status)
			if strings.Contains(scanner.Text(), "200 OK") {
				w.mu.Lock()
				w.Status = true
				w.mu.Unlock()
			}
			log.Printf("Worker %d (%s) estado: %t", w.ID, w.URL, w.Status)
		}(worker)
	}
}

// maneja la conexion, crea la nueva tarea, asigan la nueva tarea y envia la solicitud al servidor del worker
func (d *Dispatcher) HandleConnection(conn net.Conn) {

	print("Nueva conexión aceptada\n")
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Error leyendo conexión: %v", err)
		return
	}

	request := string(buffer[:n])
	method, path := utils.ParseRequestLine(request)

	// creo que podria dar problemas con los workers que envian solicitudes POST
	/*if method != "GET" {
		utils.SendResponse(conn, "405 Method Not Allowed", "Solo se permite GET")
		return
	}*/
	route, params := utils.ParseRoute(path)
	/*if !d.isValidRoute(route) {
		utils.SendResponse(conn, "404 Not Found", "Ruta no encontrada")
		return
	}*/

	d.Metrics.mu.Lock()
	d.Metrics.TotalRequests++
	d.Metrics.mu.Unlock()

	newRequest := Request{
		Method: method,
		Path:   route,
		Params: params,
		Done:   make(chan bool),
	}

	newTask := Task{
		ID:         fmt.Sprintf("%d", d.Metrics.TotalRequests),
		Conn:       conn,
		Request:    &newRequest,
		Response:   nil,
		Status:     TaskPending,
		CreatedAt:  time.Now(),
		RetryCount: 0,
	}

	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	responseChan := make(chan []byte)

	pendingMu.Lock()
	pendingTasks[requestID] = &PendingTask{Conn: conn, Response: responseChan}
	pendingMu.Unlock()

	// Añade request_id a los parámetros
	params["request_id"] = requestID
	params["callback_url"] = "http://dispatcher:8080/callback"

	worker := seleccionarWorker(d)
	if worker == nil {
		utils.SendResponse(conn, "503 Service Unavailable", "No workers available")
		return
	}

	worker.taskQueue <- &newTask
	//envia la tarea al worker
	go func() {
		err := d.sendToWorker(worker, path, params, &newTask)
		if err != nil {
			utils.SendResponse(conn, "502 Bad Gateway", "Failed to send to worker")
		}
	}()

	// espera asincronicamente la respuesta
	go func() {
		select {
		case response := <-responseChan:
			conn.Write(response)
		case <-time.After(10 * time.Second):
			utils.SendResponse(conn, "504 Gateway Timeout", "Worker timeout")
		}
		conn.Close()
	}()
}

// revisa si el endpoint existe
func (d *Dispatcher) isValidRoute(route string) bool {
	for _, cmd := range d.Commands {
		if route == cmd {
			return true
		}
	}
	return false
}

// selecciona el worker que se va a usar para procesar la tarea
func seleccionarWorker(d *Dispatcher) *Worker {
	// Estrategia de round robin
	if EstrategiaRed == 1 {
		for i := 0; i < len(d.Workers); i++ {
			// Buscar el siguiente worker disponible después del último usado
			idx := (d.lastWorkerIndex + i + 1) % len(d.Workers)
			worker := d.Workers[idx]

			if worker.Status == true {
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
			if worker.Status == true {
				// Si es el primer worker disponible o tiene menos carga
				if minLoad == -1 || worker.cargadas < minLoad {
					minLoad = worker.cargadas
					selectedWorker = worker
				}
			}
		}

		if selectedWorker != nil {
			selectedWorker.cargadas++ // Incrementamos su carga
			return selectedWorker
		}

	}
	return nil // Aquí se implementaría la lógica para seleccionar un worker
}

func generateRequestID() string {
	b := make([]byte, 8) // 16 caracteres hex
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("req-%d", time.Now().UnixNano())
	}
	return "req-" + hex.EncodeToString(b)
}

func (d *Dispatcher) sendToWorker(worker *Worker, path string, params map[string]string, task *Task) error {
	// Añadir parámetros esenciales
	/*params["request_id"] = generateRequestID()

	params["callback_url"] = "http://dispatcher:8080/callback"*/

	// Construir URL
	query := url.Values{}
	for k, v := range params {
		query.Add(k, v)

	}
	//log.Printf("Enviando solicitud al worker %d: %s?%s", worker.ID, path, query.Encode())
	workerURL := fmt.Sprintf("http://%s%s?%s", worker.URL, path, query.Encode())
	log.Printf("URL enviada al worker: %s", workerURL)

	// Enviar solicitud
	resp, err := http.Get(workerURL)
	if err != nil {
		return fmt.Errorf("error al contactar worker: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("worker respondió con status inesperado: %s", resp.Status)
	}

	return nil
}

func encodeParams(params map[string]string) string {
	var values url.Values = make(url.Values)
	for k, v := range params {
		values.Add(k, v)
	}
	return values.Encode()
}

func (d *Dispatcher) HandleCallback(w http.ResponseWriter, r *http.Request) {
	// Parsear formulario
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	requestID := r.FormValue("request_id")
	result := r.FormValue("result")

	if requestID == "" || result == "" {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	pendingMu.Lock()
	task, exists := pendingTasks[requestID]
	pendingMu.Unlock()

	if !exists {
		http.Error(w, "Unknown request ID", http.StatusNotFound)
		return
	}

	// Enviar respuesta al cliente original
	if _, err := task.Conn.Write([]byte(result)); err != nil {
		log.Printf("Error enviando respuesta al cliente: %v", err)
	}

	// Limpiar
	pendingMu.Lock()
	delete(pendingTasks, requestID)
	pendingMu.Unlock()

	w.WriteHeader(http.StatusOK)
}
