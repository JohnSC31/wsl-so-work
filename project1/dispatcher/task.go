// task.go (en módulo dispatcher)
package main

import (
	"net"
	"time"
)

type TaskStatus int

const (
	TaskPending TaskStatus = iota
	TaskProcessing
	TaskCompleted
	TaskFailed
)

type Task struct {
	ID          int      // UUID sería mejor para distribución
	Conn        net.Conn // Conexión cliente original
	Request     *Request // Datos de la solicitud
	Response    []byte   // Respuesta del worker
	Status      TaskStatus
	AssignedTo  *Worker // Worker asignado
	CreatedAt   time.Time
	CompletedAt time.Time
	RetryCount  int // Para reintentos
	Content 	string
}

type Request struct {
	Method string
	Path   string
	Params map[string]string
	Done   chan bool
}
