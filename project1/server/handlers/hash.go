package handlers

import (
	"crypto/sha256"
	"fmt"
	"net"
	"strings"
)

func Hash(conn net.Conn, text string, sendResponse SendResponseFunc) {

	if strings.TrimSpace(text) == "" {
        sendResponse(conn, "400 Bad Request", "Texto no puede ser vacio\n")
        return
    }

	hash := sha256.New()
	hash.Write([]byte(text))
	hashedText := hash.Sum(nil)

	hashedHex := fmt.Sprintf("%x", hashedText)

	body := "El hash SHA-256 del texto es:\n\n"
	body += hashedHex + "\n"

	sendResponse(conn, "200 OK", body)

}
