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
		handlers.Timestamp(req.Conn, utils.SendResponse)

	case "/fibonacci":
		print("Fibonacci request received\n")
		handlers.Fibonacci(req.Conn, req.Parametros, utils.SendResponse)

	case "/createfile":
		handlers.CreateFile(req.Conn, req.Parametros, utils.SendResponse)

	case "/deletefile":
		handlers.DeleteFile(req.Conn, req.Parametros, utils.SendResponse)

	case "/reverse":
		handlers.Reverse(req.Conn, req.Parametros, utils.SendResponse)

	case "/toupper":
		handlers.ToUpper(req.Conn, req.Parametros, utils.SendResponse)

	case "/random":
		handlers.Random(req.Conn, req.Parametros["min"], req.Parametros["max"], req.Parametros["count"], utils.SendResponse)

	case "/hash":
		handlers.Hash(req.Conn, req.Parametros["text"], utils.SendResponse)

	case "/simulate":
		handlers.Simulate(req.Conn, req.Parametros["seconds"], req.Parametros["task"], utils.SendResponse)

	case "/sleep":
		handlers.Sleep(req.Conn, req.Parametros["seconds"], utils.SendResponse)

	case "/loadtest":
		handlers.Loadtest(req.Conn, req.Parametros["tasks"], req.Parametros["sleep"], utils.SendResponse)

	case "/ping":
		handlers.HandlePing(req.Conn)

	default:
		utils.SendResponse(req.Conn, "404 Not Found", "Ruta no encontrada")
	}
}
