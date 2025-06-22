package main

import (
	"log"
	"net"
	"time"
	// Replace with the actual module path if different
)

func main() {
	dispatcher := newDispatcher()

	// Inicia health checks periódicos
	go func() {
		ticker := time.NewTicker(HealthCheckInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				dispatcher.HealthCheck()
			}
		}
	}()

	// Inicia el servidor HTTP del dispatcher
	ln, err := net.Listen("tcp", DispatcherPort)
	if err != nil {
		log.Fatalf("Error al iniciar dispatcher: %v", err)
	}
	defer ln.Close()

	log.Printf("Dispatcher escuchando en %s", DispatcherPort)

	// Bucle principal para aceptar conexiones
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Error aceptando conexión: %v", err)
			continue
		}

		go dispatcher.HandleConnection(conn)
	}
}
