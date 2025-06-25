package main

import (
	"testing"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/assert"
)

// MockConn es un mock simple para simular una conexión net.Conn
type MockConn struct {
	mock.Mock
}

func (m *MockConn) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return len(p), args.Error(1)
}

func (m *MockConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Prueba simplificada para Write
func TestWriteMethodCalled(t *testing.T) {
	mockConn := new(MockConn)
	mockConn.On("Write", mock.Anything).Return(len("GET /ping HTTP/1.1\r\nHost: localhost:8080\r\n\r\n"), nil).Once()

	mockConn.Write([]byte("GET /ping HTTP/1.1\r\nHost: localhost:8080\r\n\r\n"))

	mockConn.AssertExpectations(t)
}

// Nueva prueba para Close
func TestCloseMethodCalled(t *testing.T) {
	mockConn := new(MockConn)

	// Definir lo que se espera: que Close sea llamado sin argumentos
	mockConn.On("Close").Return(nil).Once()

	// Llamamos al método Close en el mock
	err := mockConn.Close()

	// Verificamos que la llamada a Close se haya realizado y no hubo error
	mockConn.AssertExpectations(t)
	assert.NoError(t, err)
}
