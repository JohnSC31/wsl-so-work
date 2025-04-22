package main

// Pruebas unitarias
// Status
// 

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
	"encoding/json"

	// "strings"

	_ "http-servidor/handlers"
	"http-servidor/utils"
)

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

type Server struct {
	ServerId int
	FastPool *WorkerPool
	SlowPool *WorkerPool
	Metrics  *Metricas
	listener net.Listener  // Socket subyacente
	doneChan chan struct{} // Para shutdown
}

type Metricas struct {
	Mu            sync.Mutex
	TiempoInicio  time.Time
	TotalRequests int
	ActWorkers    int
}

const PORT = ":8080"

func NewServer() *Server {
	return &Server{
		ServerId: 1,
		FastPool: NewWorkerPool(3),  // 3 workers rápidos
		SlowPool: NewWorkerPool(10), // 10 workers lentos
		Metrics: &Metricas{TiempoInicio: time.Now(),
			TotalRequests: 0,
			ActWorkers:    0,
		},
		doneChan: make(chan struct{}),
	}
}

func printAllWorkers(server *Server){

	for _, fastworker := range server.FastPool.Workers {
		printWorker(fastworker)
	}

	for _, slowWorker := range server.SlowPool.Workers {
		printWorker(slowWorker)
		// fmt.Printf("PID: %d | Task: %s | Estado: %s\n", worker.PID, worker.Task, worker.State)
	}
}

func printWorker(w *Worker){
	// fmt.Printf("Worker %d estado %d comando %d", w.ID, w.Status, w.ReqActual.Ruta)
	fmt.Printf("PID: %d | Task: %s | Estado: %s\n", w.ID, w.ReqActual.Ruta, w.Status)
}

func main() {

	Server := NewServer()
	Server.FastPool.Start()
	Server.SlowPool.Start()

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

func handleConnection(conn net.Conn, server *Server) {
	//defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Error leyendo: %v", err)
		return
	}

	request := string(buffer[:n])
	method, path := utils.ParseRequestLine(request)

	if method != "GET" {
		utils.SendResponse(conn, "405 Method Not Allowed", "Solo se permite GET")
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

	switch route {

	case "/help", "/timestamp", "/reverse", "/toupper", "/hash", "/random":
		server.FastPool.RequestChan <- newRequest

	case "/fibonacci", "/simulate", "/sleep", "/loadtest", "/createfile", "/deletefile":
		server.SlowPool.RequestChan <- newRequest

	case "/status":
		// el estado del servidor
		data := map[string]interface{}{
			"uptime":              server.Metrics.TiempoInicio,
			"main_pid":            server.ServerId,
			"total_connections":   server.Metrics.TotalRequests,
			"workers":             printAllWorkers(server),
		}
	
		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			utils.SendResponse(conn, "500 Internal Server Error", "Error generando JSON")
			return
		}
	
		utils.SendJSON(conn, "200 OK", jsonData)

	default:
		utils.SendResponse(conn, "404 Not Found", "Ruta no encontrada")
	}

	newRequest.Listo <- true
}
