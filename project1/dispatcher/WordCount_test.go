package main

import (
	"testing"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/assert"
)

// Mock para el Dispatcher
type MockDispatcher struct {
	mock.Mock
}

func (m *MockDispatcher) HandleWordCount(input string) map[string]string {
	args := m.Called(input)
	return args.Get(0).(map[string]string)
}

func TestHandleWordCount(t *testing.T) {
	// Crear una instancia del mock
	mockDispatcher := new(MockDispatcher)

	// Prueba 1: Configurar el mock para que devuelva un resultado esperado con el input "hello world"
	mockDispatcher.On("HandleWordCount", "hello world").Return(map[string]string{"Content-Length": "42"})

	// Llamar al método real del mock
	result := mockDispatcher.HandleWordCount("hello world")

	// Aserciones: Comprobar que el resultado es el esperado
	assert.Equal(t, map[string]string{"Content-Length": "42"}, result)

	// Verificar que el mock fue llamado correctamente
	mockDispatcher.AssertExpectations(t)

	// Prueba 2: Verificar que se maneje un input diferente y devuelva el valor esperado
	mockDispatcher.On("HandleWordCount", "goodbye world").Return(map[string]string{"Content-Length": "24"})

	// Llamada con el nuevo input
	result2 := mockDispatcher.HandleWordCount("goodbye world")

	// Aserciones para el segundo caso
	assert.Equal(t, map[string]string{"Content-Length": "24"}, result2)

	// Verificar que el mock fue llamado correctamente
	mockDispatcher.AssertExpectations(t)

	// Prueba 3: Comprobar que si se pasa un input inesperado, el mock devuelve el valor adecuado
	mockDispatcher.On("HandleWordCount", "unknown input").Return(map[string]string{"Content-Length": "0"})

	// Llamada con input no esperado
	result3 := mockDispatcher.HandleWordCount("unknown input")

	// Aserción para el input inesperado
	assert.Equal(t, map[string]string{"Content-Length": "0"}, result3)

	// Verificar que el mock fue llamado correctamente
	mockDispatcher.AssertExpectations(t)

	// Prueba 4: Verificar que el método maneja una entrada vacía correctamente
	mockDispatcher.On("HandleWordCount", "").Return(map[string]string{"Content-Length": "0"})

	// Llamada con entrada vacía
	result4 := mockDispatcher.HandleWordCount("")

	// Aserción para la entrada vacía
	assert.Equal(t, map[string]string{"Content-Length": "0"}, result4)

	// Verificar que el mock fue llamado correctamente
	mockDispatcher.AssertExpectations(t)
}
