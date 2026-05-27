package list

import (
	"fmt"
	"net"
	"strconv"
)

func HandleList(conn net.Conn, list map[string][]string, key string, values []string, num int) {

	switch num {
	case 0:
		for idx := range values {
			list[key] = append(list[key], values[idx])
		}

		length := len(list[key])

		// RESP Integer
		response := fmt.Sprintf(":%d\r\n", length)

		conn.Write([]byte(response))
	case 1:
		for idx := 0; idx < len(values); idx++ {
			list[key] = append([]string{values[idx]}, list[key]...)
		}

		length := len(list[key])

		// RESP Integer
		response := fmt.Sprintf(":%d\r\n", length)

		conn.Write([]byte(response))
	case 2:
		length := len(list[key])

		// RESP Integer
		response := fmt.Sprintf(":%d\r\n", length)

		conn.Write([]byte(response))
	case 3:

		// Empty list
		if len(list[key]) == 0 {
			conn.Write([]byte("$-1\r\n"))
			return
		}

		// -----------------------------
		// NORMAL: LPOP key
		// Return BULK STRING
		// -----------------------------
		if len(values) == 0 {

			value := list[key][0]

			response := fmt.Sprintf(
				"$%d\r\n%s\r\n",
				len(value),
				value,
			)

			// Remove first element
			list[key] = list[key][1:]

			conn.Write([]byte(response))
			return
		}

		// -----------------------------
		// LPOP key count
		// Return RESP ARRAY
		// -----------------------------
		count, _ := strconv.Atoi(values[0])

		if count > len(list[key]) {
			count = len(list[key])
		}

		response := fmt.Sprintf("*%d\r\n", count)

		for idx := 0; idx < count; idx++ {

			value := list[key][idx]

			response += fmt.Sprintf(
				"$%d\r\n%s\r\n",
				len(value),
				value,
			)
		}

		list[key] = list[key][count:]

		conn.Write([]byte(response))

	}

}

// PREPENDING ELEMENTS IN GO SLICES
//
// Existing slice:
//
//	list[key] = ["banana", "orange"]
//
// We want to insert "apple" at the FRONT.
//
// Step 1:
//
//	[]string{values[idx]}
//
// If:
//
//	values[idx] = "apple"
//
// Then this creates a NEW slice:
//
//	["apple"]
//
// Step 2:
//
//	list[key]...
//
// "..." means:
// expand all elements individually
//
// So:
//
//	list[key]...
//
// becomes:
//
//	"banana", "orange"
//
// Step 3:
//
//	append([]string{"apple"}, list[key]...)
//
// becomes:
//
//	append(["apple"], "banana", "orange")
//
// Final result:
//
//	["apple", "banana", "orange"]
//
// Therefore:
//
//	list[key] = append([]string{values[idx]}, list[key]...)
//
// means:
// "create a new slice with current value at front,
// then append all old elements after it"
//
// ---------------------------------------------------
// WHY LOOP BACKWARD FOR LPUSH?
//
// Suppose:
//
//	values = ["a", "b", "c"]
//
// If we prepend LEFT -> RIGHT:
//
//	[a]
//	[b a]
//	[c b a]
//
// order becomes reversed.
//
// So we iterate from RIGHT -> LEFT:
//
//	for idx := len(values)-1; idx >= 0; idx--
//
// Result:
//
//	[c]
//	[b c]
//	[a b c]
//
// Original order preserved.
// This matches Redis LPUSH behavior.
