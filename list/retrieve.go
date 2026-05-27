package list

import (
	"fmt"
	"net"
	"strconv"
)

func Retrievelist(
	conn net.Conn,
	key string,
	start string,
	end string,
	list map[string][]string,
) {

	si, _ := strconv.Atoi(start)
	ei, _ := strconv.Atoi(end)

	values := list[key]
	length := len(values)

	// Handle negative indexes
	if si < 0 {
		si = si + length
	}

	if ei < 0 {
		ei = ei + length
	}

	endIndex := ei + 1

	// Prevent end from going out of bounds
	if endIndex > length {
		endIndex = length
	}

	// Prevent start from being negative
	if si < 0 {
		si = 0
	}

	// Handle empty or invalid ranges safely
	if si >= length || si >= endIndex {
		conn.Write([]byte("*0\r\n"))
		return
	}

	count := endIndex - si

	// RESP array header
	response := fmt.Sprintf("*%d\r\n", count)

	for idx := si; idx < endIndex; idx++ {

		response += fmt.Sprintf(
			"$%d\r\n%s\r\n",
			len(values[idx]),
			values[idx],
		)

	}

	// Send RESP array response
	conn.Write([]byte(response))
}
