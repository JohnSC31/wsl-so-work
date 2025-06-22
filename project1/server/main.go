package main

import (
	"bufio"
	"encoding/json"
	"http-servidor/utils"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
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
func handleConnection(conn net.Conn, server *Server) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	request, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error leyendo: %v", err)
		return
	}

	method, fullpath := utils.ParseRequestLine(request)
	log.Printf("Nueva solicitud: %s %s", method, fullpath)

	if method != "GET" {
		utils.SendResponse(conn, "405 Method Not Allowed", "Solo se permite GET")
		return
	}

	route, params := parseFullPath(fullpath)

	if params["request_id"] == "" || params["callback_url"] == "" {
		log.Printf("ERROR: Parámetros esenciales faltantes. request_id: '%s', callback_url: '%s'",
			params["request_id"], params["callback_url"])
		utils.SendResponse(conn, "400 Bad Request", "Missing required parameters")
		return
	}

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

func parseFullPath(fullPath string) (string, map[string]string) {
	// Separar path de query parameters
	parts := strings.SplitN(fullPath, "?", 2)
	route := parts[0]
	params := make(map[string]string)

	if len(parts) > 1 {
		// Parsear cada par key=value
		pairs := strings.Split(parts[1], "&")
		for _, pair := range pairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) == 2 {
				// Decodificar valores URL-encoded
				decodedValue, err := url.QueryUnescape(kv[1])
				if err != nil {
					decodedValue = kv[1] // Usar original si falla decoding
				}
				params[kv[0]] = decodedValue
			}
		}
	}

	log.Printf("Parsed route: %s, params: %v", route, params)
	return route, params
}
