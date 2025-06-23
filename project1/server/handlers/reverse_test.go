// reverse_test.go
package handlers

import (
	"testing"
	"strings"
)

// TestReverse_ValidInput prueba casos exitosos con entradas de texto válidas
func TestReverse_ValidInput(t *testing.T) {
	tests := []struct {
		name     string
		inputText string
		expected string
	}{
		{"English Word", "hello", "olleh\n"},
		{"Sentence", "Go is fun", "nuf si oG\n"},
		{"Palindrome", "madam", "madam\n"},
		{"Empty String (should be handled by valid input, not error)", "", "\n"}, // empty string is valid for reverseString, error for Reverse handler
		{"Unicode Characters", "你好世界", "界世好你\n"}, // Chinese characters
		{"String with spaces", " a b ", " b a \n"},
		{"Single character", "a", "a\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockConn{}
			testStatus = ""
			testBody = ""

			params := map[string]string{"text": tt.inputText}
			Reverse(mockConn, params, mockSendResponse)

			// Note: The Reverse handler itself checks for TrimSpace(text) == ""
			// So an empty string or just spaces will result in a 400 Bad Request
			if strings.TrimSpace(tt.inputText) == "" {
				if testStatus != "400 Bad Request" {
					t.Errorf("Esperado status '400 Bad Request' para '%s', obtenido '%s'", tt.inputText, testStatus)
				}
				if testBody != "Falta el parámetro 'text'\n" {
					t.Errorf("Esperado body 'Falta el parámetro 'text'\\n', obtenido '%s'", testBody)
				}
			} else {
				if testStatus != "200 OK" {
					t.Errorf("Esperado status '200 OK' para '%s', obtenido '%s'", tt.inputText, testStatus)
				}
				if testBody != tt.expected {
					t.Errorf("Esperado body '%s', obtenido '%s' para input '%s'", tt.expected, testBody, tt.inputText)
				}
			}
		})
	}
}

// TestReverse_MissingParam prueba el caso de falta del parámetro 'text'
func TestReverse_MissingParam(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	params := map[string]string{} // Parámetros sin 'text'
	Reverse(mockConn, params, mockSendResponse)

	if testStatus != "400 Bad Request" {
		t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
	}
	if testBody != "Falta el parámetro 'text'\n" {
		t.Errorf("Esperado body 'Falta el parámetro 'text'\\n', obtenido '%s'", testBody)
	}
}

// TestReverse_EmptyOrWhitespaceText prueba casos de texto vacío o solo espacios
func TestReverse_EmptyOrWhitespaceText(t *testing.T) {
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
			Reverse(mockConn, params, mockSendResponse)

			if testStatus != "400 Bad Request" {
				t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
			}
			if testBody != "Falta el parámetro 'text'\n" { // The handler returns this for empty/whitespace after TrimSpace
				t.Errorf("Esperado body 'Falta el parámetro 'text'\\n', obtenido '%s'", testBody)
			}
		})
	}
}

// Test_reverseString_HelperFunction prueba directamente la función auxiliar reverseString
func Test_reverseString_HelperFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"English Word", "hello", "olleh"},
		{"Sentence", "Go is fun", "nuf si oG"},
		{"Palindrome", "madam", "madam"},
		{"Empty String", "", ""},
		{"Unicode Characters", "你好世界", "界世好你"},
		{"String with spaces", " a b ", " b a "},
		{"Single character", "a", "a"},
		{"Special characters", "!@#$%^", "^%$#@!"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reverseString(tt.input)
			if result != tt.expected {
				t.Errorf("Para input '%s', esperado '%s', obtenido '%s'", tt.input, tt.expected, result)
			}
		})
	}
}