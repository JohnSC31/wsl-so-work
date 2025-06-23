// hash_test.go
package handlers

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

// TestHash_ValidInput prueba un caso exitoso con una entrada de texto válida
func TestHash_ValidInput(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	inputText := "hello world"
	// Calcular el hash SHA-256 esperado de "hello world"
	hasher := sha256.New()
	hasher.Write([]byte(inputText))
	expectedHash := fmt.Sprintf("%x", hasher.Sum(nil))

	Hash(mockConn, inputText, mockSendResponse)

	if testStatus != "200 OK" {
		t.Errorf("Esperado status '200 OK', obtenido '%s'", testStatus)
	}

	expectedBody := "El hash SHA-256 del texto es:\n\n" + expectedHash + "\n"
	if testBody != expectedBody {
		t.Errorf("Cuerpo de la respuesta inesperado.\nEsperado:\n%sObtenido:\n%s", expectedBody, testBody)
	}
}

// TestHash_EmptyInput prueba el caso de texto vacío
func TestHash_EmptyInput(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	inputText := ""
	Hash(mockConn, inputText, mockSendResponse)

	if testStatus != "400 Bad Request" {
		t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
	}
	if testBody != "Texto no puede ser vacio\n" {
		t.Errorf("Esperado body 'Texto no puede ser vacio\\n', obtenido '%s'", testBody)
	}
}

// TestHash_WhitespaceInput prueba el caso de texto con solo espacios en blanco
func TestHash_WhitespaceInput(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	inputText := "   \t\n " // Espacios, tabulaciones, nueva línea
	Hash(mockConn, inputText, mockSendResponse)

	if testStatus != "400 Bad Request" {
		t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
	}
	if testBody != "Texto no puede ser vacio\n" {
		t.Errorf("Esperado body 'Texto no puede ser vacio\\n', obtenido '%s'", testBody)
	}
}

// TestHash_DifferentInput prueba con un texto diferente para asegurar que el hash es correcto
func TestHash_DifferentInput(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	inputText := "GoLang rocks!"
	// Calcular el hash SHA-256 esperado de "GoLang rocks!"
	hasher := sha256.New()
	hasher.Write([]byte(inputText))
	expectedHash := fmt.Sprintf("%x", hasher.Sum(nil))

	Hash(mockConn, inputText, mockSendResponse)

	if testStatus != "200 OK" {
		t.Errorf("Esperado status '200 OK', obtenido '%s'", testStatus)
	}

	expectedBody := "El hash SHA-256 del texto es:\n\n" + expectedHash + "\n"
	if testBody != expectedBody {
		t.Errorf("Cuerpo de la respuesta inesperado.\nEsperado:\n%sObtenido:\n%s", expectedBody, testBody)
	}
}