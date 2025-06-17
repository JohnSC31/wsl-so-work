package handlers

import (
    "net"
    "os"
    "http-servidor/utils"
)

// /deletefile?name=filename

func DeleteFile(conn net.Conn, params map[string]string) {
    name, ok := params["name"]
    if !ok {
        utils.SendResponse(conn, "400 Bad Request", "Falta el parámetro 'name'\n")
        return
    }

    utils.FilesMutex.Lock() // uso del mutex
    defer utils.FilesMutex.Unlock()

    err := os.Remove("files/"+name)
    if err != nil {
        utils.SendResponse(conn, "500 Internal Server Error", "Error al eliminar el archivo (puede que no exista)\n")
        return
    }

    utils.SendResponse(conn, "200 OK", "Archivo eliminado exitosamente\n")
}
