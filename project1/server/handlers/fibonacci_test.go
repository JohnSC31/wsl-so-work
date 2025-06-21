
package handlers 

import (
	"testing"
)


// TestFibonacci_ValidInput prueba casos exitosos con entradas válidas
func TestFibonacci_ValidInput(t *testing.T) {
	tests := []struct {
		name     string
		inputNum string
		expected string
	}{
		{"Fib(0)", "0", "0\n"},
		{"Fib(1)", "1", "1\n"},
		{"Fib(2)", "2", "1\n"},
		{"Fib(3)", "3", "2\n"},
		{"Fib(5)", "5", "5\n"},
		{"Fib(10)", "10", "55\n"},
		{"Fib(15)", "15", "610\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockConn{}
			testStatus = ""
			testBody = ""

			params := map[string]string{"num": tt.inputNum}
			Fibonacci(mockConn, params, mockSendResponse)

			if testStatus != "200 OK" {
				t.Errorf("Esperado status '200 OK', obtenido '%s'", testStatus)
			}
			if testBody != tt.expected {
				t.Errorf("Esperado body '%s', obtenido '%s'", tt.expected, testBody)
			}
		})
	}
}

// TestFibonacci_MissingParam prueba el caso de falta del parámetro 'num'
func TestFibonacci_MissingParam(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	params := map[string]string{} // Parámetros sin 'num'
	Fibonacci(mockConn, params, mockSendResponse)

	if testStatus != "400 Bad Request" {
		t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
	}
	if testBody != "Falta el parámetro 'num'\n" {
		t.Errorf("Esperado body 'Falta el parámetro 'num'\\n', obtenido '%s'", testBody)
	}
}

// TestFibonacci_InvalidNum prueba casos de 'num' no numérico o negativo
func TestFibonacci_InvalidNum(t *testing.T) {
	tests := []struct {
		name     string
		inputNum string
		expected string
	}{
		{"non-numeric", "abc", "El parámetro 'num' debe ser un entero positivo\n"},
		{"negative", "-5", "El parámetro 'num' debe ser un entero positivo\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &MockConn{}
			testStatus = ""
			testBody = ""

			params := map[string]string{"num": tt.inputNum}
			Fibonacci(mockConn, params, mockSendResponse)

			if testStatus != "400 Bad Request" {
				t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
			}
			if testBody != tt.expected {
				t.Errorf("Esperado body '%s', obtenido '%s'", tt.expected, testBody)
			}
		})
	}
}

// TestFibonacci_LargeInput verifica que la función fibonacci maneje números más grandes correctamente
// Aunque la implementación recursiva pura es ineficiente para números grandes, este test
// asegura que el valor devuelto sea el correcto para un N moderadamente grande.
// Ten en cuenta que para valores de N muy grandes (ej. > 40-45), la recursión pura será muy lenta.
func TestFibonacci_LargeInput(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	params := map[string]string{"num": "20"}
	Fibonacci(mockConn, params, mockSendResponse)

	if testStatus != "200 OK" {
		t.Errorf("Esperado status '200 OK', obtenido '%s'", testStatus)
	}
	// Fibonacci(20) = 6765
	if testBody != "6765\n" {
		t.Errorf("Esperado body '6765\\n', obtenido '%s'", testBody)
	}
}

// Test_fibonacci_HelperFunction prueba directamente la función auxiliar fibonacci
func Test_fibonacci_HelperFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"Fib(0)", 0, 0},
		{"Fib(1)", 1, 1},
		{"Fib(2)", 2, 1},
		{"Fib(3)", 3, 2},
		{"Fib(5)", 5, 5},
		{"Fib(10)", 10, 55},
		{"Fib(15)", 15, 610},
		{"Fib(20)", 20, 6765},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fibonacci(tt.input)
			if result != tt.expected {
				t.Errorf("Para n=%d, esperado %d, obtenido %d", tt.input, tt.expected, result)
			}
		})
	}
}