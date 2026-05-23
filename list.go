package main

import (
	"fmt"
	"net"
)

func HandleList(conn net.Conn, list map[string][]string, key string, value string) {

	list[key] = append(list[key], value)

	length := len(list[key])

	// RESP Integer
	response := fmt.Sprintf(":%d\r\n", length)

	conn.Write([]byte(response))
}
