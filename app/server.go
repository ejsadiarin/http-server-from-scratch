package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var (
	_ = net.Listen
	_ = os.Exit
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	// extract URL path from request
	r, err := io.ReadAll(conn)
	if err != nil {
		fmt.Println("Error reading from connection", err.Error())
		os.Exit(1)
	}
	fmt.Printf("%v\n", string(r))
	rSlice := strings.Split(string(r), " ")
	urlPath := rSlice[1]
	if urlPath == "/" {
		success := "HTTP/1.1 200 OK\r\n\r\n"
		_, err = conn.Write([]byte(success))
		if err != nil {
			fmt.Println("Error writing data", err.Error())
			os.Exit(1)
		}
	} else {
		fail := "HTTP/1.1 404 Not Found\r\n\r\n"
		_, err = conn.Write([]byte(fail))
		if err != nil {
			fmt.Println("Error writing data", err.Error())
			os.Exit(1)
		}
	}
}
