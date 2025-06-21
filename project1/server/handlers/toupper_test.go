// toupper_test.go
package handlers

import (
	"testing"
)

// TestToUpper_ValidInput prueba casos exitosos con entradas de texto válidas.
func TestToUpper_ValidInput(t *testing.T) {
	tests := []struct {
		name     string
		inputText string
		expected string
	}{
		{"English Word", "hello", "HELLO\n"},
		{"Sentence", "Go is fun", "GO IS FUN\n"},
		{"Mixed Case", "HeLlO WoRlD", "HELLO WORLD\n"},
		{"Numbers and Symbols", "123!@#abc", "123!@#ABC\n"},
		{"Unicode Characters (non-alphabetic)", "你好世界", "你好世界\n"}, // ToUpper only affects letters
		{"String with leading/trailing spaces", "  text  ", "  TEXT  \n"}, // spaces are preserved
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockConn{}
			testStatus = ""
			testBody = ""

			params := map[string]string{"text": tt.inputText}
			ToUpper(mockConn, params, mockSendResponse)

			if testStatus != "200 OK" {
				t.Errorf("Esperado status '200 OK', obtenido '%s'", testStatus)
			}
			if testBody != tt.expected {
				t.Errorf("Esperado body '%s', obtenido '%s' para input '%s'", tt.expected, testBody, tt.inputText)
			}
		})
	}
}

// TestToUpper_MissingParam prueba el caso de falta del parámetro 'text'.
func TestToUpper_MissingParam(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	params := map[string]string{} // Parámetros sin 'text'
	ToUpper(mockConn, params, mockSendResponse)

	if testStatus != "400 Bad Request" {
		t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
	}
	if testBody != "Falta el parámetro 'text'\n" {
		t.Errorf("Esperado body 'Falta el parámetro 'text'\\n', obtenido '%s'", testBody)
	}
}

// TestToUpper_EmptyOrWhitespaceText prueba casos de texto vacío o solo espacios.
func TestToUpper_EmptyOrWhitespaceText(t *testing.T) {
	tests := []struct {
		name     string
		inputText string
	}{
		{"Empty string", ""},
		{"Whitespace only", "   "},
		{"Tabs and newlines", "\t\n "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockConn{}
			testStatus = ""
			testBody = ""

			params := map[string]string{"text": tt.inputText}
			ToUpper(mockConn, params, mockSendResponse)

			if testStatus != "400 Bad Request" {
				t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
			}
			if testBody != "Falta el parámetro 'text'\n" { // The handler returns this for empty/whitespace after TrimSpace
				t.Errorf("Esperado body 'Falta el parámetro 'text'\\n', obtenido '%s'", testBody)
			}
		})
	}
}