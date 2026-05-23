package main

import (
	"fmt"
	"net"
)

func HandleList(conn net.Conn, list map[string][]string, key string, values []string) {

	for idx := 0; idx < len(values); idx++ {
		list[key] = append(list[key], values[idx])
	}

	length := len(list[key])

	// RESP Integer
	response := fmt.Sprintf(":%d\r\n", length)

	conn.Write([]byte(response))
}
