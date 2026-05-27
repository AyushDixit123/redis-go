package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"redis-go/list"
)

// var _ = net.Listen
// var _ = os.Exit

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

					//Because RESP format alternates between:

					/*length
					  actual value
					  length
					  actual value*/
					for i := 6; i < len(parts); i += 2 {
						values = append(values, parts[i])
					}

					list.HandleList(conn, liststore, key, values, 0) //0 fro RPush 1 fro Lpush 2 for Llen

				case "LRANGE":
					key := parts[4]

					start := parts[6]

					end := parts[8]

					list.Retrievelist(conn, key, start, end, liststore)

				case "LPUSH":

					key := parts[4]
					values := []string{}

					//Because RESP format alternates between:

					/*length
					  actual value
					  length
					  actual value*/
					for i := 6; i < len(parts); i += 2 {
						values = append(values, parts[i])
					}

					list.HandleList(conn, liststore, key, values, 1)

				case "LLEN":

					key := parts[4]

					values := []string{}

					list.HandleList(conn, liststore, key, values, 2)

				}
			}
		}()

	}

}
