package handlers

import (
    "net"
    "strconv"
    "http-servidor/utils"
)

//  /fibonacci?num=N

func Fibonacci(conn net.Conn, params map[string]string) {
    nStr, ok := params["num"]
    if !ok {
        utils.SendResponse(conn, "400 Bad Request", "Falta el parámetro 'num'\n")
        return
    }

    n, err := strconv.Atoi(nStr)
    if err != nil || n < 0 {
        utils.SendResponse(conn, "400 Bad Request", "El parámetro 'num' debe ser un entero positivo\n")
        return
    }

    result := fibonacci(n)
    utils.SendResponse(conn, "200 OK", strconv.Itoa(result)+"\n")
}

func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}
