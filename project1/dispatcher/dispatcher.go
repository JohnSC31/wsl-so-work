package main

import (
	"bufio"
	"fmt"
	"http-servidor/utils"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
	"encoding/json"
)

const (
	DispatcherPort      = ":8080"
	HealthCheckInterval = 10 * time.Second
	WorkerTimeout       = 10 * time.Second
	EstrategiaRed       = 1 //cambiar a 2 si se quiere usar least loaded
	primero             = 1 // Usar round robin para seleccionar el primer worker
)

type Dispatcher struct {
	ID			 	int
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
		ID:         1, // ID del dispatcher
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
		dispatcher.Workers[i].taskQueue = make(chan *Task, 1000)
		log.Printf("Worker %d registrado en %s", i+1, url)
	}

	//dispatcher.HealthCheck()

	return dispatcher
}

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

                }
            }
            return
        }
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

	if method != "GET" {
		utils.SendResponse(conn, "405 Method Not Allowed", "Solo se permite GET")
		return
	}
	route, params := utils.ParseRoute(path)
	/*if !d.isValidRoute(route) {
		utils.SendResponse(conn, "404 Not Found", "Ruta no encontrada")
		return
	}*/
	if route == "/workers" {
		workerStatus(conn, d)
		return
	}

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
		ID:         d.Metrics.TotalRequests,
		Conn:       conn,
		Request:    &newRequest,
		Response:   nil,
		Status:     TaskPending,
		CreatedAt:  time.Now(),
		RetryCount: 0,
	}

	worker := seleccionarWorker(d)
	if worker == nil {
		utils.SendResponse(conn, "503 Service Unavailable", "No hay workers disponibles")
		d.Metrics.mu.Lock()
		d.Metrics.RequestsFailed++
		d.Metrics.mu.Unlock()
		return
	}

	path = worker.URL + route
	print("Ruta del worker: ", path, "\n")

	if !d.checkWorkerStatus(worker) {
        log.Printf("Worker %d (%s) marcado como inactivo", worker.ID, worker.URL)
		log.Printf("Cantidad de %s tareas pendientes del worker %d", len(worker.taskQueue), worker.ID)
		d.redistributeTasks(worker)
        utils.SendResponse(conn, "503 Service Unavailable", "Worker no disponible")
        return
    }

	
	log.Printf("Enviando tarea %s a worker %d (%s)", newTask.ID, worker.ID, worker.URL)

	worker.taskQueue <- &newTask

	worker.mu.Lock()
	worker.CompletedTasks++ // Incrementamos la carga del worker
	worker.activeTasks++ // Incrementamos el contador de tareas activas
	worker.mu.Unlock()

	log.Printf("Tareas %s asignadas al worker %d", len(worker.taskQueue), worker.ID)
	err = d.sendToWorker(worker, &newTask)
	if err != nil {
		log.Printf("Error enviando tarea a worker %d: %v", worker.ID, err)
		// Volver a verificar estado
        if !d.checkWorkerStatus(worker) {
            // Redistribuir tareas pendientes si es necesario
            d.redistributeTasks(worker)
        }
		conn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\nError al comunicarse con el worker"))
		return
	}

	//log.Printf("respuesta enviada al cliente para tarea %s", newTask.Response)
	//taskFinalizada := <-worker.taskQueue
	worker.mu.Lock()
	worker.activeTasks-- // Decrementamos el contador de tareas activas
	worker.mu.Unlock()
	//log.Printf("Tarea %s completada por worker %d y sacada de la cola", taskFinalizada.ID, worker.ID)
	utils.SendResponse(conn, "200 OK", string(newTask.Response))
	log.Printf("Tarea %s completada por worker %d", newTask.ID, worker.ID)
	worker.cleanCompletedTasks() // Limpiar tareas completadas del worker

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


func (d *Dispatcher) sendToWorker(worker *Worker, task *Task) error {
	// Construir URL
	url := fmt.Sprintf("http://%s%s", worker.URL, task.Request.Path)

	// Crear solicitud HTTP
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creando request: %v", err)
	}

	// Añadir parámetros
	q := req.URL.Query()
	for key, value := range task.Request.Params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	// Configurar headers
	req.Header.Set("Host", worker.URL)
	req.Header.Set("X-Request-ID", fmt.Sprintf("%d", task.ID))

	// Bloquear worker para actualizar estado
	worker.mu.Lock()
	//worker.activeTasks++
	task.Status = TaskProcessing
	worker.mu.Unlock()

	// Enviar solicitud con timeout
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		worker.mu.Lock()
		worker.activeTasks--
		worker.mu.Unlock()
		
		return fmt.Errorf("error enviando a worker: %v", err)
	}
	defer resp.Body.Close()

	var responseBuilder strings.Builder
	responseBuilder.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", resp.StatusCode, resp.Status))

	for k, v := range resp.Header {
		responseBuilder.WriteString(fmt.Sprintf("%s: %s\r\n", k, strings.Join(v, ", ")))
	}
	responseBuilder.WriteString("\r\n") // separa headers del body

	// Leer la respuesta completa
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error leyendo respuesta: %v", err)
	}
	responseBuilder.Write(body)

	// Guardar la respuesta en la tarea
	fullResponse := responseBuilder.String()
	task.Response = []byte(fullResponse)
	

	worker.mu.Lock()
	task.Response = body
	task.Status = TaskCompleted
	task.CompletedAt = time.Now()
	//worker.activeTasks--
	worker.mu.Unlock()

	_, err = task.Conn.Write([]byte(fullResponse))
	if err != nil {
		return fmt.Errorf("error escribiendo al cliente: %v", err)
	}

	// Forzar flush si es necesario
	if conn, ok := task.Conn.(interface{ Flush() error }); ok {
		conn.Flush()
	}

	return nil
}


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
			log.Printf("PUTAAA Worker %d (%s) activo", worker.ID, worker.URL)
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


func (w *Worker) cleanCompletedTasks() {
    w.mu.Lock()
    defer w.mu.Unlock()
    
    // Limpiar queue sin bloquear
    for {
        select {
        case <-w.taskQueue:
            // Simplemente vaciar
        default:
            return
        }
    }
}