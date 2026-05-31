package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"redis-go/list"
	"redis-go/transaction"
)

// var _ = net.Listen
// var _ = os.Exit
var waitersMu sync.Mutex //adding mutex to avoud race conditions

func main() {

	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	store := make(map[string]string)

	expiry := make(map[string]time.Time) //outside becasue redis has shared database

	liststore := make(map[string][]string)

	waiters := make(map[string][]chan string) //key -> list of blocked clients waiting for data

	intrans := false
	steps := 0

	for {

		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
			break
		}

		go func() {

			// access to tcp connection
			//conn, err := l.Accept()
			// conn is an object

			// Send the response to the client
			//	conn.Write([]byte("+PONG\r\n"))

			buf := make([]byte, 1024) //make()-> does memory allocation similar to new()
			//slice of byte is created to hold incoming data

			for {
				n, err := conn.Read(buf)

				if err != nil {
					break // sender has stopped sending commands
				}

				input := string(buf[:n]) //converts a byte slice (buf) into a string.

				parts := strings.Split(input, "\r\n")

				cmd := strings.ToUpper(parts[2])

				switch cmd {

				case "PING":
					conn.Write([]byte("+PONG\r\n")) //Redis client libraries (especially in languages like Go) use a slice of bytes ([]byte) instead of standard strings because slices of bytes prevent memory allocations and reduce CPU overhead.
				//also Network sockets read data directly into byte buffers. If the library returned a string, it would have to copy those bytes into a brand new string object in memory. Returning []byte allows the library to pass a direct pointer to the network data without copying it.

				case "ECHO":
					message := parts[4]

					response := fmt.Sprintf("$%d\r\n%s\r\n", len(message), message)

					conn.Write([]byte(response))

					//saving set key value pairs in map
				case "SET":

					key := parts[4]

					value := parts[6]

					store[key] = value

					//Atoi (short for "ASCII to integer") is a function in the strconv package used to convert a string representation of a base-10 number into an int.

					if len(parts) > 8 { // if exp used

						option := strings.ToUpper(parts[8])

						if option == "PX" {

							durationMs, _ := strconv.Atoi(parts[10])

							expiry[key] = time.Now().Add(
								time.Duration(durationMs) * time.Millisecond,
							)
						} else {
							sec, _ := strconv.Atoi(parts[10])

							expiry[key] = time.Now().Add(
								time.Duration(sec) * time.Second)
						}
					}

					response := fmt.Sprintf("+OK\r\n")
					conn.Write([]byte(response))

				case "GET":

					key := parts[4]
					expTime, exists := expiry[key]

					if exists && time.Now().After(expTime) {

						delete(store, key)
						delete(expiry, key)

						conn.Write([]byte("$-1\r\n"))

					} else {

						value, exists := store[key]

						if !exists {

							conn.Write([]byte("$-1\r\n"))

						} else {

							response := fmt.Sprintf(
								"$%d\r\n%s\r\n",
								len(value),
								value,
							)

							conn.Write([]byte(response))
						}
					}

				case "RPUSH":

					key := parts[4]

					values := []string{}

					// RESP alternates:
					// length
					// actual value
					for i := 6; i < len(parts); i += 2 {
						values = append(values, parts[i])
					}

					handled := false

					waitersMu.Lock()

					if len(waiters[key]) > 0 {

						ch := waiters[key][0]

						// Remove first waiting client
						waiters[key] = waiters[key][1:]

						waitersMu.Unlock()

						// Wake blocked BLPOP
						ch <- values[0]

						handled = true

					} else {

						waitersMu.Unlock()
					}

					// No blocked clients
					if !handled {

						list.HandleList(conn, liststore, key, values, 0)

					} else {

						// Redis still returns pushed length
						conn.Write([]byte(":1\r\n"))
					}

				case "LRANGE":
					key := parts[4]

					start := parts[6]

					end := parts[8]

					list.Retrievelist(conn, key, start, end, liststore)

				case "LPUSH":

					key := parts[4]

					values := []string{}

					for i := 6; i < len(parts); i += 2 {
						values = append(values, parts[i])
					}

					handled := false

					waitersMu.Lock()

					if len(waiters[key]) > 0 {

						ch := waiters[key][0]

						// Remove first waiting client
						waiters[key] = waiters[key][1:]

						waitersMu.Unlock()

						// LPUSH inserts at front
						ch <- values[len(values)-1]

						handled = true

					} else {

						waitersMu.Unlock()
					}

					if !handled {

						list.HandleList(conn, liststore, key, values, 1)

					} else {

						// Redis still returns pushed length
						conn.Write([]byte(":1\r\n"))
					}

				case "LLEN":

					key := parts[4]

					values := []string{}

					list.HandleList(conn, liststore, key, values, 2)

				case "LPOP":

					key := parts[4]

					values := []string{}
					if len(parts) > 6 {
						values = append(values, parts[6])
					}

					list.HandleList(conn, liststore, key, values, 3)

				// A select statement blocks until one of its cases is ready to execute.Case Execution: Each case must be a channel operation (either sending or receiving data).
				// 						/select means:

				// "Wait for whichever event happens first."

				case "INCR":
					key := parts[4]

					values := []string{}

					transaction.Handletransaction(conn, store, key, values)
				case "MULTI":
					intrans = true
					conn.Write([]byte("+OK\r\n"))

				case "EXEC":

					if !intrans {
						conn.Write([]byte("-ERR EXEC without MULTI\r\n"))
						continue
					}

					if steps == 0 {
						conn.Write([]byte("*0\r\n"))

						intrans = false
						steps = 0

						continue
					}

					intrans = false
					steps = 0

				case "BLPOP":

					keys := []string{}

					for i := 4; i < len(parts)-2; i += 2 {
						keys = append(keys, parts[i])
					}

					// Last argument is timeout
					timeoutStr := parts[len(parts)-2]

					timeout, _ := strconv.ParseFloat(timeoutStr, 64)

					// -----------------------------------
					// First check if ANY key already has values
					// Redis checks LEFT -> RIGHT
					// -----------------------------------

					found := false

					for _, key := range keys {

						if len(liststore[key]) > 0 {

							value := liststore[key][0]

							// Remove first element
							liststore[key] = liststore[key][1:]

							response := fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(value), value)

							conn.Write([]byte(response))

							found = true

							break
						}
					}

					// -----------------------------------
					// If some key had value,
					// already handled
					// -----------------------------------

					if found {
						break
					}

					// -----------------------------------
					// No list had values
					// Need to BLOCK
					// -----------------------------------

					// Create private channel
					ch := make(chan string)

					// Register waiter on ALL keys
					waitersMu.Lock()
					for _, key := range keys {
						waiters[key] = append(waiters[key], ch)
					}
					waitersMu.Unlock()
					// -----------------------------------
					// timeout = 0
					// wait forever
					// -----------------------------------

					if timeout == 0 {

						value := <-ch

						selectedKey := ""

						for _, key := range keys {

							selectedKey = key

							break
						}

						response := fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(selectedKey), selectedKey, len(value), value)

						conn.Write([]byte(response))

					} else {

						select {

						// RPUSH/LPUSH woke client
						case value := <-ch:

							selectedKey := ""

							for _, key := range keys {

								selectedKey = key

								break
							}

							response := fmt.Sprintf(
								"*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",
								len(selectedKey),
								selectedKey,
								len(value),
								value,
							)

							conn.Write([]byte(response))

						// Timeout
						case <-time.After(
							time.Duration(timeout * float64(time.Second)),
						):

							waitersMu.Lock()

							for _, key := range keys {

								filtered := []chan string{}

								for _, waiter := range waiters[key] {

									if waiter != ch {
										filtered = append(filtered, waiter)
									}
								}

								waiters[key] = filtered
							}

							waitersMu.Unlock()

							conn.Write([]byte("*-1\r\n"))
						}
					}
				}
			}
		}()

	}

}
