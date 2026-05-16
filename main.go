package main

import (
	"fmt"
	"net"
	"os"
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

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	// access to tcp connection
	//conn, err := l.Accept()
	// conn is an object

	// Send the response to the client
	//	conn.Write([]byte("+PONG\r\n"))

	buf := make([]byte, 1024) //make()-> does memory allocation similar to new()
	//slice of byte is created to hold incoming data

	for {
		_, err = conn.Read(buf)

		if err != nil {
			break // sender has stopped sending commands
		}

		conn.Write([]byte("+PONG\r\n")) //Redis client libraries (especially in languages like Go) use a slice of bytes ([]byte) instead of standard strings because slices of bytes prevent memory allocations and reduce CPU overhead.
		//also Network sockets read data directly into byte buffers. If the library returned a string, it would have to copy those bytes into a brand new string object in memory. Returning []byte allows the library to pass a direct pointer to the network data without copying it.
	}

}
