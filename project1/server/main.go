package main

import (
	"encoding/json"
	"http-servidor/utils"
	"log"
	"net"
	"os"
	"sync"
	"time"
	"bufio"
	"io"
	"strings"
	"math/rand" // Para generar números aleatorios
	"strconv"
	"fmt"
)

// CONSTANTES
const PORT = ":8080"

// Structs
// Request
type Request struct {
	ID           int
	Conn         net.Conn
	Ruta         string
	Parametros   map[string]string
	TiempoInicio time.Time
	Listo        chan bool
	Body		 string 
}

// Server
type Server struct {
	ServerId     int
	CommandPools map[string]*WorkerPool
	Metrics      *Metricas
	listener     net.Listener  // Socket subyacente
	doneChan     chan struct{} // Para shutdown
}

// Metricas del servidor
type Metricas struct {
	Mu            sync.Mutex
	TiempoInicio  time.Time
	TotalRequests int
	ActWorkers    int
}

// Funcion para inicializar el servidor
func NewServer() *Server {
	return &Server{
		ServerId: 1,
		CommandPools: map[string]*WorkerPool{
			"/help":       NewWorkerPool(2),
			"/timestamp":  NewWorkerPool(2),
			"/fibonacci":  NewWorkerPool(3),
			"/reverse":    NewWorkerPool(2),
			"/toupper":    NewWorkerPool(2),
			"/hash":       NewWorkerPool(2),
			"/random":     NewWorkerPool(2),
			"/simulate":   NewWorkerPool(3),
			"/sleep":      NewWorkerPool(3),
			"/loadtest":   NewWorkerPool(3),
			"/createfile": NewWorkerPool(3),
			"/deletefile": NewWorkerPool(3),
			"/ping":       NewWorkerPool(2),
		},
		Metrics: &Metricas{
			TiempoInicio:  time.Now(),
			TotalRequests: 0,
			ActWorkers:    0,
		},
		doneChan: make(chan struct{}),
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Puerto por defecto
	}

	Server := NewServer()
	log.Printf("Servidor iniciado en :%s", port)
	for _, pool := range Server.CommandPools {
		pool.Start()
	}

	ln, err := net.Listen("tcp", PORT)
	if err != nil {
		log.Fatalf("Error al iniciar servidor: %v", err)
	}
	defer ln.Close()
	log.Printf("Servidor escuchando en %s", PORT)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error aceptando conexión: %v", err)
			continue
		}
		go handleConnection(conn, Server)
	}
}

// Gestion las solicitudes que le llegan al servidor
func handleConnectionOld(conn net.Conn, server *Server) {
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Error leyendo: %v", err)
		return
	}

	request := string(buffer[:n])
	method, path := utils.ParseRequestLine(request)

	if method != "GET" || method != "POST" {
		utils.SendResponse(conn, "405 Method Not Allowed", "Solo se permite GET y POST")
		return
	}

	route, params := utils.ParseRoute(path)

	// Se incrementa el contador de solicitudes
	server.Metrics.Mu.Lock()
	server.Metrics.TotalRequests++
	server.Metrics.Mu.Unlock()

	newRequest := Request{
		ID:           server.Metrics.TotalRequests + 1,
		Conn:         conn,
		Ruta:         route,
		Parametros:   params,
		TiempoInicio: time.Now(),
		Listo:        make(chan bool),
	}

	print("Request ID: ", newRequest.ID, " Ruta: ", newRequest.Ruta, "\n")

	if route == "/status" {

		serverStatus(conn, server)

	} else if pool, exists := server.CommandPools[route]; exists {
		// Enviar la solicitud al pool correspondiente
		pool.RequestChan <- newRequest
	} else {
		// Ruta no encontrada
		utils.SendResponse(conn, "404 Not Found", "Ruta no encontrada")
	}
	newRequest.Listo <- true
}

// NUEVA FUNCIÓN PARA MANEJAR POST /countchunk y otros comandos
func handleConnection(conn net.Conn, server *Server) {
	defer conn.Close()
	log.Printf("Worker: Handle connection started.")

	reader := bufio.NewReader(conn)

	requestLineWithCRLF, err := reader.ReadString('\n')
	if err != nil {
		if err != io.EOF {
			log.Printf("Worker: Error leyendo request line: %v", err)
			utils.SendResponse(conn, "400 Bad Request", "Error leyendo la solicitud HTTP")
		}
		return
	}

	requestLine := strings.TrimSpace(requestLineWithCRLF)

	log.Printf("Worker: Parsed Request Line (trimmed): '%s'", requestLine) // <-- Log crucial

	// Tu utils.ParseRequestLine devuelve 2 valores, así que se asigna a 2.
	method, pathAndQuery := utils.ParseRequestLine(requestLine)
	route, params := utils.ParseRoute(pathAndQuery)

	// Leer los encabezados HTTP
	headers := make(map[string]string)
	log.Printf("Worker: Starting header read loop...") // Log para depuración
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Worker: Error durante lectura de header: %v", err)
			utils.SendResponse(conn, "400 Bad Request", "Error leyendo encabezados HTTP")
			return
		}

		trimmedLine := strings.TrimSpace(line) // Esto elimina \r y \n
		log.Printf("Worker: Leído Header Line: '%s' (Trimmed: '%s')", strings.ReplaceAll(line, "\n", "\\n"), trimmedLine) // <-- Log crucial

		if trimmedLine == "" { // Si es una línea vacía después de trim, es el fin de los encabezados
			log.Printf("Worker: Línea vacía de encabezado encontrada, terminando lectura de headers.") // Log para depuración
			break
		}

		parts := strings.SplitN(trimmedLine, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			headers[strings.ToLower(key)] = value // Guardar en minúsculas para fácil acceso
			log.Printf("Worker: Encabezado parseado: %s -> %s", key, value) // Log para depuración
		} else {
			log.Printf("Worker: No se pudo parsear línea de encabezado: '%s'", trimmedLine) // Log para depuración
		}
	}
	log.Printf("Worker: Mapa final de encabezados: %v", headers) // <-- Log crucial

	// Se incrementa el contador de solicitudes
	server.Metrics.Mu.Lock()
	server.Metrics.TotalRequests++
	server.Metrics.Mu.Unlock()

	log.Printf("Worker: Request ID: %d Method: %s Route: %s Params: %v", server.Metrics.TotalRequests, method, route, params)

	// Lógica para manejar POST /countchunk
	if method == "POST" && route == "/countchunk" {
		log.Printf("Worker: Received POST request for /countchunk. Delegando a handleCountChunkInWorker.")
		handleCountChunkInWorker(conn, headers, reader, server)
		return // Termina el manejo de la conexión aquí
	}

	// Lógica para manejar GET /calculatepi
	if method == "GET" && route == "/calculatepi" {
		log.Printf("Worker: Received GET request for /calculatepi with params: %v. Delegando a handleCalculatePiInWorker.", params)
		handleCalculatePiInWorker(conn, params, server)
		return // Termina el manejo de la conexión aquí
	}

	// Lógica para otros métodos y rutas (ej. GET /ping, GET /timestamp)
	if method != "GET" { // Ahora, si no es POST /countchunk o GET /calculatepi, solo permitimos GET
		utils.SendResponse(conn, "405 Method Not Allowed", "Método no permitido para esta ruta")
		return
	}

	// Manejo para GET (ping, timestamp, etc.)
	newRequest := Request{
		ID:           server.Metrics.TotalRequests, // Usa el contador actualizado
		Conn:         conn,
		Ruta:         route,
		Parametros:   params,
		TiempoInicio: time.Now(),
		Listo:        make(chan bool),
		Body:         "", // No hay cuerpo para solicitudes GET
	}

	log.Printf("Worker: Request ID: %d Ruta: %s (delegando a CommandPool)", newRequest.ID, newRequest.Ruta)

	if route == "/status" {
		serverStatus(conn, server) // Asume que serverStatus existe y envía la respuesta
	} else if route == "/ping" { // Manejar /ping directamente si no está en CommandPools
		utils.SendResponse(conn, "200 OK", "pong")
	} else if pool, exists := server.CommandPools[route]; exists {
		pool.RequestChan <- newRequest
		// Esperar a que la tarea esté lista si tu sistema de pools lo requiere
		<-newRequest.Listo // ¡Esto es crucial si quieres que el dispatcher reciba una respuesta!
	} else {
		utils.SendResponse(conn, "404 Not Found", "Ruta no encontrada")
	}
	// Si el canal Listo ya es manejado por el pool (cerrado o enviado), esto podría causar pánico.
	// Solo descomentar si el pool espera que TÚ cierres el canal.
	// newRequest.Listo <- true
}

// Genera el estado del servidor y retornar la respuesta en formato JSON
func serverStatus(conn net.Conn, s *Server) {
	s.Metrics.Mu.Lock()
	uptime := time.Since(s.Metrics.TiempoInicio).Truncate(time.Second).String()
	totalRequests := s.Metrics.TotalRequests
	s.Metrics.Mu.Unlock()
	totalWorkers := 0

	// Armamos una estructura por comando
	workersByCommand := make(map[string][]map[string]interface{})

	for ruta, pool := range s.CommandPools {
		var workers []map[string]interface{}
		for _, w := range pool.Workers {
			task := "ninguna"
			if w.ReqActual != nil {
				task = w.ReqActual.Ruta
			}
			workers = append(workers, map[string]interface{}{
				"pid":   w.ID,
				"task":  task,
				"state": w.Status,
			})
			totalWorkers += 1
		}
		workersByCommand[ruta] = workers
	}

	// Estado global
	data := map[string]interface{}{
		"uptime":            uptime,
		"main_pid":          s.ServerId,
		"total_connections": totalRequests,
		"total_workers":     totalWorkers,
		"workers":           workersByCommand,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		utils.SendResponse(conn, "500 Internal Server Error", "Error generando JSON")
		return
	}

	utils.SendJSON(conn, "200 OK", jsonData)
}

// handleCountChunkInWorker: Función para procesar el chunk de conteo de palabras
func handleCountChunkInWorker(conn net.Conn, headers map[string]string, reader *bufio.Reader, server *Server) {
	contentLengthStr, ok := headers["content-length"] // Los headers los parseamos a minúsculas
	var contentLength int
	if ok {
		var err error
		contentLength, err = strconv.Atoi(contentLengthStr)
		if err != nil {
			utils.SendResponse(conn, "400 Bad Request", "Content-Length inválido")
			log.Printf("Worker: Error: Content-Length inválido: %v", err)
			return
		}
	} else {
		log.Println("Worker Advertencia: No Content-Length header. Leyendo hasta EOF/timeout.")
		// Para POST, es muy recomendable tener Content-Length. Si no está presente,
		// leer hasta EOF puede ser problemático si la conexión no se cierra inmediatamente.
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
		// Condición de salida: si se ha leído todo el Content-Length o se llegó al EOF
		if err == io.EOF || (contentLength > 0 && bytesRead >= contentLength) {
			break
		}
		if err != nil {
			utils.SendResponse(conn, "500 Internal Server Error", "Error leyendo el cuerpo del archivo")
			log.Printf("Worker: Error leyendo el cuerpo del archivo: %v", err)
			return
		}
	}
	chunkContent := contentBuilder.String()
	log.Printf("Worker: Chunk recibido para conteo, tamaño: %d bytes. Contenido (primeros 100 chars): '%s'", len(chunkContent), chunkContent[:min(len(chunkContent), 100)]) // <-- Log crucial

	wordCount := countWords(chunkContent) // Asume que countWords existe y es correcto
	log.Printf("Worker: Conteo de palabras para chunk: %d", wordCount)

	utils.SendResponse(conn, "200 OK", fmt.Sprintf("%d", wordCount))
}

// Función auxiliar para contar palabras (ya la tienes en tu worker.go)
func countWords(text string) int {
	if len(strings.TrimSpace(text)) == 0 {
		return 0
	}
	words := strings.Fields(text)
	return len(words)
}

// handleCalculatePiInWorker: Función para calcular Pi usando Monte Carlo
func handleCalculatePiInWorker(conn net.Conn, params map[string]string, server *Server) {
	iterationsStr, ok := params["iterations"]
	if !ok {
		utils.SendResponse(conn, "400 Bad Request", "Parámetro 'iterations' requerido")
		return
	}

	iterations, err := strconv.Atoi(iterationsStr)
	if err != nil || iterations <= 0 {
		utils.SendResponse(conn, "400 Bad Request", "Parámetro 'iterations' debe ser un número entero positivo")
		return
	}

	log.Printf("Worker: Calculando Pi con %d iteraciones...", iterations)

	pointsInCircle := 0
	// `rand.Float64()` genera un float64 pseudoaleatorio en [0.0, 1.0)
	for i := 0; i < iterations; i++ {
		x := rand.Float64()
		y := rand.Float64()
		if (x*x + y*y) <= 1.0 { // Si el punto cae dentro del cuarto de círculo (radio 1)
			pointsInCircle++
		}
	}

	log.Printf("Worker: %d puntos dentro del círculo de %d iteraciones.", pointsInCircle, iterations)
	utils.SendResponse(conn, "200 OK", fmt.Sprintf("%d", pointsInCircle)) // Devuelve solo el conteo
}
