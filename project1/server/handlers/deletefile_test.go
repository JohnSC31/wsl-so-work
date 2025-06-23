// deletefile_test.go
package handlers

import (
	"os"
	"testing"
)


// TestDeleteFile_Success prueba la eliminación exitosa de un archivo
func TestDeleteFile_Success(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	// Creamos un directorio 'files' y un archivo temporal para la prueba
	err := os.MkdirAll("files", 0755)
	if err != nil {
		t.Fatalf("No se pudo crear el directorio 'files' para la prueba: %v", err)
	}
	defer os.RemoveAll("files") // Limpiar después de la prueba

	fileName := "file_to_delete.txt"
	filePath := "files/" + fileName
	err = os.WriteFile(filePath, []byte("contenido de prueba"), 0644)
	if err != nil {
		t.Fatalf("No se pudo crear el archivo temporal para la prueba: %v", err)
	}


	params := map[string]string{
		"name": fileName,
	}

	// Llama a DeleteFile pasando nuestro mockSendResponse
	DeleteFile(mockConn, params, mockSendResponse)

	if testStatus != "200 OK" {
		t.Errorf("Esperado status '200 OK', obtenido '%s'", testStatus)
	}
	if testBody != "Archivo eliminado exitosamente\n" {
		t.Errorf("Esperado body 'Archivo eliminado exitosamente\\n', obtenido '%s'", testBody)
	}

	// Verificar que el archivo realmente fue eliminado
	_, err = os.Stat(filePath)
	if !os.IsNotExist(err) {
		t.Errorf("El archivo '%s' no fue eliminado como se esperaba, o error: %v", fileName, err)
	}
}

// TestDeleteFile_MissingNameParam prueba el caso de falta del parámetro 'name'
func TestDeleteFile_MissingNameParam(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	params := map[string]string{} // Parámetros sin 'name'

	// Llama a DeleteFile pasando nuestro mockSendResponse
	DeleteFile(mockConn, params, mockSendResponse)

	if testStatus != "400 Bad Request" {
		t.Errorf("Esperado status '400 Bad Request', obtenido '%s'", testStatus)
	}
	if testBody != "Falta el parámetro 'name'\n" {
		t.Errorf("Esperado body 'Falta el parámetro 'name'\\n', obtenido '%s'", testBody)
	}
}

// TestDeleteFile_FileNotFound prueba el caso en que el archivo no existe
func TestDeleteFile_FileNotFound(t *testing.T) {
	mockConn := &MockConn{}
	testStatus = ""
	testBody = ""

	// Creamos un directorio 'files' para que os.Remove pueda intentar buscar en él
	err := os.MkdirAll("files", 0755)
	if err != nil {
		t.Fatalf("No se pudo crear el directorio 'files' para la prueba: %v", err)
	}
	defer os.RemoveAll("files") // Limpiar después de la prueba

	params := map[string]string{
		"name": "non_existent_file.txt",
	}

	// Llama a DeleteFile pasando nuestro mockSendResponse
	DeleteFile(mockConn, params, mockSendResponse)

	if testStatus != "500 Internal Server Error" {
		t.Errorf("Esperado status '500 Internal Server Error', obtenido '%s'", testStatus)
	}
	if testBody != "Error al eliminar el archivo (puede que no exista)\n" {
		t.Errorf("Esperado body 'Error al eliminar el archivo (puede que no exista)\\n', obtenido '%s'", testBody)
	}
}