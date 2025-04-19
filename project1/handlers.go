package main

import (
	"http-servidor/handlers"
	"http-servidor/utils"
)

func HandleRequest(req Request) {
	switch req.Ruta {

	case "/help":
		handlers.Help(req.Conn)

	case "/timestamp":
		handlers.Timestamp(req.Conn)

	case "/fibonacci":
		handlers.Fibonacci(req.Conn, req.Parametros)

	case "/createfile":
		handlers.CreateFile(req.Conn, req.Parametros)

	case "/deletefile":
		handlers.DeleteFile(req.Conn, req.Parametros)

	case "/reverse":
		handlers.Reverse(req.Conn, req.Parametros)

	case "/toupper":
		handlers.ToUpper(req.Conn, req.Parametros)

	case "/random":
		handlers.Random(req.Conn, req.Parametros["min"], req.Parametros["max"], req.Parametros["count"])

	case "/hash":
		handlers.Hash(req.Conn, req.Parametros["text"])

	case "/simulate":
		handlers.Simulate(req.Conn, req.Parametros["seconds"], req.Parametros["task"])

	case "/sleep":
		handlers.Sleep(req.Conn, req.Parametros["seconds"])

	case "/loadtest":
		handlers.Loadtest(req.Conn, req.Parametros["tasks"], req.Parametros["sleep"])
	default:
		utils.SendResponse(req.Conn, "404 Not Found", "Ruta no encontrada")
	}
}
