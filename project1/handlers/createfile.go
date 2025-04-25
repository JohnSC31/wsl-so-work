package handlers

import (
    "net"
    "os"
    "strconv"
    "strings"
    "http-servidor/utils"
)

// /createfile?name=filename&content=text&repeat=x

func CreateFile(conn net.Conn, params map[string]string) {
    name, nameOk := params["name"]
    content, contentOk := params["content"]
    repeatStr, repeatOk := params["repeat"]

    if !nameOk || !contentOk || !repeatOk {
        utils.SendResponse(conn, "400 Bad Request", "Faltan parámetros: name, content, repeat\n")
        return
    }

    repeat, err := strconv.Atoi(repeatStr)
    if err != nil || repeat <= 0 {
        utils.SendResponse(conn, "400 Bad Request", "'repeat' debe ser un entero positivo\n")
        return
    }

    repeated := strings.Repeat(content+"\n", repeat)

    syncutils.FilesMutex.Lock() // uso del mutex
    defer syncutils.FilesMutex.Unlock()

    err = os.WriteFile("files/" + name, []byte(repeated), 0644)
    if err != nil {
        utils.SendResponse(conn, "500 Internal Server Error", "No se pudo crear el archivo\n")
        return
    }

    utils.SendResponse(conn, "200 OK", "Archivo creado exitosamente\n")
}
