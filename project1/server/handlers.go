package main

import (
	"bytes"
	"http-servidor/handlers"
	"http-servidor/utils"
	"log"
	"net/http"
	"net/url"
)

func HandleRequest(req Request) {
	// Extraer request_id y callback_url de los parámetros
	requestID := req.Parametros["request_id"]
	callbackURL := req.Parametros["callback_url"]
	// Validar parámetros requeridos
	if requestID == "" || callbackURL == "" {
		log.Printf("Faltan parámetros request_id o callback_url ", requestID, callbackURL)
		return
	}

	// Eliminar estos parámetros para que no interfieran con los handlers
	delete(req.Parametros, "request_id")
	delete(req.Parametros, "callback_url")

	// Buffer para capturar la respuesta
	var buf bytes.Buffer
	fakeConn := &utils.FakeConn{Buffer: &buf}
	switch req.Ruta {

	case "/help":
		handlers.Help(fakeConn)

	case "/timestamp":
		handlers.Timestamp(fakeConn)

	case "/fibonacci":
		print("Fibonacci request received\n")
		handlers.Fibonacci(fakeConn, req.Parametros)

	case "/createfile":
		handlers.CreateFile(fakeConn, req.Parametros)

	case "/deletefile":
		handlers.DeleteFile(fakeConn, req.Parametros)

	case "/reverse":
		handlers.Reverse(fakeConn, req.Parametros)

	case "/toupper":
		handlers.ToUpper(fakeConn, req.Parametros)

	case "/random":
		handlers.Random(fakeConn, req.Parametros["min"], req.Parametros["max"], req.Parametros["count"])

	case "/hash":
		handlers.Hash(fakeConn, req.Parametros["text"])

	case "/simulate":
		handlers.Simulate(fakeConn, req.Parametros["seconds"], req.Parametros["task"])

	case "/sleep":
		handlers.Sleep(fakeConn, req.Parametros["seconds"])

	case "/loadtest":
		handlers.Loadtest(fakeConn, req.Parametros["tasks"], req.Parametros["sleep"])

	case "/ping":
		handlers.HandlePing(fakeConn)

	default:
		utils.SendResponse(fakeConn, "404 Not Found", "Ruta no encontrada")
	}

	// hace el callback
	go func() {
		data := url.Values{}
		data.Set("request_id", requestID)
		data.Set("result", buf.String())

		resp, err := http.PostForm(callbackURL, data)
		if err != nil {
			log.Printf("Error enviando callback: %v", err)
			return
		}
		defer resp.Body.Close()
		log.Printf("Callback enviado correctamente, status: %s", resp.Status)
	}()
}
