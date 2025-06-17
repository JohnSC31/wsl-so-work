package handlers

import (
    "net"
    "strings"
    "http-servidor/utils"
)

// /toupper?text=abc

func ToUpper(conn net.Conn, params map[string]string) {
    text, ok := params["text"]
    if !ok || strings.TrimSpace(text) == "" {
        utils.SendResponse(conn, "400 Bad Request", "Falta el parámetro 'text'\n")
        return
    }

    utils.SendResponse(conn, "200 OK", strings.ToUpper(text) + "\n")
}
