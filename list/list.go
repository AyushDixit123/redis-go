package list

import (
	"fmt"
	"net"
)

func HandleList(conn net.Conn, list map[string][]string, key string, values []string, lpush bool) {

	if lpush {
		for idx := 0; idx < len(values); idx++ {
			list[key] = append(list[key], values[idx])
		}

		length := len(list[key])

		// RESP Integer
		response := fmt.Sprintf(":%d\r\n", length)

		conn.Write([]byte(response))
	} else {
		for idx := 0; idx < len(values); idx++ {
			list[key] = append([]string{values[idx]}, list[key]...)
		}

		length := len(list[key])

		// RESP Integer
		response := fmt.Sprintf(":%d\r\n", length)

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
