// sleep_test.go
package handlers

import (
	"testing"
	"time"
)


// TestSleep_ValidInput prueba un caso exitoso con entrada válida
func TestSleep_ValidInput(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	seconds := "1" // Dormir 1 segundo para la prueba

	// Capturar el tiempo antes de llamar a Sleep
	startTime := time.Now()

	// Ahora, pasamos mockSendResponse directamente a la función Sleep
	Sleep(mockConn, seconds, mockSendResponse)

	// Capturar el tiempo después de que Sleep ha terminado
	endTime := time.Now()

	if testStatus != "200 OK" {
		t.Errorf("Esperado status '200 OK', obtenido '%s'", testStatus)
	}

	// Verificar que la duración real de la ejecución es aproximadamente la esperada
	duration := endTime.Sub(startTime)
	expectedDuration := time.Duration(1) * time.Second
	// Usamos un margen de +/- 100ms para la prueba de tiempo real
	if duration < expectedDuration-100*time.Millisecond || duration > expectedDuration+100*time.Millisecond {
		t.Errorf("Duración del sleep inesperada. Esperado ~%s, obtenido %s", expectedDuration, duration)
	}

	// Verificar el contenido del cuerpo de la respuesta
	expectedBody := "Sleep realizado durante " + seconds + " segundos\n"
	if testBody != expectedBody {
		t.Errorf("Cuerpo de la respuesta inesperado.\nEsperado:\n%sObtenido:\n%s", expectedBody, testBody)
	}
}

// TestSleep_InvalidSeconds prueba casos de 'seconds' inválidos
func TestSleep_InvalidSeconds(t *testing.T) {
	tests := []struct {
		name     string
		seconds  string
		expected string
	}{
		{"non-numeric", "abc", "Seconds debe ser un numero valido, entero y postivo\n"},
		{"zero", "0", "Seconds debe ser un numero valido, entero y postivo\n"},
		{"negative", "-5", "Seconds debe ser un numero valido, entero y postivo\n"},
		{"missing", "", "Seconds debe ser un numero valido, entero y postivo\n"}, // Asumiendo que "" resulta en error o 0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockConn{}
			testStatus = ""
			testBody = ""

			// Ahora, pasamos mockSendResponse directamente a la función Sleep
			Sleep(mockConn, tt.seconds, mockSendResponse)

			if testStatus != "400 Bad Request" {
				t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
			}
			if testBody != tt.expected {
				t.Errorf("Esperado body '%s', obtenido '%s'", tt.expected, testBody)
			}
		})
	}
}

// TestSleep_LongDuration prueba con una duración más larga para asegurar que funcione
func TestSleep_LongDuration(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	seconds := "2" // Dormir 2 segundos

	startTime := time.Now()
	// Ahora, pasamos mockSendResponse directamente a la función Sleep
	Sleep(mockConn, seconds, mockSendResponse)
	endTime := time.Now()

	if testStatus != "200 OK" {
		t.Errorf("Esperado status '200 OK', obtenido '%s'", testStatus)
	}

	duration := endTime.Sub(startTime)
	expectedDuration := time.Duration(2) * time.Second
	if duration < expectedDuration-100*time.Millisecond || duration > expectedDuration+100*time.Millisecond {
		t.Errorf("Duración del sleep inesperada. Esperado ~%s, obtenido %s", expectedDuration, duration)
	}

	expectedBody := "Sleep realizado durante " + seconds + " segundos\n"
	if testBody != expectedBody {
		t.Errorf("Cuerpo de la respuesta inesperado.\nEsperado:\n%sObtenido:\n%s", expectedBody, testBody)
	}
}