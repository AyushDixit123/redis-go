package transaction

import (
	"fmt"
	"net"
	"strconv"
)

func Handletransaction(conn net.Conn, list map[string]string, key string, values []string) {

	length := len(list[key])

	if length == 0 {
		list[key] = strconv.Itoa(1)

		response := fmt.Sprintf(":%d\r\n", length+1)

		conn.Write([]byte(response))
	} else {
		answer, err := strconv.Atoi(list[key])

		if err != nil {
			conn.Write([]byte(fmt.Sprintf("-ERR value is not an integer or out of range\r\n")))
			return
		}
		list[key] = strconv.Itoa(answer + 1)

		response := fmt.Sprintf(":%d\r\n", answer+1)

		conn.Write([]byte(response))
	}
}
