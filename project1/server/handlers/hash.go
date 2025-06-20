package handlers

import (
	"crypto/sha256"
	"fmt"
	"http-servidor/utils"
	"net"
	"strings"
)

func Hash(conn net.Conn, text string) {

	if strings.TrimSpace(text) == "" {
        utils.SendResponse(conn, "400 Bad Request", "Texto no puede ser vacio\n")
        return
    }

	hash := sha256.New()
	hash.Write([]byte(text))
	hashedText := hash.Sum(nil)

	hashedHex := fmt.Sprintf("%x", hashedText)

	body := "El hash SHA-256 del texto es:\n\n"
	body += hashedHex + "\n"

	utils.SendResponse(conn, "200 OK", body)

}
