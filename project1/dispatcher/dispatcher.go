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
	"strconv"
)

const (
	DispatcherPort      = ":8080"
	HealthCheckInterval = 10 * time.Second
	WorkerTimeout       = 10 * time.Second
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


// Estructura para la respuesta de conteo de palabras de un worker (reutilizada para Pi)
type WorkerResult struct { // Renombrada para ser más genérica
	WorkerID string
	Count    int   // Puede ser wordCount o pointsInCircle
	Error    error
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
			"/wordcount", // Nuevo comando para conteo de palabras
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
			if err != nil {
				w.mu.Lock()
				w.Status = false
				w.mu.Unlock()
				log.Printf("Worker %d (%s) inactivo: %v", w.ID, w.URL, err)
				redistribuirTareas(w)
				return
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

	// Utiliza bufio.Reader para leer la solicitud línea por línea de forma eficiente
	reader := bufio.NewReader(conn)

	// Leer la primera línea de la solicitud (e.g., "GET /path HTTP/1.1")
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error leyendo request line: %v", err)
		utils.SendResponse(conn, "400 Bad Request", "Error leyendo la solicitud HTTP")
		return
	}

	method, path := utils.ParseRequestLine(requestLine)

	route, params := utils.ParseRoute(path)
	/*if !d.isValidRoute(route) {
		utils.SendResponse(conn, "404 Not Found", "Ruta no encontrada")
		return
	}*/

	// Leer los encabezados HTTP
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error leyendo headers: %v", err)
			utils.SendResponse(conn, "400 Bad Request", "Error leyendo encabezados HTTP")
			return
		}
		if strings.TrimSpace(line) == "" { // Línea vacía indica fin de encabezados
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers[key] = value
		}
	}
	
	// Sumar a las metricas
	d.Metrics.mu.Lock()
	d.Metrics.TotalRequests++
	d.Metrics.mu.Unlock()

	if route == "/countwords" && method == "POST" {
		log.Println("Received /countwords POST request.")
		d.handleWordCount(conn, method, route, params, headers, reader)
		return 
	}

	// Cálculo de Pi (GET con parámetros)
	if route == "/calculatepi" && method == "GET" {
		log.Println("Received /calculatepi GET request.")
		d.handleCalculatePi(conn, params)
		return
	}

	if method != "GET" {
		utils.SendResponse(conn, "405 Method Not Allowed", "Solo se permite GET y POST")
		return
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

	// agregar la tarea al canal de tareas
	// Eliminarlo al obtener la respuesta para evitar que sea reenviada por el health check

	worker := seleccionarWorker(d)
	path = worker.URL + route
	print("Ruta del worker: ", path, "\n")

	if worker == nil {
		utils.SendResponse(conn, "503 Service Unavailable", "No hay workers disponibles")
		d.Metrics.mu.Lock()
		d.Metrics.RequestsFailed++
		d.Metrics.mu.Unlock()
		return
	}
	log.Printf("Enviando tarea %s a worker %d (%s)", newTask.ID, worker.ID, worker.URL)
	err = d.sendToWorker(worker, &newTask)
	if err != nil {
		log.Printf("Error enviando tarea a worker %d: %v", worker.ID, err)
		conn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\nError al comunicarse con el worker"))
		return
	}

	//log.Printf("respuesta enviada al cliente para tarea %s", newTask.Response)
	utils.SendResponse(conn, "200 OK", string(newTask.Response))
	log.Printf("Tarea %s completada por worker %d", newTask.ID, worker.ID)
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

func redistribuirTareas(w *Worker) {
	print("Redistribuyendo tareas...\n")
	// Aquí se implementaría la lógica para redistribuir las tareas pendientes
	// a los workers activos.

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
	worker.activeTasks++
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
	responseBuilder.WriteString("\r\n") // Línea vacía que separa headers del body

	// Leer la respuesta completa
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		worker.mu.Lock()
		worker.activeTasks--
		worker.mu.Unlock()
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
	worker.activeTasks--
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


// NUEVA FUNCIÓN: Maneja la solicitud de conteo de palabras de archivos grandes
// Ahora recibe el bufio.Reader para leer el cuerpo de la solicitud POST
func (d *Dispatcher) handleWordCount(conn net.Conn, method, path string, params map[string]string, headers map[string]string, reader *bufio.Reader) {
	// 1. Recibir el contenido del archivo desde el cuerpo de la solicitud POST
	contentLengthStr, ok := headers["Content-Length"]
	var contentLength int
	if ok {
		var err error
		contentLength, err = strconv.Atoi(contentLengthStr)
		if err != nil {
			utils.SendResponse(conn, "400 Bad Request", "Content-Length inválido")
			log.Printf("Error: Content-Length inválido: %v", err)
			d.Metrics.mu.Lock()
			d.Metrics.RequestsFailed++
			d.Metrics.mu.Unlock()
			return
		}
	} else {
		log.Println("Advertencia: No Content-Length header. Leyendo hasta EOF/timeout.")
	}

	var contentBuilder strings.Builder
	var bytesRead int
	buffer := make([]byte, 4096) // Buffer para leer chunks del cuerpo
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			contentBuilder.Write(buffer[:n])
			bytesRead += n
		}
		if err == io.EOF || (contentLength > 0 && bytesRead >= contentLength) {
			break
		}
		if err != nil {
			utils.SendResponse(conn, "500 Internal Server Error", "Error leyendo el cuerpo del archivo")
			log.Printf("Error leyendo el cuerpo del archivo: %v", err)
			d.Metrics.mu.Lock()
			d.Metrics.RequestsFailed++
			d.Metrics.mu.Unlock()
			return
		}
	}
	content := contentBuilder.String()
	log.Printf("Archivo recibido, tamaño: %d bytes", len(content))

	// 2. Dividir el archivo en chunks
	lines := strings.Split(content, "\n")
	// Limpieza: si el archivo termina con \n, Split crea una última cadena vacía. La eliminamos.
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	// Si el archivo está completamente vacío después de limpiar las líneas
	if len(lines) == 0 {
		utils.SendResponse(conn, "200 OK", "Conteo total de palabras: 0\n")
		log.Println("Archivo vacío después de procesar, retorno 0 palabras.")
		d.Metrics.mu.Lock()
		d.Metrics.RequestsHandled++
		d.Metrics.mu.Unlock()
		return
	}

	numWorkers := len(d.Workers)
	if numWorkers == 0 {
		utils.SendResponse(conn, "503 Service Unavailable", "No hay workers disponibles para el conteo de palabras")
		log.Println("No hay workers disponibles para conteo de palabras")
		d.Metrics.mu.Lock()
		d.Metrics.RequestsFailed++
		d.Metrics.mu.Unlock()
		return
	}

	// Calcula el tamaño base del chunk y cuántos trabajadores recibirán un chunk extra
	baseChunkSize := len(lines) / numWorkers
	extraChunks := len(lines) % numWorkers

	var wg sync.WaitGroup
	// El tamaño del canal de resultados debe ser al menos el número de workers,
	// pero no se bloqueará si lanzamos menos goroutines.
	resultsChan := make(chan WorkerResult, numWorkers)

	// Contar los workers a los que realmente se les asignará una tarea
	workersReceivingTasks := 0 
	for i := 0; i < numWorkers; i++ {
		workerLines := baseChunkSize
		if i < extraChunks {
			workerLines++ // Distribuir las líneas restantes
		}

		// Si este worker no tiene líneas para procesar, NO LANZAMOS GOROUTINE
		if workerLines == 0 {
			log.Printf("Dispatcher: Saltando worker %d, 0 líneas asignadas.", i+1)
			continue
		}

		startLine := 0
		if i > 0 {
			// Calcular el inicio sumando los tamaños de los chunks anteriores
			for j := 0; j < i; j++ {
				prevChunkSize := baseChunkSize
				if j < extraChunks {
					prevChunkSize++
				}
				startLine += prevChunkSize
			}
		}
		endLine := startLine + workerLines

		// Protección extra, aunque con la lógica de baseChunkSize/extraChunks debería ser raro
		if startLine >= endLine || endLine > len(lines) {
			log.Printf("Dispatcher: Error lógico en la división de chunks para worker %d (start: %d, end: %d, total lines: %d). Saltando.", i+1, startLine, endLine, len(lines))
			continue
		}

		chunkContent := strings.Join(lines[startLine:endLine], "\n")
		
		worker := seleccionarWorker(d)
		if worker == nil {
			log.Printf("Dispatcher: No se pudo seleccionar worker para el chunk de worker %d, reintentando o fallando.", i+1)
			// Aquí se podría implementar una cola de reintentos o marcar la tarea como fallida
			continue
		}
		
		workersReceivingTasks++ // Solo incrementamos si realmente se lanza una goroutine

		wg.Add(1)
		go func(w *Worker, currentChunkContent string, chunkID int) {
			defer wg.Done()
			log.Printf("Dispatcher: Enviando chunk %d (tamaño %d bytes) a worker %d (%s)", chunkID, len(currentChunkContent), w.ID, w.URL)

			wordCountStr, err := d.sendPostToWorker(w, "/countchunk", currentChunkContent)
			if err != nil {
				resultsChan <- WorkerResult{WorkerID: fmt.Sprintf("Worker-%d", w.ID), Error: fmt.Errorf("error enviando chunk %d a worker %s: %w", chunkID, w.URL, err)}
				return
			}
			log.Printf("Dispatcher: Worker %d respondió con: '%s'", w.ID, wordCountStr)

			wordCount, err := strconv.Atoi(strings.TrimSpace(wordCountStr))
			if err != nil {
				resultsChan <- WorkerResult{WorkerID: fmt.Sprintf("Worker-%d", w.ID), Error: fmt.Errorf("error parseando conteo de palabras de worker %s para chunk %d: %w", w.URL, chunkID, err)}
				return
			}
			resultsChan <- WorkerResult{WorkerID: fmt.Sprintf("Worker-%d", w.ID), Count: wordCount}

		}(worker, chunkContent, i+1)
	}

	// Si no se lanzaron tareas a ningún worker (ej. archivo muy pequeño para un solo worker o no hay workers disponibles)
	if workersReceivingTasks == 0 {
		utils.SendResponse(conn, "200 OK", "Conteo total de palabras: 0 (No se enviaron tareas a workers)\n")
		log.Println("No se enviaron tareas a workers, retorno 0 palabras.")
		d.Metrics.mu.Lock()
		d.Metrics.RequestsHandled++
		d.Metrics.mu.Unlock()
		return
	}


	// Cerrar el canal de resultados cuando todas las goroutines de workers terminen
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// 4. Recopilar y sumar los resultados
	totalWordCount := 0
	var errors []error
	for res := range resultsChan {
		if res.Error != nil {
			log.Printf("Error recibido de worker %s: %v", res.WorkerID, res.Error)
			errors = append(errors, res.Error)
		} else {
			totalWordCount += res.Count
			log.Printf("Worker %s contribuyó con %d palabras.", res.WorkerID, res.Count)
		}
	}

	// 5. Retornar el resultado total al cliente
	if len(errors) > 0 {
		errMsg := fmt.Sprintf("Errores durante el procesamiento: %v. Conteo parcial: %d", errors, totalWordCount)
		utils.SendResponse(conn, "500 Internal Server Error", errMsg)
		d.Metrics.mu.Lock()
		d.Metrics.RequestsFailed++
		d.Metrics.mu.Unlock()
		return
	}

	log.Printf("Conteo total de palabras: %d", totalWordCount)
	utils.SendResponse(conn, "200 OK", fmt.Sprintf("Conteo total de palabras: %d\n", totalWordCount))
	d.Metrics.mu.Lock()
	d.Metrics.RequestsHandled++
	d.Metrics.mu.Unlock()
}

// NUEVA FUNCIÓN: sendPostToWorker
// Envía una solicitud POST HTTP manual a un worker con el comando y el cuerpo de contenido.
// Retorna el cuerpo de la respuesta del worker (el conteo de palabras como string) o un error.
func (d *Dispatcher) sendPostToWorker(worker *Worker, command string, content string) (string, error) {
	workerHost := strings.Split(worker.URL, ":")[0] // Obtener solo el host para el header Host
	// workerPort := strings.Split(worker.URL, ":")[1] // Obtener el puerto

	requestBody := []byte(content)
	requestHeaders := []string{
		fmt.Sprintf("POST %s HTTP/1.1", command),
		fmt.Sprintf("Host: %s", workerHost),
		fmt.Sprintf("Content-Type: text/plain"),
		fmt.Sprintf("Content-Length: %d", len(requestBody)),
		"Connection: close", // Indicar al worker que cierre la conexión después de la respuesta
		"",                  // Línea vacía para separar headers del body
	}
	fullRequest := strings.Join(requestHeaders, "\r\n") + "\r\n" + string(requestBody)

	log.Printf("Full request (%s)", fullRequest)

	// Bloquear worker para actualizar estado (esto se gestionará a un nivel superior si es una "tarea" general)
	// Aquí solo estamos enviando la solicitud, la gestión de activeTasks del worker
	// se hará en el handleWordCount o en un nivel de orquestación de tareas más general si se crea.
	// Por ahora, lo mantenemos simple para esta función específica de comunicación.

	// Establecer conexión TCP con el worker
	workerConn, err := net.DialTimeout("tcp", worker.URL, WorkerTimeout)
	if err != nil {
		return "", fmt.Errorf("error conectando a worker %s: %w", worker.URL, err)
	}
	defer workerConn.Close()

	// Enviar la solicitud HTTP al worker
	_, err = workerConn.Write([]byte(fullRequest))
	if err != nil {
		return "", fmt.Errorf("error enviando solicitud a worker %s: %w", worker.URL, err)
	}

	// Leer la respuesta del worker
	workerReader := bufio.NewReader(workerConn)
	responseStatusLine, err := workerReader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error leyendo status line de worker %s: %w", worker.URL, err)
	}
	if !strings.Contains(responseStatusLine, "200 OK") {
		// Leer el resto de la respuesta para el log de error
		responseBody, _ := io.ReadAll(workerReader)
		return "", fmt.Errorf("worker %s retornó status no OK: %s - %s", worker.URL, strings.TrimSpace(responseStatusLine), string(responseBody))
	}

	// Leer y descartar headers del worker
	for {
		line, err := workerReader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("error leyendo headers de worker %s: %w", worker.URL, err)
		}
		if strings.TrimSpace(line) == "" {
			break // Fin de los headers
		}
	}

	log.Printf("Before reading the body of worker %s", worker.URL)

	// Leer el cuerpo de la respuesta (el conteo de palabras)
	wordCountBody, err := io.ReadAll(workerReader)
	if err != nil {
		return "", fmt.Errorf("error leyendo body de worker %s: %w", worker.URL, err)
	}

	log.Printf("Result worker %s: %d", worker.URL, wordCountBody)

	return strings.TrimSpace(string(wordCountBody)), nil
}

// handleCalculatePi: Nueva función para coordinar el cálculo distribuido de Pi
func (d *Dispatcher) handleCalculatePi(conn net.Conn, params map[string]string) {
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

		worker := seleccionarWorker(d)
		if worker == nil {
			log.Printf("No se pudo seleccionar worker para la tarea de Pi del worker %d, saltando.", i)
			continue
		}

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

		}(worker, workerIterations, i+1)
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

// sendGetToWorker: Nueva función para enviar solicitudes GET manuales a un worker.
// Retorna el cuerpo de la respuesta del worker o un error.
func (d *Dispatcher) sendGetToWorker(worker *Worker, command string, params map[string]string) (string, error) {
	workerHost := strings.Split(worker.URL, ":")[0]

	// Construir los parámetros de la URL
	queryParams := ""
	if len(params) > 0 {
		var q []string
		for k, v := range params {
			q = append(q, fmt.Sprintf("%s=%s", k, v))
		}
		queryParams = "?" + strings.Join(q, "&")
	}

	requestHeaders := []string{
		fmt.Sprintf("GET %s%s HTTP/1.1", command, queryParams),
		fmt.Sprintf("Host: %s", workerHost),
		"Connection: close", // Indicar al worker que cierre la conexión después de la respuesta
		"",                  // Línea vacía final para separar headers del body (aunque no hay body en GET)
	}
	fullRequest := strings.Join(requestHeaders, "\r\n") + "\r\n"

	workerConn, err := net.DialTimeout("tcp", worker.URL, WorkerTimeout)
	if err != nil {
		return "", fmt.Errorf("error conectando a worker %s: %w", worker.URL, err)
	}
	defer workerConn.Close()

	_, err = workerConn.Write([]byte(fullRequest))
	if err != nil {
		return "", fmt.Errorf("error enviando solicitud GET a worker %s: %w", worker.URL, err)
	}

	workerReader := bufio.NewReader(workerConn)
	responseStatusLine, err := workerReader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error leyendo status line de worker %s: %w", worker.URL, err)
	}
	if !strings.Contains(responseStatusLine, "200 OK") {
		responseBody, _ := io.ReadAll(workerReader)
		return "", fmt.Errorf("worker %s retornó status no OK: %s - %s", worker.URL, strings.TrimSpace(responseStatusLine), string(responseBody))
	}

	// Leer y descartar headers del worker
	for {
		line, err := workerReader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("error leyendo headers de worker %s: %w", worker.URL, err)
		}
		if strings.TrimSpace(line) == "" {
			break // Fin de los headers
		}
	}

	// Leer el cuerpo de la respuesta (el resultado del cálculo)
	responseBody, err := io.ReadAll(workerReader)
	if err != nil {
		return "", fmt.Errorf("error leyendo body de worker %s: %w", worker.URL, err)
	}

	return strings.TrimSpace(string(responseBody)), nil
}