// random_test.go
package handlers // Debe ser el mismo paquete que tu función Random

import (
	"bytes"
	"net"
	"strings"
	"testing"
	"time" // Necesario para rand.Seed
	"math/rand" // Necesario para rand.Seed
)

// Esto es una "conexión falsa" que podemos usar para las pruebas.
// Simplemente almacena lo que se "escribe" en ella.
type MockConn struct {
	Written bytes.Buffer // Aquí guardaremos la respuesta que se "envía"
}

// Estos métodos son necesarios para cumplir la interfaz net.Conn
func (m *MockConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *MockConn) Write(b []byte) (n int, err error)  { return m.Written.Write(b) } // Capturamos la escritura
func (m *MockConn) Close() error                       { return nil }
func (m *MockConn) LocalAddr() net.Addr                { return nil }
func (m *MockConn) RemoteAddr() net.Addr               { return nil }
func (m *MockConn) SetDeadline(t time.Time) error      { return nil }
func (m *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *MockConn) SetWriteDeadline(t time.Time) error { return nil }

// Estas variables globales guardarán la ÚLTIMA respuesta enviada por nuestra función simulada.
var (
	testStatus string
	testBody   string
)

// Esta es nuestra función `SendResponse` simulada.
// En lugar de enviar la respuesta por red, simplemente la guarda en nuestras variables.
func mockSendResponse(conn net.Conn, status string, body string) {
	testStatus = status
	testBody = body
	// Puedes opcionalmente escribirlo a la MockConn para ver el resultado completo
	// _, _ = conn.Write([]byte(fmt.Sprintf("HTTP/1.1 %s\r\n\r\n%s", status, body)))
}

// --- TUS PRUEBAS ---

// Prueba un caso exitoso con entradas válidas
func TestRandom_ValidInput(t *testing.T) {
	// IMPORTANTE: Si quieres que los números aleatorios sean siempre los mismos en el test, usa una semilla fija.
	// Si solo te interesa el formato y no los valores exactos, puedes usar time.Now().UnixNano().
	rand.Seed(42) // Para resultados predecibles en los números aleatorios

	mockConn := &MockConn{} // Crea una conexión falsa
	testStatus = ""         // Resetea las variables de la respuesta simulada
	testBody = ""

	// Llama a tu función Random, pero ahora le pasas nuestro 'mockSendResponse'
	Random(mockConn, "1", "10", "5", mockSendResponse)

	// Ahora verificamos lo que `mockSendResponse` guardó
	if testStatus != "200 OK" {
		t.Errorf("Esperado status '200 OK', obtenido '%s'", testStatus)
	}

	expectedPrefix := "Se generaron 5 numeros aleatorios entre 1 y 10:"
	if !strings.HasPrefix(testBody, expectedPrefix) {
		t.Errorf("El cuerpo de la respuesta no empieza como se esperaba.\nEsperado: %s...\nObtenido: %s", expectedPrefix, testBody)
	}

	// Puedes añadir más verificaciones si quieres:
	// Por ejemplo, que el cuerpo contenga "Indice\tNumero" y "------\t------"
	if !strings.Contains(testBody, "Indice\tNumero") || !strings.Contains(testBody, "------\t------") {
		t.Errorf("El formato de la tabla no es el esperado:\n%s", testBody)
	}

	// Puedes verificar que hay 5 líneas de números aleatorios + las líneas de encabezado
	lines := strings.Split(strings.TrimSpace(testBody), "\n")
	if len(lines) != 4+5 { // 4 líneas de encabezado (intro, título, Indice/Numero, ------) + 5 líneas de números
		t.Errorf("Se esperaban %d líneas de salida, se obtuvieron %d. Salida:\n%s", 4+5, len(lines), testBody)
	}
}

// Prueba el caso de una cantidad no numérica
func TestRandom_InvalidCantidad(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	Random(mockConn, "3", "10", "abc", mockSendResponse)

	if testStatus != "400 Bad Request" {
		t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
	}
	if testBody != "Cantidad debe ser un numero valido\n" {
		t.Errorf("Esperado body 'Cantidad debe ser un numero valido\\n', obtenido '%s'", testBody)
	}
}

// Prueba el caso de una cantidad negativa
func TestRandom_NegativeCantidad(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	Random(mockConn, "1", "10", "-5", mockSendResponse)

	if testStatus != "400 Bad Request" {
		t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
	}
	if testBody != "La cantidad debe ser un numero entero positivo\n" {
		t.Errorf("Esperado body 'La cantidad debe ser un numero entero positivo\\n', obtenido '%s'", testBody)
	}
}

// Agrega más funciones Test para los otros casos de error (min/max inválidos, min >= max)
// siguiendo el mismo patrón.
// Ejemplo:
func TestRandom_MinGreaterThanMax(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	Random(mockConn, "10", "5", "3", mockSendResponse)

	if testStatus != "400 Bad Request" {
		t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
	}
	if testBody != "El minimo debe ser menor al maximo\n" {
		t.Errorf("Esperado body 'El minimo debe ser menor al maximo\\n', obtenido '%s'", testBody)
	}
}