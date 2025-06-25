package main
import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"http-servidor/utils"
)

// NUEVA FUNCIÓN: Maneja la solicitud de conteo de palabras de archivos grandes
// Ahora recibe el bufio.Reader para leer el cuerpo de la solicitud POST
func (d *Dispatcher) handleWordCount(conn net.Conn, method, path string, params map[string]string, headers map[string]string, reader *bufio.Reader) {
	// 1. Recibir el contenido del archivo desde el cuerpo de la solicitud POST
	contentLengthStr, ok := headers["Content-Length"]
	var contentLength int
	if ok {
		var err error
		contentLength, err = strconv.Atoi(contentLengthStr)
		if err != nil {
			utils.SendResponse(conn, "400 Bad Request", "Content-Length inválido")
			log.Printf("Error: Content-Length inválido: %v", err)
			d.Metrics.mu.Lock()
			d.Metrics.RequestsFailed++
			d.Metrics.mu.Unlock()
			return
		}
	} else {
		log.Println("Advertencia: No Content-Length header. Leyendo hasta EOF/timeout.")
	}

	var contentBuilder strings.Builder
	var bytesRead int
	buffer := make([]byte, 4096) // Buffer para leer chunks del cuerpo
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			contentBuilder.Write(buffer[:n])
			bytesRead += n
		}
		if err == io.EOF || (contentLength > 0 && bytesRead >= contentLength) {
			break
		}
		if err != nil {
			utils.SendResponse(conn, "500 Internal Server Error", "Error leyendo el cuerpo del archivo")
			log.Printf("Error leyendo el cuerpo del archivo: %v", err)
			d.Metrics.mu.Lock()
			d.Metrics.RequestsFailed++
			d.Metrics.mu.Unlock()
			return
		}
	}
	content := contentBuilder.String()
	log.Printf("Archivo recibido, tamaño: %d bytes", len(content))

	// 2. Dividir el archivo en chunks
	lines := strings.Split(content, "\n")
	// Limpieza: si el archivo termina con \n, Split crea una última cadena vacía. La eliminamos.
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	// Si el archivo está completamente vacío después de limpiar las líneas
	if len(lines) == 0 {
		utils.SendResponse(conn, "200 OK", "Conteo total de palabras: 0\n")
		log.Println("Archivo vacío después de procesar, retorno 0 palabras.")
		d.Metrics.mu.Lock()
		d.Metrics.RequestsHandled++
		d.Metrics.mu.Unlock()
		return
	}

	numWorkers := len(d.Workers)
	if numWorkers == 0 {
		utils.SendResponse(conn, "503 Service Unavailable", "No hay workers disponibles para el conteo de palabras")
		log.Println("No hay workers disponibles para conteo de palabras")
		d.Metrics.mu.Lock()
		d.Metrics.RequestsFailed++
		d.Metrics.mu.Unlock()
		return
	}

	// Calcula el tamaño base del chunk y cuántos trabajadores recibirán un chunk extra
	baseChunkSize := len(lines) / numWorkers
	extraChunks := len(lines) % numWorkers

	var wg sync.WaitGroup
	// El tamaño del canal de resultados debe ser al menos el número de workers,
	// pero no se bloqueará si lanzamos menos goroutines.
	resultsChan := make(chan WorkerResult, numWorkers)

	// Contar los workers a los que realmente se les asignará una tarea
	workersReceivingTasks := 0 
	for i := 0; i < numWorkers; i++ {
		workerLines := baseChunkSize
		if i < extraChunks {
			workerLines++ // Distribuir las líneas restantes
		}

		// Si este worker no tiene líneas para procesar, NO LANZAMOS GOROUTINE
		if workerLines == 0 {
			log.Printf("Dispatcher: Saltando worker %d, 0 líneas asignadas.", i+1)
			continue
		}

		startLine := 0
		if i > 0 {
			// Calcular el inicio sumando los tamaños de los chunks anteriores
			for j := 0; j < i; j++ {
				prevChunkSize := baseChunkSize
				if j < extraChunks {
					prevChunkSize++
				}
				startLine += prevChunkSize
			}
		}
		endLine := startLine + workerLines

		// Protección extra, aunque con la lógica de baseChunkSize/extraChunks debería ser raro
		if startLine >= endLine || endLine > len(lines) {
			log.Printf("Dispatcher: Error lógico en la división de chunks para worker %d (start: %d, end: %d, total lines: %d). Saltando.", i+1, startLine, endLine, len(lines))
			continue
		}

		chunkContent := strings.Join(lines[startLine:endLine], "\n")
		
		worker := seleccionarWorker(d)
		if worker == nil {
			log.Printf("Dispatcher: No se pudo seleccionar worker para el chunk de worker %d, reintentando o fallando.", i+1)
			// Aquí se podría implementar una cola de reintentos o marcar la tarea como fallida
			continue
		}
		
		workersReceivingTasks++ // Solo incrementamos si realmente se lanza una goroutine

		wg.Add(1)
		go func(w *Worker, currentChunkContent string, chunkID int) {
			defer wg.Done()
			log.Printf("Dispatcher: Enviando chunk %d (tamaño %d bytes) a worker %d (%s)", chunkID, len(currentChunkContent), w.ID, w.URL)

			wordCountStr, err := d.sendPostToWorker(w, "/countchunk", currentChunkContent)
			if err != nil {
				resultsChan <- WorkerResult{WorkerID: fmt.Sprintf("Worker-%d", w.ID), Error: fmt.Errorf("error enviando chunk %d a worker %s: %w", chunkID, w.URL, err)}
				d.redistributeTasks(w) // Redistribuir tareas si el worker falla
				return
			}
			log.Printf("Dispatcher: Worker %d respondió con: '%s'", w.ID, wordCountStr)

			wordCount, err := strconv.Atoi(strings.TrimSpace(wordCountStr))
			if err != nil {
				resultsChan <- WorkerResult{WorkerID: fmt.Sprintf("Worker-%d", w.ID), Error: fmt.Errorf("error parseando conteo de palabras de worker %s para chunk %d: %w", w.URL, chunkID, err)}
				return
			}
			resultsChan <- WorkerResult{WorkerID: fmt.Sprintf("Worker-%d", w.ID), Count: wordCount}

		}(worker, chunkContent, i+1)
	}

	// Si no se lanzaron tareas a ningún worker (ej. archivo muy pequeño para un solo worker o no hay workers disponibles)
	if workersReceivingTasks == 0 {
		utils.SendResponse(conn, "200 OK", "Conteo total de palabras: 0 (No se enviaron tareas a workers)\n")
		log.Println("No se enviaron tareas a workers, retorno 0 palabras.")
		d.Metrics.mu.Lock()
		d.Metrics.RequestsHandled++
		d.Metrics.mu.Unlock()
		return
	}


	// Cerrar el canal de resultados cuando todas las goroutines de workers terminen
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// 4. Recopilar y sumar los resultados
	totalWordCount := 0
	var errors []error
	for res := range resultsChan {
		if res.Error != nil {
			log.Printf("Error recibido de worker %s: %v", res.WorkerID, res.Error)
			errors = append(errors, res.Error)
		} else {
			totalWordCount += res.Count
			log.Printf("Worker %s contribuyó con %d palabras.", res.WorkerID, res.Count)
		}
	}

	// 5. Retornar el resultado total al cliente
	if len(errors) > 0 {
		errMsg := fmt.Sprintf("Errores durante el procesamiento: %v. Conteo parcial: %d", errors, totalWordCount)
		utils.SendResponse(conn, "500 Internal Server Error", errMsg)
		d.Metrics.mu.Lock()
		d.Metrics.RequestsFailed++
		d.Metrics.mu.Unlock()
		return
	}

	log.Printf("Conteo total de palabras: %d", totalWordCount)
	utils.SendResponse(conn, "200 OK", fmt.Sprintf("Conteo total de palabras: %d\n", totalWordCount))
	d.Metrics.mu.Lock()
	d.Metrics.RequestsHandled++
	d.Metrics.mu.Unlock()
}