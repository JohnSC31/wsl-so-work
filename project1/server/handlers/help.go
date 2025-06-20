package handlers

import (
	"http-servidor/utils"
	"net"
)

func Help(conn net.Conn) {
	body := `
    Rutas disponibles:
    - /help
    - /fibonacci?num=N
    - /createfile?name=filename&content=text&repeat=x
    - /deletefile?name=filename
    - /status
    - /reverse?text=abcdef
    - /toupper?text=abcd
    - /random?count=n&min=a&max=b
    - /timestamp
    - /hash?text=someinput
    - /simulate?seconds=s&task=name
    - /sleep?seconds=s
    - /loadtest?tasks=n&sleep=x
    `
	utils.SendResponse(conn, "200 OK", body)
}
