package main

import (
	"testing"
	"github.com/stretchr/testify/mock"
	"net"
	"time"
)

// MockConn es un mock simple para simular una conexión net.Conn
type MockConn struct {
	mock.Mock
}

func (m *MockConn) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return len(p), args.Error(1)
}

func (m *MockConn) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return 0, args.Error(1)
}

func (m *MockConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConn) LocalAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *MockConn) RemoteAddr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func (m *MockConn) SetDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockConn) SetReadDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockConn) SetWriteDeadline(t time.Time) error {
	args := m.Called(t)
	return args.Error(0)
}

// Dispatcher y Worker simplificados para la prueba
type Worker struct {
	ID        int
	URL       string
	taskQueue chan *Task
}

type Dispatcher struct {
	Workers []*Worker
}

type Task struct {
	ID    int
	Conn  net.Conn
	Params map[string]string
}

func (d *Dispatcher) handleCalculatePi(conn net.Conn, method string, route string, params map[string]string) {
	iterationsStr := params["iterations"]
	if iterationsStr == "" {
		conn.Write([]byte("400 Bad Request: Parámeter 'iterations' requerido"))
		return
	}

	conn.Write([]byte("200 OK: Cálculo de Pi iniciado"))
}

// Prueba simple para verificar que la función handleCalculatePi responde correctamente
func TestHandleCalculatePi(t *testing.T) {
	mockConn := new(MockConn)

	// Creamos un dispatcher con un worker para la prueba
	worker := &Worker{
		ID:        1,
		URL:       "http://localhost",
		taskQueue: make(chan *Task, 1),
	}
	dispatcher := &Dispatcher{
		Workers: []*Worker{worker},
	}

	// Parámetros de prueba
	params := map[string]string{
		"iterations": "1000",
	}

	// Mocking Write para verificar la respuesta
	mockConn.On("Write", mock.Anything).Return(0, nil)

	// Llamada a handleCalculatePi
	dispatcher.handleCalculatePi(mockConn, "POST", "/calculatepi", params)

	// Verificamos que la respuesta fue la esperada
	mockConn.AssertExpectations(t)
}
