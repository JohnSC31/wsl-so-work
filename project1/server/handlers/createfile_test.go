// createfile_test.go
package handlers

import (
	"os"
	"testing"
)


// TestCreateFile_Success prueba la creación exitosa de un archivo
func TestCreateFile_Success(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	// Crear un directorio 'files' temporal para las pruebas
	err := os.MkdirAll("files", 0755)
	if err != nil {
		t.Fatalf("No se pudo crear el directorio 'files' para la prueba: %v", err)
	}
	defer os.RemoveAll("files") // Limpiar después de la prueba

	params := map[string]string{
		"name":    "testfile.txt",
		"content": "hello world",
		"repeat":  "2",
	}

	CreateFile(mockConn, params, mockSendResponse)

	if testStatus != "200 OK" {
		t.Errorf("Esperado status '200 OK', obtenido '%s'", testStatus)
	}
	if testBody != "Archivo creado exitosamente\n" {
		t.Errorf("Esperado body 'Archivo creado exitosamente\\n', obtenido '%s'", testBody)
	}

	// Verificar que el archivo fue creado y tiene el contenido correcto
	filePath := "files/testfile.txt"
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Error al leer el archivo creado: %v", err)
	}

	expectedContent := "hello world\nhello world\n"
	if string(content) != expectedContent {
		t.Errorf("Contenido del archivo inesperado.\nEsperado:\n%sObtenido:\n%s", expectedContent, string(content))
	}
}

// TestCreateFile_MissingParams prueba el caso de parámetros faltantes
func TestCreateFile_MissingParams(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	tests := []struct {
		name     string
		params   map[string]string
		expected string
	}{
		{
			name:     "missing name",
			params:   map[string]string{"content": "abc", "repeat": "1"},
			expected: "Faltan parámetros: name, content, repeat\n",
		},
		{
			name:     "missing content",
			params:   map[string]string{"name": "file.txt", "repeat": "1"},
			expected: "Faltan parámetros: name, content, repeat\n",
		},
		{
			name:     "missing repeat",
			params:   map[string]string{"name": "file.txt", "content": "abc"},
			expected: "Faltan parámetros: name, content, repeat\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CreateFile(mockConn, tt.params, mockSendResponse)

			if testStatus != "400 Bad Request" {
				t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
			}
			if testBody != tt.expected {
				t.Errorf("Esperado body '%s', obtenido '%s'", tt.expected, testBody)
			}
		})
	}
}

// TestCreateFile_InvalidRepeat prueba el caso de 'repeat' no numérico o negativo/cero
func TestCreateFile_InvalidRepeat(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	tests := []struct {
		name     string
		repeat   string
		expected string
	}{
		{
			name:     "non-numeric repeat",
			repeat:   "abc",
			expected: "'repeat' debe ser un entero positivo\n",
		},
		{
			name:     "zero repeat",
			repeat:   "0",
			expected: "'repeat' debe ser un entero positivo\n",
		},
		{
			name:     "negative repeat",
			repeat:   "-5",
			expected: "'repeat' debe ser un entero positivo\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]string{
				"name":    "file.txt",
				"content": "some content",
				"repeat":  tt.repeat,
			}
			CreateFile(mockConn, params, mockSendResponse)

			if testStatus != "400 Bad Request" {
				t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
			}
			if testBody != tt.expected {
				t.Errorf("Esperado body '%s', obtenido '%s'", tt.expected, testBody)
			}
		})
	}
}

// TestCreateFile_WriteFileError prueba un error al escribir el archivo
// Esto es más complejo de simular directamente con os.WriteFile,
// normalmente requeriría inyectar una interfaz para la operación de escritura de archivos.
// Para simplificar, podemos simular una ruta de archivo inválida que cause un error de escritura.
func TestCreateFile_WriteFileError(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	// Intentamos escribir en una ruta inválida para forzar un error
	// Por ejemplo, un directorio que no existe y no puede ser creado por os.WriteFile
	// (si no tiene permisos, etc.). Aquí usamos un carácter inválido en el nombre para forzar error.
	params := map[string]string{
		"name":    "invalid/file?.txt", // Carácter inválido para un nombre de archivo
		"content": "content",
		"repeat":  "1",
	}

	CreateFile(mockConn, params, mockSendResponse)

	if testStatus != "500 Internal Server Error" {
		t.Errorf("Esperado status '500 Internal Server Error', obtenido '%s'", testStatus)
	}
	if testBody != "No se pudo crear el archivo\n" {
		t.Errorf("Esperado body 'No se pudo crear el archivo\\n', obtenido '%s'", testBody)
	}
	// Limpiar cualquier intento de creación de directorio o archivo si lo hubiera
	os.RemoveAll("invalid")
}