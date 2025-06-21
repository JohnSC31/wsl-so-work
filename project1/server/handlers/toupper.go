package handlers

import (
    "net"
    "strings"
)

// /toupper?text=abc

func ToUpper(conn net.Conn, params map[string]string, sendResponse SendResponseFunc) {
    text, ok := params["text"]
    if !ok || strings.TrimSpace(text) == "" {
        sendResponse(conn, "400 Bad Request", "Falta el parámetro 'text'\n")
        return
    }

    sendResponse(conn, "200 OK", strings.ToUpper(text) + "\n")
}
