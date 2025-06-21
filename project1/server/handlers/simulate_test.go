// simulate_test.go
package handlers

import (
	"strings"
	"testing"
	"time"
)

// TestSimulate_ValidInput prueba un caso exitoso con entrada válida
func TestSimulate_ValidInput(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	seconds := "1" // Simulación de 1 segundo
	taskName := "MiTareaDeSimulacion"

	// Capturar el tiempo antes de llamar a Simulate
	startTime := time.Now()

	Simulate(mockConn, seconds, taskName, mockSendResponse)

	// Capturar el tiempo después de que Simulate ha terminado
	endTime := time.Now()

	if testStatus != "200 OK" {
		t.Errorf("Esperado status '200 OK', obtenido '%s'", testStatus)
	}

	// Verificar que la duración real de la ejecución es aproximadamente la esperada
	// Damos un margen para el overhead del test y el scheduler
	duration := endTime.Sub(startTime)
	expectedDuration := time.Duration(1) * time.Second
	// Usamos un margen de +/- 100ms para la prueba de tiempo real
	if duration < expectedDuration-100*time.Millisecond || duration > expectedDuration+100*time.Millisecond {
		t.Errorf("Duración de la simulación inesperada. Esperado ~%s, obtenido %s", expectedDuration, duration)
	}

	// Verificar el contenido del cuerpo de la respuesta
	if !strings.Contains(testBody, "Simulacion completada\n\n") {
		t.Errorf("El cuerpo de la respuesta no contiene el mensaje de completado.\nObtenido:\n%s", testBody)
	}
	if !strings.Contains(testBody, "Nombre de la tarea: "+taskName+"\n") {
		t.Errorf("El cuerpo de la respuesta no contiene el nombre de la tarea.\nObtenido:\n%s", testBody)
	}
	if !strings.Contains(testBody, "Duracion: "+seconds+" segundos\n") {
		t.Errorf("El cuerpo de la respuesta no contiene la duración correcta.\nObtenido:\n%s", testBody)
	}

	// Verificar el formato de la hora de finalización (no el valor exacto ya que es dinámico)
	if !strings.Contains(testBody, "Hora de finalizacion:") {
		t.Errorf("El cuerpo de la respuesta no contiene la hora de finalización.\nObtenido:\n%s", testBody)
	}
	// Podrías intentar parsear la fecha y verificar que es reciente, pero es más complejo.
	// Por ahora, con el string.Contains es suficiente para el formato.
}

// TestSimulate_InvalidSeconds prueba casos de 'seconds' inválidos
func TestSimulate_InvalidSeconds(t *testing.T) {
	tests := []struct {
		name     string
		seconds  string
		expected string
	}{
		{"non-numeric", "abc", "Seconds debe ser un numero valido, entero y positivo\n"},
		{"zero", "0", "Seconds debe ser un numero valido, entero y positivo\n"},
		{"negative", "-5", "Seconds debe ser un numero valido, entero y positivo\n"},
		{"missing", "", "Seconds debe ser un numero valido, entero y positivo\n"}, // Asumiendo que "" resulta en error o 0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockConn{}
			testStatus = ""
			testBody = ""

			Simulate(mockConn, tt.seconds, "AnyTask", mockSendResponse)

			if testStatus != "400 Bad Request" {
				t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
			}
			if testBody != tt.expected {
				t.Errorf("Esperado body '%s', obtenido '%s'", tt.expected, testBody)
			}
		})
	}
}

// TestSimulate_EmptyTaskName prueba el caso donde el nombre de la tarea está vacío
// Aunque la función Simulate no tiene validación explícita para 'nombre' vacío,
// este test asegura que no cause un pánico y que el resultado incluya el nombre vacío.
func TestSimulate_EmptyTaskName(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	seconds := "1"
	taskName := "" // Nombre de tarea vacío

	Simulate(mockConn, seconds, taskName, mockSendResponse)

	if testStatus != "200 OK" {
		t.Errorf("Esperado status '200 OK', obtenido '%s'", testStatus)
	}

	// Verificar que el cuerpo de la respuesta incluye el nombre vacío
	if !strings.Contains(testBody, "Nombre de la tarea: \n") {
		t.Errorf("El cuerpo de la respuesta no contiene el nombre de la tarea vacío.\nObtenido:\n%s", testBody)
	}
}

// TestSimulate_LongDuration prueba con una duración más larga para asegurar que funcione
func TestSimulate_LongDuration(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	seconds := "2" // Simulación de 2 segundos
	taskName := "LongRunningTask"

	startTime := time.Now()
	Simulate(mockConn, seconds, taskName, mockSendResponse)
	endTime := time.Now()

	if testStatus != "200 OK" {
		t.Errorf("Esperado status '200 OK', obtenido '%s'", testStatus)
	}

	duration := endTime.Sub(startTime)
	expectedDuration := time.Duration(2) * time.Second
	if duration < expectedDuration-100*time.Millisecond || duration > expectedDuration+100*time.Millisecond {
		t.Errorf("Duración de la simulación inesperada. Esperado ~%s, obtenido %s", expectedDuration, duration)
	}

	if !strings.Contains(testBody, "Nombre de la tarea: "+taskName+"\n") {
		t.Errorf("El cuerpo de la respuesta no contiene el nombre de la tarea.\nObtenido:\n%s", testBody)
	}
	if !strings.Contains(testBody, "Duracion: "+seconds+" segundos\n") {
		t.Errorf("El cuerpo de la respuesta no contiene la duración correcta.\nObtenido:\n%s", testBody)
	}
}