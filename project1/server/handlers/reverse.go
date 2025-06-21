package handlers

import (
    "net"
    "strings"
)

// /reverse?text=abc

func Reverse(conn net.Conn, params map[string]string, sendResponse SendResponseFunc) {
    text, ok := params["text"]
    if !ok || strings.TrimSpace(text) == ""{
        sendResponse(conn, "400 Bad Request", "Falta el parámetro 'text'\n")
        return
    }

    reversed := reverseString(text)
    sendResponse(conn, "200 OK", reversed + "\n")
}

func reverseString(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}
