// loadtest_test.go
package handlers

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestLoadtest_ValidInput_NoSleep prueba un caso exitoso con sleep=0
func TestLoadtest_ValidInput_NoSleep(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	tasks := "5"
	sleep := "0" // No sleep for quicker testing

	Loadtest(mockConn, tasks, sleep, mockSendResponse)

	if testStatus != "200 OK" {
		t.Errorf("Esperado status '200 OK', obtenido '%s'", testStatus)
	}

	// Verificar que el cuerpo contiene la información correcta
	expectedTasks := 5
	expectedSleep := 0
	if !strings.Contains(testBody, "Se ejecutaron "+strconv.Itoa(expectedTasks)+" tareas concurrentes") {
		t.Errorf("El cuerpo de la respuesta no menciona el número de tareas esperado.\nObtenido:\n%s", testBody)
	}
	if !strings.Contains(testBody, "con "+strconv.Itoa(expectedSleep)+" segundos de espera cada una.") {
		t.Errorf("El cuerpo de la respuesta no menciona el tiempo de espera esperado.\nObtenido:\n%s", testBody)
	}

	// La duración debe ser muy cercana a cero si sleep es 0 y las tareas son mínimas.
	// Extraer la duración y verificar que sea pequeña.
	lines := strings.Split(testBody, "\n")
	if len(lines) < 4 {
		t.Fatalf("El cuerpo de la respuesta tiene formato inesperado: %s", testBody)
	}
	durationLine := lines[len(lines)-1] // La última línea es la duración
	if !strings.HasPrefix(durationLine, "Duracion total:") {
		t.Errorf("La última línea no es la duración total: %s", durationLine)
	}
	// Extraer el valor numérico de la duración
	durationStr := strings.TrimSuffix(strings.TrimPrefix(durationLine, "Duracion total: "), " segundos")
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		t.Fatalf("No se pudo parsear la duración de la respuesta: %v", err)
	}

	// Con sleep=0, la duración total debe ser muy pequeña, casi instantánea.
	// Le damos un pequeño margen.
	if duration > 0.01 { // Permitimos hasta 10ms por overhead de goroutines y scheduling
		t.Errorf("La duración esperada era cercana a 0, obtenida %.2f segundos", duration)
	}
}

// TestLoadtest_ValidInput_WithSleep prueba un caso exitoso con sleep
func TestLoadtest_ValidInput_WithSleep(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	tasks := "3"
	sleep := "1" // 1 segundo de espera por tarea

	startTime := time.Now()
	Loadtest(mockConn, tasks, sleep, mockSendResponse)
	endTime := time.Now()

	if testStatus != "200 OK" {
		t.Errorf("Esperado status '200 OK', obtenido '%s'", testStatus)
	}

	// La duración total debería ser aproximadamente el valor de 'sleep'
	// porque las tareas se ejecutan concurrentemente.
	actualDuration := endTime.Sub(startTime).Seconds()
	expectedDurationMin := float64(1) // Mínimo 1 segundo (el sleep de una tarea)
	expectedDurationMax := float64(1.5) // Un poco de margen para el scheduling de goroutines

	if actualDuration < expectedDurationMin || actualDuration > expectedDurationMax {
		t.Errorf("Duración inesperada para %s tareas y %s seg de sleep. Esperado entre %.2f y %.2f segundos, obtenido %.2f segundos",
			tasks, sleep, expectedDurationMin, expectedDurationMax, actualDuration)
	}

	// Verificar el contenido del body de la respuesta
	if !strings.Contains(testBody, "Se ejecutaron 3 tareas concurrentes con 1 segundos de espera cada una.") {
		t.Errorf("El cuerpo de la respuesta no contiene el resumen esperado.\nObtenido:\n%s", testBody)
	}
}

// TestLoadtest_InvalidTasks prueba casos de 'tasks' inválidos
func TestLoadtest_InvalidTasks(t *testing.T) {
	tests := []struct {
		name     string
		tasks    string
		expected string
	}{
		{"non-numeric", "abc", "El parametro 'tasks' debe ser un número valido mayor que 0\n"},
		{"zero", "0", "El parametro 'tasks' debe ser un número valido mayor que 0\n"},
		{"negative", "-5", "El parametro 'tasks' debe ser un número valido mayor que 0\n"},
		{"missing", "", "El parametro 'tasks' debe ser un número valido mayor que 0\n"}, // Assuming "" converts to 0 or errors out
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockConn{}
			testStatus = ""
			testBody = ""

			Loadtest(mockConn, tt.tasks, "1", mockSendResponse) // sleep can be valid

			if testStatus != "400 Bad Request" {
				t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
			}
			if testBody != tt.expected {
				t.Errorf("Esperado body '%s', obtenido '%s'", tt.expected, testBody)
			}
		})
	}
}

// TestLoadtest_InvalidSleep prueba casos de 'sleep' inválidos
func TestLoadtest_InvalidSleep(t *testing.T) {
	tests := []struct {
		name     string
		sleep    string
		expected string
	}{
		{"non-numeric", "xyz", "El parametro 'sleep' debe ser un numero valido\n"},
		{"negative", "-1", "El parametro 'sleep' debe ser un numero valido\n"}, // Sleep can be 0, but not negative
		{"missing", "", "El parametro 'sleep' debe ser un numero valido\n"},    // Assuming "" converts to 0 or errors out
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockConn{}
			testStatus = ""
			testBody = ""

			Loadtest(mockConn, "1", tt.sleep, mockSendResponse) // tasks can be valid

			if testStatus != "400 Bad Request" {
				t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
			}
			if testBody != tt.expected {
				t.Errorf("Esperado body '%s', obtenido '%s'", tt.expected, testBody)
			}
		})
	}
}