// timestamp_test.go
package handlers

import (
	"encoding/json"
	"testing"
	"time"
)

// TestTimestamp_Success prueba que la función devuelve un timestamp válido en formato JSON.
func TestTimestamp_Success(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	// Llamar a la función Timestamp
	Timestamp(mockConn, mockSendResponse)

	// Verificar el estado de la respuesta
	if testStatus != "200 OK" {
		t.Errorf("Esperado status '200 OK', obtenido '%s'", testStatus)
	}

	// Verificar el cuerpo de la respuesta
	// Debería ser un JSON con un campo "timestamp"
	var response map[string]string
	err := json.Unmarshal([]byte(testBody), &response)
	if err != nil {
		t.Fatalf("No se pudo parsear el cuerpo JSON de la respuesta: %v, cuerpo: %s", err, testBody)
	}

	timestampStr, ok := response["timestamp"]
	if !ok {
		t.Errorf("El cuerpo JSON no contiene el campo 'timestamp'. Cuerpo: %s", testBody)
	}

	// Intentar parsear el timestamp para asegurar que tiene el formato correcto (RFC3339)
	parsedTime, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		t.Errorf("El timestamp '%s' no está en formato RFC3339 válido: %v", timestampStr, err)
	}

	// Verificar que el timestamp es razonablemente reciente (ej. dentro de unos segundos de la ejecución del test)
	// Como la ejecución es muy rápida, podemos esperar que sea muy cercano al tiempo actual.
	timeNow := time.Now()
	// Diferencia entre el tiempo parseado y el tiempo actual del test
	diff := timeNow.Sub(parsedTime)

	// Aceptar una diferencia muy pequeña (e.g., 100ms) para dar margen al scheduler del sistema.
	// La diferencia debe ser positiva (el timestamp generado es un poco antes que time.Now() de la verificación)
	// o muy ligeramente negativa si el sistema de scheduling es inusual.
	if diff < -500*time.Millisecond || diff > 500*time.Millisecond {
		t.Errorf("La diferencia entre el tiempo generado y el tiempo actual es demasiado grande. Generado: %s, Actual: %s, Diferencia: %s", parsedTime.Format(time.RFC3339), timeNow.Format(time.RFC3339), diff)
	}
}