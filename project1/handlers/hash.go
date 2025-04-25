package handlers

import (
	"crypto/sha256"
	"fmt"
	"http-servidor/utils"
	"net"
)

func Hash(conn net.Conn, text string) {

	hash := sha256.New()
	hash.Write([]byte(text))
	hashedText := hash.Sum(nil)

	hashedHex := fmt.Sprintf("%x", hashedText)

	body := "El hash SHA-256 del texto es:\n\n"
	body += hashedHex

	utils.SendResponse(conn, "200 OK", body)

}
