package main

import (
    // "fmt"
    "log"
    "net"
    // "strings"
    "http-servidor/handlers"
    "http-servidor/utils"
)

const PORT = ":8080"

func main() {
    ln, err := net.Listen("tcp", PORT)
    if err != nil {
        log.Fatalf("Error al iniciar servidor: %v", err)
    }
    defer ln.Close()
    log.Printf("Servidor escuchando en %s", PORT)

    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Printf("Error aceptando conexión: %v", err)
            continue
        }
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()

    buffer := make([]byte, 1024)
    n, err := conn.Read(buffer)
    if err != nil {
        log.Printf("Error leyendo: %v", err)
        return
    }

    request := string(buffer[:n])
    method, path := utils.ParseRequestLine(request)

    if method != "GET" {
        utils.SendResponse(conn, "405 Method Not Allowed", "Solo se permite GET")
        return
    }

    route, _ := utils.ParseRoute(path)

    switch route {
    case "/help":
        handlers.Help(conn)
    case "/timestamp":
        handlers.Timestamp(conn)
    default:
        utils.SendResponse(conn, "404 Not Found", "Ruta no encontrada")
    }
}
