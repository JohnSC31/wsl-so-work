package main

import (
	// "fmt"
	"log"
	"net"
	"sync"
	"time"

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
		FastPool: NewWorkerPool(3),  // 3 workers rápidos
		SlowPool: NewWorkerPool(10), // 10 workers lentos
		Metrics: &Metricas{TiempoInicio: time.Now(),
			TotalRequests: 0,
			ActWorkers:    0,
		},
		doneChan: make(chan struct{}),
	}
}

func main() {

	Server := NewServer()
	Server.FastPool.Start()
	Server.SlowPool.Start()
	fast := Server.FastPool
	slow := Server.SlowPool
	metricas := Server.Metrics

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
		go handleConnection(conn, metricas, fast, slow)
	}
}

func handleConnection(conn net.Conn, metricas *Metricas, fast *WorkerPool, slow *WorkerPool) {
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
	metricas.Mu.Lock()
	metricas.TotalRequests++
	metricas.Mu.Unlock()

	newRequest := Request{
		ID:           metricas.TotalRequests,
		Conn:         conn,
		Ruta:         route,
		Parametros:   params,
		TiempoInicio: time.Now(),
		Listo:        make(chan bool),
	}

	print("Request ID: ", newRequest.ID, " Ruta: ", newRequest.Ruta, "\n")
	switch route {
	case "/help", "/timestamp", "/reverse", "/toupper", "/hash", "/random":
		fast.RequestChan <- newRequest
	case "/fibonacci", "/simulate", "/sleep", "/loadtest", "/createfile", "/deletefile":
		slow.RequestChan <- newRequest
	default:
		utils.SendResponse(conn, "404 Not Found", "Ruta no encontrada")
	}

	newRequest.Listo <- true
}
